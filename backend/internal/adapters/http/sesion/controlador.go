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
	reloj           func() time.Time
}

// Opcion configura el controlador sin ampliar la firma de Rutas
// (patron functional options).
type Opcion func(*Controlador)

// ConReloj inyecta la fuente de la hora actual. Produccion usa time.Now;
// las pruebas fijan un momento determinista dentro o fuera de los plazos.
func ConReloj(reloj func() time.Time) Opcion {
	return func(controlador *Controlador) {
		controlador.reloj = reloj
	}
}

func Rutas(
	repositorio modelo.Repositorio,
	autenticar func(http.Handler) http.Handler,
	origenPermitido string,
	opciones ...Opcion,
) http.Handler {
	controlador := &Controlador{
		repositorio:     repositorio,
		origenPermitido: origenPermitido,
		reloj:           time.Now,
	}
	for _, opcion := range opciones {
		opcion(controlador)
	}

	multiplexor := http.NewServeMux()
	multiplexor.HandleFunc("GET /api/sesion/salud", func(response http.ResponseWriter, _ *http.Request) {
		vista.EscribirJSON(response, http.StatusOK, map[string]string{
			"flujo":  "sesion",
			"estado": "implementado",
		})
	})
	multiplexor.Handle("GET /api/sesion/{id}", autenticar(http.HandlerFunc(controlador.consultar)))
	// I05: consultar la sesion asociada a una reserva (navegacion desde la agenda).
	multiplexor.Handle("GET /api/sesion/por-reserva/{reservaID}", autenticar(http.HandlerFunc(controlador.consultarPorReserva)))
	multiplexor.Handle("POST /api/sesion/{id}/asistencias", autenticar(http.HandlerFunc(controlador.declarar)))
	// I11: ausencia propia y resultado de la sesion.
	multiplexor.Handle("POST /api/sesion/{id}/ausencias", autenticar(http.HandlerFunc(controlador.reportarAusencia)))
	multiplexor.Handle("POST /api/sesion/{id}/resultado", autenticar(http.HandlerFunc(controlador.registrarResultado)))
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

	ausencias := make([]map[string]any, 0, len(sesion.Ausencias()))
	for _, ausencia := range sesion.Ausencias() {
		ausencias = append(ausencias, presentarAusencia(ausencia))
	}

	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"id":          sesion.ID(),
		"reserva_id":  sesion.ReservaID(),
		"tutor_id":    sesion.TutorID(),
		"escolar_id":  sesion.EscolarID(),
		"inicio":      sesion.Inicio(),
		"fin":         sesion.Fin(),
		"ausencias":   ausencias,
		"resultado":   presentarResultado(sesion.Resultado()),
		"plazos":      presentarPlazos(sesion),
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
		controlador.ahora(),
	)
	if err != nil {
		responderErrorRepositorio(response, err)
		return
	}

	vista.EscribirJSON(response, http.StatusCreated, presentarAsistencia(asistencia))
}

func (controlador *Controlador) consultarPorReserva(response http.ResponseWriter, request *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(request.Context())
	if usuario == nil {
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnauthorized, codigo: "no_autenticado", mensaje: "Debes iniciar sesion.",
		})
		return
	}

	sesion, err := controlador.repositorio.SesionPorReservaID(request.Context(), request.PathValue("reservaID"))
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

	vista.EscribirJSON(response, http.StatusOK, map[string]any{
		"id":         sesion.ID(),
		"reserva_id": sesion.ReservaID(),
	})
}

func (controlador *Controlador) ahora() time.Time {
	return controlador.reloj().UTC()
}

func presentarPlazos(sesion *modelo.Sesion) map[string]any {
	return map[string]any{
		"declaracion": presentarIntervalo(sesion.VentanaDeclaracion()),
		"resultado":   presentarIntervalo(sesion.VentanaResultado()),
	}
}

func presentarIntervalo(intervalo modelo.Intervalo) map[string]any {
	return map[string]any{
		"desde": intervalo.Desde(),
		"hasta": intervalo.Hasta(),
	}
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
	case errors.Is(err, modelo.ErrResultadoDuplicado):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusConflict, codigo: "resultado_duplicado", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrResultadoInvalido):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnprocessableEntity, codigo: "resultado_invalido", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrAusenciaDuplicada):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusConflict, codigo: "ausencia_duplicada", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrDeclaracionContradictoria):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusConflict, codigo: "declaracion_contradictoria", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrAsistenciaDuplicada):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusConflict, codigo: "asistencia_duplicada", mensaje: err.Error(),
		})
	case errors.Is(err, modelo.ErrFueraDePlazo):
		responderError(response, detalleErrorHTTP{
			estado: http.StatusUnprocessableEntity, codigo: "fuera_de_plazo", mensaje: err.Error(),
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
