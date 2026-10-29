package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	ctrl "uniapp/internal/adapters/http/acceso"
	ctrlCatalogo "uniapp/internal/adapters/http/catalogo"
	ctrlOferta "uniapp/internal/adapters/http/oferta"
	ctrlReserva "uniapp/internal/adapters/http/reserva"
	ctrlSesion "uniapp/internal/adapters/http/sesion"
	ctrlTutor "uniapp/internal/adapters/http/tutor"
	accesoRepo "uniapp/internal/adapters/repositories/postgres/acceso"
	catalogoRepo "uniapp/internal/adapters/repositories/postgres/catalogo"
	ofertaRepo "uniapp/internal/adapters/repositories/postgres/oferta"
	reservaRepo "uniapp/internal/adapters/repositories/postgres/reserva"
	sesionRepo "uniapp/internal/adapters/repositories/postgres/sesion"
	tutorRepo "uniapp/internal/adapters/repositories/postgres/tutor"
	modelo "uniapp/internal/application/acceso"
	modeloCatalogo "uniapp/internal/application/catalogo"
	modeloReserva "uniapp/internal/application/reserva"
	modeloTutor "uniapp/internal/application/tutor"
	infra "uniapp/internal/infrastructure"
	"uniapp/internal/infrastructure/config"
	storage "uniapp/internal/infrastructure/storage"
	"uniapp/internal/infrastructure/swaggerui"
	"uniapp/migrations"
)

const (
	limiteDeEnvio     = 10 * time.Second
	intentosDeEnvio   = 3
	esperaEntreEnvios = 2 * time.Second
	intervaloLimpieza = time.Minute
)

func main() {
	registro := log.New(os.Stdout, "", log.LstdFlags|log.LUTC)

	configuracion, err := config.Cargar()
	if err != nil {
		registro.Fatal(err)
	}
	registro.Printf("frontend de enlaces: %s (APP_ENV=%s; orígenes permitidos: %v)", configuracion.OrigenFrontend, configuracion.Entorno, configuracion.OrigenesFrontend)

	ctx, detener := context.WithCancel(context.Background())
	defer detener()

	baseDatos, err := migrations.Abrir(configuracion.DatabaseURL)
	if err != nil {
		registro.Fatalf("base de datos: %v", err)
	}

	if err := migrations.Migrar(ctx, baseDatos, migrations.FS); err != nil {
		registro.Fatalf("migraciones: %v", err)
	}

	if _, err := anticipacionRecordatorio(); err != nil {
		registro.Fatal(err)
	}
	iniciarOutbox(ctx, baseDatos, configuracion, registro)

	servicio := construirServicio(ctx, registro, configuracion, baseDatos)
	go limpiarRegistrosVencidosPeriodicamente(ctx, registro, servicio)

	if configuracion.SembrarSintetico {
		repoSesion := sesionRepo.RepositorioPostgres{DB: baseDatos}

		if err := sesionRepo.SembrarSintetico(
			ctx,
			baseDatos,
			repoSesion,
			primerDominio(configuracion.DominiosTutor),
		); err != nil {
			registro.Fatalf("semilla I05: %v", err)
		}

		registro.Println("sesion sintetica I05 lista")
	}

	controlador := ctrl.Nuevo(servicio, configuracion.CookieSegura)

	servicioTutor := modeloTutor.NuevoServicio(
		tutorRepo.RepositorioPostgres{DB: baseDatos},
		storage.AlmacenDocumentosFS{Directorio: configuracion.DirectorioDocumentos},
	)

	tutores := ctrlTutor.Nuevo(servicioTutor, controlador.Autenticar()).Rutas()

	catalogo := ctrlCatalogo.Nuevo(
		modeloCatalogo.NuevoServicio(catalogoRepo.RepositorioPostgres{DB: baseDatos}),
		controlador.Autenticar(),
	).Rutas()

	// Reservas I09: mantiene la autorización del flujo A.
	// I10 agrega la anticipación de recordatorios.
	reservas := construirServicioReserva(baseDatos, configuracion, servicio)

	// Sesiones I05: conserva el catálogo y las rutas existentes.
	manejadorSesion := construirManejadorConSesion(
		controlador,
		tutores,
		catalogo,
		reservas,
		registro,
		configuracion,
		baseDatos,
	)

	// Agenda y notificaciones I10.
	manejador := construirManejadorConAgenda(
		manejadorSesion,
		controlador.Autenticar(),
		registro,
		configuracion,
		baseDatos,
	)

	servidor := nuevoServidor(configuracion.Puerto, manejador)

	var servidorSwagger *http.Server
	if manejadorSwagger, err := swaggerui.Rutas(rutaContratoOpenAPI()); err != nil {
		registro.Printf("swagger: %v", err)
	} else {
		muxSwagger := http.NewServeMux()
		// La API se expone también en este puerto para que "Try it out"
		// de Swagger UI use el mismo origen y no dependa de CORS.
		muxSwagger.Handle("/", manejador)
		muxSwagger.Handle("/swagger", ctrl.Encadenar(
			manejadorSwagger,
			ctrl.Recuperar(registro),
			ctrl.Registro(registro),
			ctrl.CabecerasSeguras(),
			ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
		))
		muxSwagger.Handle("/swagger/", ctrl.Encadenar(
			manejadorSwagger,
			ctrl.Recuperar(registro),
			ctrl.Registro(registro),
			ctrl.CabecerasSeguras(),
			ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
		))
		servidorSwagger = nuevoServidor(configuracion.PuertoSwagger, muxSwagger)
	}

	servirHastaSenalDeParada(registro, servidor, servidorSwagger, ctx)
}

