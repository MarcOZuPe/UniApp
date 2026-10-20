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
	modeloReserva "uniapp/internal/models/reserva"
	modeloSesion "uniapp/internal/models/sesion"
	"uniapp/migrations"
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

	if configuracion.SembrarSintetico {
		repoSesion := modeloSesion.RepositorioPostgres{DB: baseDatos}

		if err := modeloSesion.SembrarSintetico(
			ctx,
			baseDatos,
			repoSesion,
			configuracion.DominioTutor,
		); err != nil {
			registro.Fatalf("semilla I05: %v", err)
		}

		registro.Println("sesion sintetica I05 lista")
	}

	controlador := ctrl.Nuevo(servicio, configuracion.CookieSegura)

	// Servicio del flujo D: reservas.
	reservas := construirServicioReserva(baseDatos)

	// Extensión del flujo E: sesiones y asistencias.
	manejador := construirManejadorConSesion(
		controlador,
		reservas,
		registro,
		configuracion,
		baseDatos,
	)

	servidor := nuevoServidor(configuracion.Puerto, manejador)
	servirHastaSenalDeParada(registro, servidor, ctx)
}

func construirServicio(
	ctx context.Context,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) *modelo.Servicio {
	hasher := adapters.NuevoHasher()
	fabrica := modelo.NuevaFabrica(hasher, configuracion.DominioTutor)

	servicio := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: modelo.RepositorioPostgres{DB: baseDatos},
		Fabrica:     fabrica,
		Hasher:      hasher,
		Bus:         events.NuevoBus(),
	})

	if configuracion.SembrarSintetico {
		if err := modelo.Sembrar(
			ctx,
			servicio,
			modelo.OpcionesSiembra{
				DominioTutor: configuracion.DominioTutor,
				Clave:        configuracion.ClaveSemilla,
			},
		); err != nil {
			registro.Fatalf("semilla: %v", err)
		}

		registro.Println("cuentas sintéticas de prueba listas")
	}

	return servicio
}

// Servicio del flujo D, conservado desde main.
func construirServicioReserva(baseDatos *sql.DB) *modeloReserva.Servicio {
	return modeloReserva.NuevoServicio(modeloReserva.Dependencias{
		Repositorio: modeloReserva.RepositorioPostgres{DB: baseDatos},
		Oferta:      modeloReserva.NuevaOfertaSintetica(time.Now().UTC()),
	})
}

// Manejador principal: conserva las rutas de acceso y reservas.
func construirManejador(
	controlador *ctrl.Controlador,
	reservas *modeloReserva.Servicio,
	registro *log.Logger,
	configuracion config.Config,
) http.Handler {
	raiz := http.NewServeMux()

	raiz.Handle("/api/tutor/", ctrlTutor.Rutas())
	raiz.Handle("/api/oferta/", ctrlOferta.Rutas())

	// Ambos patrones son necesarios para el flujo de reservas.
	manejadorReserva := ctrlReserva.Nuevo(
		reservas,
		controlador.Autenticar(),
	).Rutas()

	raiz.Handle("/api/reserva", manejadorReserva)
	raiz.Handle("/api/reserva/", manejadorReserva)

	raiz.Handle("/", controlador.Rutas())

	return ctrl.Encadenar(
		raiz,
		ctrl.Recuperar(registro),
		ctrl.Registro(registro),
		ctrl.CabecerasSeguras(),
		ctrl.OrigenPermitido(configuracion.OrigenFrontend),
	)
}

// Extensión I05: agrega sesiones sin sustituir el flujo D.
func construirManejadorConSesion(
	controlador *ctrl.Controlador,
	reservas *modeloReserva.Servicio,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) http.Handler {
	repositorioSesion := modeloSesion.RepositorioPostgres{
		DB: baseDatos,
	}

	rutasSesion := ctrlSesion.Rutas(
		repositorioSesion,
		controlador.Autenticar(),
		configuracion.OrigenFrontend,
	)

	raiz := http.NewServeMux()

	raiz.Handle(
		"/api/sesion/",
		ctrl.Encadenar(
			rutasSesion,
			ctrl.Recuperar(registro),
			ctrl.Registro(registro),
			ctrl.CabecerasSeguras(),
			ctrl.OrigenPermitido(configuracion.OrigenFrontend),
		),
	)

	raiz.Handle(
		"/",
		construirManejador(
			controlador,
			reservas,
			registro,
			configuracion,
		),
	)

	return raiz
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

func servirHastaSenalDeParada(
	registro *log.Logger,
	servidor *http.Server,
	ctx context.Context,
) {
	go func() {
		registro.Printf("API escuchando en %s", servidor.Addr)

		if err := servidor.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			registro.Fatal(err)
		}
	}()

	senalDeParada := make(chan os.Signal, 1)
	signal.Notify(senalDeParada, os.Interrupt, syscall.SIGTERM)

	<-senalDeParada

	contextoDeCierre, cancelar := context.WithTimeout(
		ctx,
		10*time.Second,
	)
	defer cancelar()

	_ = servidor.Shutdown(contextoDeCierre)
}
