package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	ctrl "uniapp/internal/adapters/http/acceso"
	ctrlSesion "uniapp/internal/adapters/http/sesion"
	sesionMemoria "uniapp/internal/adapters/repositories/memory/sesion"
	sesionRepo "uniapp/internal/adapters/repositories/postgres/sesion"
	modeloSesion "uniapp/internal/application/sesion"
	"uniapp/internal/infrastructure/config"
)

// estadoReservasI17 elige el adaptador del puerto EstadoReservas.
// Hoy: fixture sintetico opcional (I17_CANCELACIONES_FIXTURE). Cuando el
// flujo D publique la cancelacion (I16), solo cambia esta funcion.
func estadoReservasI17() (modeloSesion.EstadoReservas, error) {
	ruta := os.Getenv("I17_CANCELACIONES_FIXTURE")
	if ruta == "" {
		return sesionMemoria.NuevasCancelacionesFijas(), nil
	}
	cancelaciones, err := sesionMemoria.CargarCancelacionesFixture(ruta)
	if err != nil {
		return nil, fmt.Errorf("I17_CANCELACIONES_FIXTURE: %w", err)
	}
	return cancelaciones, nil
}

// Extension I17: minutos verificados y ajustes auditados.
func construirManejadorConMinutos(
	anterior http.Handler,
	autenticar func(http.Handler) http.Handler,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) http.Handler {
	reservas, err := estadoReservasI17()
	if err != nil {
		registro.Fatalf("I17: %v", err)
	}
	servicio := modeloSesion.ServicioMinutos{
		Sesiones: sesionRepo.RepositorioPostgres{DB: baseDatos},
		Minutos:  sesionRepo.RepositorioMinutosPostgres{DB: baseDatos},
		Reservas: reservas,
	}
	rutas := ctrl.Encadenar(
		ctrlSesion.RutasMinutos(servicio, autenticar, nil),
		ctrl.Recuperar(registro),
		ctrl.Registro(registro),
		ctrl.CabecerasSeguras(),
		ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
	)

	raiz := http.NewServeMux()
	// Patrones mas especificos que "/api/sesion/": no alteran las rutas de I05/I11.
	raiz.Handle("/api/sesion/{id}/minutos", rutas)
	raiz.Handle("/api/sesion/{id}/minutos/disputa", rutas)
	raiz.Handle("/api/minutos/", rutas)
	raiz.Handle("/", anterior)
	return raiz
}