func construirServicio(
	ctx context.Context,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) *modelo.Servicio {
	mailer := construirMailer(configuracion)
	bus := construirBus(registro, mailer, configuracion)
	hasher := infra.NuevoHasher()
	fabrica := modelo.NuevaFabrica(hasher, configuracion.DominiosTutor)

	servicio := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: accesoRepo.RepositorioPostgres{DB: baseDatos},
		Fabrica:     fabrica,
		Hasher:      hasher,
		Bus:         bus,
	})

	if liberados, err := servicio.LimpiarRegistrosVencidos(ctx); err != nil {
		registro.Printf("limpieza de registros sin confirmar: %v", err)
	} else if liberados > 0 {
		registro.Printf("registros sin confirmar vencidos liberados: %d", liberados)
	}

	if configuracion.SembrarSintetico {
		if err := modelo.Sembrar(
			ctx,
			servicio,
			modelo.OpcionesSiembra{
				DominioTutor: configuracion.DominiosTutor,
				Clave:        configuracion.ClaveSemilla,
			},
		); err != nil {
			registro.Fatalf("semilla: %v", err)
		}

		if err := modelo.SembrarVinculacion(
			ctx,
			servicio,
			modelo.EmailApoderadoSintetico,
			modelo.EmailEscolarSintetico,
		); err != nil {
			registro.Fatalf("semilla de vinculación: %v", err)
		}

		registro.Println("cuentas sintéticas de prueba listas")
	}

	return servicio
}

// Servicio de reservas I09 con autorización del flujo A y outbox I10.
func construirServicioReserva(
	baseDatos *sql.DB,
	configuracion config.Config,
	autorizacion modeloReserva.PuertoAutorizacion,
) *modeloReserva.Servicio {
	anticipacion, _ := anticipacionRecordatorio() // Validada al arrancar.

	var analizador modeloReserva.Analizador = modeloReserva.AnalizadorBasico{}
	if strings.EqualFold(configuracion.AnalisisAdjuntos, "ninguno") {
		// El adjunto queda en cuarentena si no hay análisis disponible.
		analizador = modeloReserva.AnalizadorNoDisponible{}
	}

	return modeloReserva.NuevoServicio(modeloReserva.Dependencias{
		Repositorio: reservaRepo.RepositorioPostgres{
			DB:                       baseDatos,
			AnticipacionRecordatorio: anticipacion,
		},
		Oferta: ofertaConTutorReal(baseDatos, configuracion),
		Almacen: storage.AlmacenAdjuntosFS{
			Directorio: configuracion.DirectorioAdjuntos,
		},
		Analizador:   analizador,
		Limites:      modeloReserva.LimitesPorDefecto,
		Autorizacion: autorizacion,
	})
}

