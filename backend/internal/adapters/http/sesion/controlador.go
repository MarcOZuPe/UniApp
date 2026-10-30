package sesion

import (
	"errors"
	"net/http"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/sesion"
)

type Controlador struct {
	repositorio     modelo.Repositorio
	origenPermitido string
}

func Rutas(
	repositorio modelo.Repositorio,
	autenticar func(http.Handler) http.Handler,
	origenPermitido string,
) http.Handler {
	controlador := &Controlador{
		repositorio:     repositorio,
		origenPermitido: origenPermitido,
	}

	multiplexor := http.NewServeMux()
	multiplexor.HandleFunc("GET /api/sesion/salud", func(response http.ResponseWriter, _ *http.Request) {
		vista.EscribirJSON(response, http.StatusOK, map[string]string{
			"flujo":  "sesion",
			"estado": "implementado",
		})
	})
	multiplexor.Handle("GET /api/sesion/{id}", autenticar(http.HandlerFunc(controlador.consultar)))
	multiplexor.Handle("POST /api/sesion/{id}/asistencias", autenticar(http.HandlerFunc(controlador.declarar)))
	return multiplexor
}

func (controlador *Controlador) consultar(response http.ResponseWriter, request *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(request.Context())
	if usuario == nil {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnauthorized, codigo: "no_autenticado", mensaje: "Debes iniciar sesion.",
		})
		return
	}

	sesion, err := controlador.repositorio.SesionPorID(request.Context(), request.PathValue("id"))
	if err != nil {
		responderErrorRepositorio(response, err)
		return
	}

	if usuario.ID != sesion.TutorID() && usuario.ID != sesion.EscolarID() {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusForbidden, codigo: "no_autorizado", mensaje: "No participas en esta sesion.",
		})
		return
	}

	asistencias := make([]map[string]any, 0, len(sesion.Asistencias()))
	for _, asistencia := range sesion.Asistencias() {
		asistencias = append(asistencias, presentarAsistencia(asistencia))
	}

	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"id":          sesion.ID(),
		"reserva_id":  sesion.ReservaID(),
		"tutor_id":    sesion.TutorID(),
		"escolar_id":  sesion.EscolarID(),
		"inicio":      sesion.Inicio(),
		"fin":         sesion.Fin(),
		"asistencias": asistencias,
	})
}

func (controlador *Controlador) declarar(response http.ResponseWriter, request *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(request.Context())
	if usuario == nil {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnauthorized, codigo: "no_autenticado", mensaje: "Debes iniciar sesion.",
		})
		return
	}

	if request.Body != nil && request.Body != http.NoBody {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusBadRequest, codigo: "solicitud_invalida", mensaje: "No se admite un cuerpo.",
		})
		return
	}

	asistencia, err := controlador.repositorio.DeclararAsistencia(
		request.Context(),
		request.PathValue("id"),
		usuario.ID,
		time.Now().UTC(),
	)
	if err != nil {
		responderErrorRepositorio(response, err)
		return
	}

	vista.EscribirJSON(response, http.StatusCreated, presentarAsistencia(asistencia))
}

func presentarAsistencia(asistencia modelo.Asistencia) map[string]any {
	return map[string]any{
		"usuario_id":   asistencia.UsuarioID(),
		"rol":          string(asistencia.Rol()),
		"declarada_en": asistencia.DeclaradaEn(),
	}
}

type detalleErrorHTTP struct {
	estado  int
	codigo  string
	mensaje string
}

func responderError(response http.ResponseWriter, detalle detalleErrorHTTP) {
	vista.EscribirError(response, vista.RespuestaError{
		Estado:  detalle.estado,
		Codigo:  detalle.codigo,
		Mensaje: detalle.mensaje,
	})
}

func responderErrorRepositorio(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, modelo.ErrSesionNoEncontrada):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusNotFound, codigo: "sesion_no_encontrada", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrNoAutorizado):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusForbidden, codigo: "no_autorizado", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrAsistenciaDuplicada):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusConflict, codigo: "asistencia_duplicada", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrSesionInvalida):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnprocessableEntity, codigo: "sesion_invalida", mensaje: err.Error(),
		})
	default:
		responderError(response, detalleErrorHTTP{
			estado: http.StatusInternalServerError, codigo: "error_interno", mensaje: "Ocurrio un error inesperado.",
		})
	}
}
