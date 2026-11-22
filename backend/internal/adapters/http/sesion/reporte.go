package sesion

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	comun "uniapp/internal/adapters/presentation/acceso"
	vista "uniapp/internal/adapters/presentation/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/sesion"
)

// formatoFechaReporte es el formato de desde/hasta en la consulta (dias de Lima).
const formatoFechaReporte = "2006-01-02"

// ServicioReporte son los casos de uso de I18 que usa el controlador.
type ServicioReporte interface {
	Generar(ctx context.Context, filtro modelo.FiltroReporte, coordinadorID string, ahora time.Time) (*modelo.Reporte, error)
	RegistrarDescarga(ctx context.Context, reporte *modelo.Reporte, formato modelo.FormatoReporte, contenido []byte) (modelo.RegistroReporte, error)
	Recientes(ctx context.Context) ([]modelo.RegistroReporte, error)
}

// ControladorReporte atiende I18. Solo coordinacion; la hora y el autor los
// fija el servidor.
type ControladorReporte struct {
	servicio ServicioReporte
	reloj    func() time.Time
}

// RutasReporte registra:
//   - GET /api/reportes/minutos?desde=AAAA-MM-DD&hasta=AAAA-MM-DD&formato=csv|pdf[&tutor_id=]
//   - GET /api/reportes/registro (constancias de descarga)
func RutasReporte(servicio ServicioReporte, autenticar func(http.Handler) http.Handler, reloj func() time.Time) http.Handler {
	if reloj == nil {
		reloj = time.Now
	}
	c := &ControladorReporte{servicio: servicio, reloj: reloj}
	coordinacion := func(h http.HandlerFunc) http.Handler {
		return autenticar(ctrlAcceso.RequierePermiso(modeloAcceso.PermisoCoordinacionAcceder)(h))
	}
	mux := http.NewServeMux()
	mux.Handle("GET /api/reportes/minutos", coordinacion(c.descargar))
	mux.Handle("GET /api/reportes/registro", coordinacion(c.registro))
	return mux
}

var parametrosReporte = map[string]bool{"desde": true, "hasta": true, "formato": true, "tutor_id": true}

func (c *ControladorReporte) descargar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	consulta := r.URL.Query()
	for clave, valores := range consulta {
		if !parametrosReporte[clave] || len(valores) != 1 {
			invalidaReporte(w, "solicitud_invalida", "Parametros admitidos: desde, hasta, formato y tutor_id (una vez cada uno).")
			return
		}
	}
	formato, err := modelo.ValidarFormatoReporte(consulta.Get("formato"))
	if err != nil {
		invalidaReporte(w, "formato_invalido", "El formato debe ser csv o pdf.")
		return
	}
	desde, errDesde := time.ParseInLocation(formatoFechaReporte, consulta.Get("desde"), vista.ZonaLima)
	hasta, errHasta := time.ParseInLocation(formatoFechaReporte, consulta.Get("hasta"), vista.ZonaLima)
	if errDesde != nil || errHasta != nil {
		invalidaReporte(w, "fecha_invalida", "Indica desde y hasta con el formato AAAA-MM-DD.")
		return
	}
	// "hasta" es un dia inclusivo: el filtro llega hasta el inicio del dia siguiente.
	filtro, err := modelo.NuevoFiltroReporte(desde, hasta.AddDate(0, 0, 1), consulta.Get("tutor_id"))
	if err != nil {
		responderErrorReporte(w, err)
		return
	}

	reporte, err := c.servicio.Generar(r.Context(), filtro, usuario.ID, c.reloj().UTC())
	if err != nil {
		responderErrorReporte(w, err)
		return
	}
	contenido, err := vista.RenderizarReporte(reporte, formato)
	if err != nil {
		responderErrorReporte(w, err)
		return
	}
	// Sin constancia no se entrega: la huella debe quedar antes de la descarga.
	constancia, err := c.servicio.RegistrarDescarga(r.Context(), reporte, formato, contenido)
	if err != nil {
		responderErrorReporte(w, err)
		return
	}

	cabeceras := w.Header()
	cabeceras.Set("Content-Type", vista.TipoContenidoReporte(formato))
	cabeceras.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, vista.NombreArchivoReporte(reporte, formato)))
	cabeceras.Set("Cache-Control", "no-store")
	cabeceras.Set("X-Content-Type-Options", "nosniff")
	cabeceras.Set("X-Reporte-Id", constancia.ID)
	cabeceras.Set("X-Reporte-Sha256", constancia.SHA256)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(contenido)
}

func (c *ControladorReporte) registro(w http.ResponseWriter, r *http.Request) {
	registros, err := c.servicio.Recientes(r.Context())
	if err != nil {
		responderErrorReporte(w, err)
		return
	}
	comun.EscribirJSON(w, http.StatusOK, map[string]any{"reportes": vista.PresentarRegistrosReporte(registros)})
}

func invalidaReporte(w http.ResponseWriter, codigo, mensaje string) {
	responderError(w, detalleErrorHTTP{estado: http.StatusBadRequest, codigo: codigo, mensaje: mensaje})
}

func responderErrorReporte(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, modelo.ErrFiltroReporteInvalido):
		responderError(w, detalleErrorHTTP{estado: http.StatusUnprocessableEntity, codigo: "periodo_invalido",
			mensaje: fmt.Sprintf("El periodo debe tener desde anterior o igual a hasta y abarcar como maximo %d dias.", modelo.MaxDiasReporte)})
	case errors.Is(err, modelo.ErrFormatoReporteInvalido):
		invalidaReporte(w, "formato_invalido", "El formato debe ser csv o pdf.")
	default:
		responderError(w, detalleErrorHTTP{estado: http.StatusInternalServerError, codigo: "error_interno", mensaje: "No se pudo generar el reporte."})
	}
}