// ofertaConTutorReal asocia la oferta de demostración al tutor sembrado para que la
// reserva conserve su UUID y aparezca en su agenda (I10). El horario sigue siendo
// sintético: la oferta real por curso y semana llega con I08.
func ofertaConTutorReal(baseDatos *sql.DB, configuracion config.Config) modeloReserva.PuertoOferta {
	base := time.Now().UTC()
	if !configuracion.SembrarSintetico {
		return modeloReserva.NuevaOfertaSintetica(base)
	}
	email := "tutor.prueba@" + primerDominio(configuracion.DominiosTutor)
	tutor, err := (accesoRepo.RepositorioPostgres{DB: baseDatos}).UsuarioPorEmail(context.Background(), email)
	if err != nil || tutor.ID == "" {
		return modeloReserva.NuevaOfertaSintetica(base)
	}
	return modeloReserva.NuevaOfertaSinteticaParaTutor(base, tutor.ID, tutor.NombreUso)
}

func construirMailer(configuracion config.Config) modelo.Mailer {
	var mailer modelo.Mailer = infra.LogMailer{
		MostrarCuerpo: !configuracion.CookieSegura,
	}

	if configuracion.SMTP.Host != "" {
		mailer = infra.SMTPMailer{
			Host:      configuracion.SMTP.Host,
			Puerto:    configuracion.SMTP.Puerto,
			Usuario:   configuracion.SMTP.Usuario,
			Clave:     configuracion.SMTP.Clave,
			Remitente: configuracion.SMTP.Remitente,
		}
	}

	return infra.NuevoReintentoMailer(
		infra.NuevoTiempoLimiteMailer(mailer, limiteDeEnvio),
		intentosDeEnvio,
		esperaEntreEnvios,
	)
}

func construirBus(
	registro *log.Logger,
	mailer modelo.Mailer,
	configuracion config.Config,
) *infra.Bus {
	bus := infra.NuevoBus()
	auditoria := infra.NuevaAuditoria(registro)

	for _, nombre := range []string{
		modelo.UsuarioRegistrado,
		modelo.IngresoExitoso,
		modelo.IngresoFallido,
		modelo.ClaveRestablecida,
		modelo.RecuperacionSolicitada,
		modelo.ConfirmacionSolicitada,
		modelo.CorreoConfirmado,
		modelo.VinculacionSolicitada,
		modelo.VinculacionAceptada,
	} {
		bus.Registrar(nombre, auditoria)
	}

	bus.Registrar(
		modelo.RecuperacionSolicitada,
		infra.NuevoCorreoRecuperacion(mailer, configuracion.OrigenFrontend),
	)

	bus.Registrar(
		modelo.ConfirmacionSolicitada,
		infra.NuevoCorreoConfirmacion(mailer, configuracion.OrigenFrontend),
	)

	bus.Registrar(
		modelo.VinculacionSolicitada,
		infra.NuevoCorreoVinculacion(mailer, configuracion.OrigenFrontend),
	)

	return bus
}

// Manejador principal: acceso, tutor, catálogo, oferta y reservas.
func construirManejador(
	controlador *ctrl.Controlador,
	tutores http.Handler,
	catalogo http.Handler,
	reservas *modeloReserva.Servicio,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) http.Handler {
	raiz := http.NewServeMux()

	raiz.Handle("/api/tutor/", tutores)
	raiz.Handle("/api/catalogo", catalogo)
	raiz.Handle("/api/catalogo/", catalogo)
	raiz.Handle(
		"/api/oferta/",
		ctrlOferta.Rutas(ofertaRepo.RepositorioPostgres{DB: baseDatos}, controlador.Autenticar()),
	)

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
		ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
	)
}

