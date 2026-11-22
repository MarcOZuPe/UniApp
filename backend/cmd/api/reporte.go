package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	ctrl "uniapp/internal/adapters/http/acceso"
	ctrlSesion "uniapp/internal/adapters/http/sesion"
	accesoRepo "uniapp/internal/adapters/repositories/postgres/acceso"
	sesionRepo "uniapp/internal/adapters/repositories/postgres/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloSesion "uniapp/internal/application/sesion"
	"uniapp/internal/infrastructure/config"
)

// claveSeudonimoDesarrollo solo se usa fuera de produccion. Con ella los
// seudonimos de dev no sirven para reidentificar a nadie en produccion.
const claveSeudonimoDesarrollo = "uniapp-dev-solo-datos-sinteticos-i18"

// nombresDeAcceso adapta el repositorio del flujo A al puerto DirectorioNombres
// de I18. Es el unico punto que conoce ambos flujos (composition root).
type nombresDeAcceso struct {
	repositorio accesoRepo.RepositorioPostgres
}

func (n nombresDeAcceso) NombreDe(ctx context.Context, usuarioID string) (string, error) {
	usuario, err := n.repositorio.UsuarioPorID(ctx, usuarioID)
	if errors.Is(err, modeloAcceso.ErrNoEncontrado) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return usuario.NombreUso, nil
}

// claveSeudonimoI18 lee I18_SEUDONIMO_CLAVE; en produccion es obligatoria.
func claveSeudonimoI18(configuracion config.Config, registro *log.Logger) ([]byte, error) {
	clave := os.Getenv("I18_SEUDONIMO_CLAVE")
	if clave != "" {
		if len(clave) < modeloSesion.MinBytesClaveSeudonimo {
			return nil, fmt.Errorf("I18_SEUDONIMO_CLAVE debe tener al menos %d caracteres", modeloSesion.MinBytesClaveSeudonimo)
		}
		return []byte(clave), nil
	}
	if configuracion.EnProduccion() {
		return nil, errors.New("I18_SEUDONIMO_CLAVE es obligatoria en produccion")
	}
	registro.Printf("I18: I18_SEUDONIMO_CLAVE no definida; se usa la clave de desarrollo (solo datos sinteticos)")
	return []byte(claveSeudonimoDesarrollo), nil
}

// Extension I18: reporte de minutos en CSV y PDF con constancia de descarga.
func construirManejadorConReporte(
	anterior http.Handler,
	autenticar func(http.Handler) http.Handler,
	registro *log.Logger,
	configuracion config.Config,
	baseDatos *sql.DB,
) http.Handler {
	clave, err := claveSeudonimoI18(configuracion, registro)
	if err != nil {
		registro.Fatalf("I18: %v", err)
	}
	seudonimos, err := modeloSesion.NuevoSeudonimizador(clave)
	if err != nil {
		registro.Fatalf("I18: %v", err)
	}
	reservas, err := estadoReservasI17()
	if err != nil {
		registro.Fatalf("I18: %v", err)
	}
	servicio := modeloSesion.ServicioReporte{
		Reportes: sesionRepo.RepositorioReportesPostgres{DB: baseDatos},
		Minutos: modeloSesion.ServicioMinutos{
			Sesiones: sesionRepo.RepositorioPostgres{DB: baseDatos},
			Minutos:  sesionRepo.RepositorioMinutosPostgres{DB: baseDatos},
			Reservas: reservas,
		},
		Nombres:    nombresDeAcceso{repositorio: accesoRepo.RepositorioPostgres{DB: baseDatos}},
		Seudonimos: seudonimos,
	}
	rutas := ctrl.Encadenar(
		ctrlSesion.RutasReporte(servicio, autenticar, nil),
		ctrl.Recuperar(registro),
		ctrl.Registro(registro),
		ctrl.CabecerasSeguras(),
		ctrl.OrigenPermitido(configuracion.OrigenesFrontend...),
	)

	raiz := http.NewServeMux()
	raiz.Handle("/api/reportes/", rutas)
	raiz.Handle("/", anterior)
	return raiz
}
