package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"uniapp/internal/adapters"
	"uniapp/internal/config"
	ctrl "uniapp/internal/controllers/acceso"
	ctrlOferta "uniapp/internal/controllers/oferta"
	ctrlReserva "uniapp/internal/controllers/reserva"
	ctrlSesion "uniapp/internal/controllers/sesion"
	ctrlTutor "uniapp/internal/controllers/tutor"
	"uniapp/internal/db"
	"uniapp/internal/events"
	modelo "uniapp/internal/models/acceso"
	"uniapp/internal/observers"
	"uniapp/migrations"
)

const (
	limiteDeEnvio     = 10 * time.Second
	intentosDeEnvio   = 3
	esperaEntreEnvios = 2 * time.Second
)

func main() {
	registro := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)
	configuracion, err := config.Cargar()
	if err != nil {
		registro.Fatal(err)
	}
	ctx := context.Background()
	baseDatos, err := db.Abrir(configuracion.DatabaseURL)
	if err != nil {
		registro.Fatalf("base de datos: %v", err)
	}
	if err := db.Migrar(ctx, baseDatos, migrations.FS); err != nil {
		registro.Fatalf("migraciones: %v", err)
	}
	servicio := construirServicio(ctx, registro, configuracion, baseDatos)
	controlador := ctrl.Nuevo(servicio, configuracion.CookieSegura)
	servidor := nuevoServidor(configuracion.Puerto, construirManejador(controlador, registro, configuracion))
	servirHastaSenalDeParada(registro, servidor, ctx)
}

func construirServicio(ctx context.Context, registro *log.Logger, configuracion config.Config, baseDatos *sql.DB) *modelo.Servicio {
	mailer := construirMailer(configuracion)
	bus := construirBus(registro, mailer, configuracion)
	hasher := adapters.NuevoHasher()
	fabrica := modelo.NuevaFabrica(hasher, configuracion.DominiosTutor)
	servicio := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: modelo.RepositorioPostgres{DB: baseDatos},
		Fabrica:     fabrica,
		Hasher:      hasher,
		Bus:         bus,
	})
	if configuracion.SembrarSintetico {
		if err := modelo.Sembrar(ctx, servicio, modelo.OpcionesSiembra{DominioTutor: configuracion.DominiosTutor, Clave: configuracion.ClaveSemilla}); err != nil {
			registro.Fatalf("semilla: %v", err)
		}
		registro.Println("cuentas sintéticas de prueba listas")
	}
	return servicio
}

func construirMailer(configuracion config.Config) adapters.Mailer {
	var mailer adapters.Mailer = adapters.LogMailer{MostrarCuerpo: !configuracion.CookieSegura}
	if configuracion.SMTP.Host != "" {
		mailer = adapters.SMTPMailer{
			Host: configuracion.SMTP.Host, Puerto: configuracion.SMTP.Puerto,
			Usuario: configuracion.SMTP.Usuario, Clave: configuracion.SMTP.Clave,
			Remitente: configuracion.SMTP.Remitente,
		}
	}
	return adapters.NuevoReintentoMailer(adapters.NuevoTiempoLimiteMailer(mailer, limiteDeEnvio), intentosDeEnvio, esperaEntreEnvios)
}

func construirBus(registro *log.Logger, mailer adapters.Mailer, configuracion config.Config) *events.Bus {
	bus := events.NuevoBus()
	auditoria := observers.NuevaAuditoria(registro)
	for _, nombre := range []string{events.UsuarioRegistrado, events.IngresoExitoso, events.IngresoFallido, events.ClaveRestablecida, events.RecuperacionSolicitada} {
		bus.Registrar(nombre, auditoria)
	}
	bus.Registrar(events.RecuperacionSolicitada, observers.NuevoCorreoRecuperacion(mailer, configuracion.OrigenFrontend))
	return bus
}

func construirManejador(controlador *ctrl.Controlador, registro *log.Logger, configuracion config.Config) http.Handler {
	raiz := http.NewServeMux()
	raiz.Handle("/api/tutor/", ctrlTutor.Rutas())
	raiz.Handle("/api/oferta/", ctrlOferta.Rutas())
	raiz.Handle("/api/reserva/", ctrlReserva.Rutas())
	raiz.Handle("/api/sesion/", ctrlSesion.Rutas())
	raiz.Handle("/", controlador.Rutas())
	return ctrl.Encadenar(raiz,
		ctrl.Recuperar(registro), ctrl.Registro(registro), ctrl.CabecerasSeguras(),
		ctrl.OrigenPermitido(configuracion.OrigenFrontend))
}

func nuevoServidor(puerto string, manejador http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + puerto,
		Handler:           manejador,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
}

func servirHastaSenalDeParada(registro *log.Logger, servidor *http.Server, ctx context.Context) {
	go func() {
		registro.Printf("API escuchando en %s", servidor.Addr)
		if err := servidor.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			registro.Fatal(err)
		}
	}()
	senalDeParada := make(chan os.Signal, 1)
	signal.Notify(senalDeParada, os.Interrupt, syscall.SIGTERM)
	<-senalDeParada
	contextoDeCierre, cancelar := context.WithTimeout(ctx, 10*time.Second)
	defer cancelar()
	_ = servidor.Shutdown(contextoDeCierre)
}