// Extensión I05: sesiones y asistencias sin sustituir el flujo D.
func construirManejadorConSesion(
	controlador *ctrl.Controlador,
	tutores http.Handler,
	catalogo http.Handler,
	reservas *modeloReserva.Servicio,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) http.Handler {
	repositorioSesion := sesionRepo.RepositorioPostgres{
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
			ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
		),
	)

	raiz.Handle(
		"/",
		construirManejador(
			controlador,
			tutores,
			catalogo,
			reservas,
			registro,
			configuracion,
			baseDatos,
		),
	)

	return raiz
}

// Extensión I10: agenda y notificaciones.
func construirManejadorConAgenda(
	anterior http.Handler,
	autenticar func(http.Handler) http.Handler,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) http.Handler {
	raiz := http.NewServeMux()

	rutasAgenda := ctrlReserva.RutasAgenda(
		reservaRepo.RepositorioPostgres{DB: baseDatos},
		autenticar,
	)

	raiz.Handle(
		"/api/agenda",
		ctrl.Encadenar(
			rutasAgenda,
			ctrl.Recuperar(registro),
			ctrl.Registro(registro),
			ctrl.CabecerasSeguras(),
			ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
		),
	)

	raiz.Handle(
		"/api/notificaciones",
		ctrl.Encadenar(
			ctrlSesion.RutasNotificaciones(
				sesionRepo.RepositorioOutboxPostgres{DB: baseDatos},
				autenticar,
			),
			ctrl.Recuperar(registro),
			ctrl.Registro(registro),
			ctrl.CabecerasSeguras(),
			ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
		),
	)

	raiz.Handle("/", anterior)
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

func limpiarRegistrosVencidosPeriodicamente(
	ctx context.Context,
	registro *log.Logger,
	servicio *modelo.Servicio,
) {
	ticker := time.NewTicker(intervaloLimpieza)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if liberados, err := servicio.LimpiarRegistrosVencidos(ctx); err != nil {
				registro.Printf("limpieza de registros sin confirmar: %v", err)
			} else if liberados > 0 {
				registro.Printf("registros sin confirmar vencidos liberados: %d", liberados)
			}
		case <-ctx.Done():
			return
		}
	}
}

func rutaContratoOpenAPI() string {
	if valor := strings.TrimSpace(os.Getenv("OPENAPI_PATH")); valor != "" {
		return valor
	}
	directorio, err := os.Getwd()
	if err != nil {
		return "docs/openapi.yaml"
	}
	for i := 0; i < 10; i++ {
		candidato := filepath.Join(directorio, "docs", "openapi.yaml")
		if info, err := os.Stat(candidato); err == nil && !info.IsDir() {
			return candidato
		}
		padre := filepath.Dir(directorio)
		if padre == directorio {
			break
		}
		directorio = padre
	}
	return "docs/openapi.yaml"
}

func servirHastaSenalDeParada(
	registro *log.Logger,
	servidor *http.Server,
	servidorSwagger *http.Server,
	ctx context.Context,
) {
	go func() {
		registro.Printf("API escuchando en %s", servidor.Addr)
		if err := servidor.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			registro.Fatal(err)
		}
	}()

	if servidorSwagger != nil {
		go func() {
			registro.Printf("Swagger UI escuchando en %s", servidorSwagger.Addr)
			if err := servidorSwagger.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				registro.Fatal(err)
			}
		}()
	}

	senalDeParada := make(chan os.Signal, 1)
	signal.Notify(senalDeParada, os.Interrupt, syscall.SIGTERM)
	<-senalDeParada

	contextoDeCierre, cancelar := context.WithTimeout(ctx, 10*time.Second)
	defer cancelar()
	_ = servidor.Shutdown(contextoDeCierre)
	if servidorSwagger != nil {
		_ = servidorSwagger.Shutdown(contextoDeCierre)
	}
}

// Primer dominio no vacío de la lista.
func primerDominio(texto string) string {
	for _, parte := range strings.Split(texto, ",") {
		if dominio := strings.TrimSpace(parte); dominio != "" {
			return dominio
		}
	}
	return ""
}
