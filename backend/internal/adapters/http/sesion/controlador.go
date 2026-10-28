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
	c := &Controlador{
		repositorio:     repositorio,
		origenPermitido: origenPermitido,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/sesion/salud", func(w http.ResponseWriter, _ *http.Request) {
		vista.EscribirJSON(w, http.StatusOK, map[string]string{
			"flujo":  "sesion",
			"estado": "implementado",
		})
	})

	mux.Handle(
		"GET /api/sesion/{id}",
		autenticar(http.HandlerFunc(c.consultar)),
	)

	mux.Handle(
		"POST /api/sesion/{id}/asistencias",
		autenticar(http.HandlerFunc(c.declarar)),
	)

	return mux
}

func (c *Controlador) consultar(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		responderError(w, http.StatusUnauthorized, "no_autenticado", "Debes iniciar sesion.")
		return
	}

	s, err := c.repositorio.SesionPorID(r.Context(), r.PathValue("id"))
	if err != nil {
		responderErrorRepositorio(w, err)
		return
	}

	if usuario.ID != s.TutorID() && usuario.ID != s.EscolarID() {
		responderError(w, http.StatusForbidden, "no_autorizado", "No participas en esta sesion.")
		return
	}

	asistencias := make([]map[string]any, 0, len(s.Asistencias()))
	for _, a := range s.Asistencias() {
		asistencias = append(asistencias, presentarAsistencia(a))
	}

	vista.EscribirJSON(w, http.StatusOK, map[string]any{
		"id":          s.ID(),
		"reserva_id":  s.ReservaID(),
		"tutor_id":    s.TutorID(),
		"escolar_id":  s.EscolarID(),
		"inicio":      s.Inicio(),
		"fin":         s.Fin(),
		"asistencias": asistencias,
	})
}

func (c *Controlador) declarar(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		responderError(w, http.StatusUnauthorized, "no_autenticado", "Debes iniciar sesion.")
		return
	}

	// Este endpoint no recibe un identificador de usuario del cliente.
	if r.Body != nil && r.Body != http.NoBody {
		responderError(w, http.StatusBadRequest, "solicitud_invalida", "No se admite un cuerpo.")
		return
	}

	a, err := c.repositorio.DeclararAsistencia(
		r.Context(),
		r.PathValue("id"),
		usuario.ID,
		time.Now().UTC(),
	)
	if err != nil {
		responderErrorRepositorio(w, err)
		return
	}

	vista.EscribirJSON(w, http.StatusCreated, presentarAsistencia(a))
}

func presentarAsistencia(a modelo.Asistencia) map[string]any {
	return map[string]any{
		"usuario_id":   a.UsuarioID(),
		"rol":          string(a.Rol()),
		"declarada_en": a.DeclaradaEn(),
	}
}

func responderError(
	w http.ResponseWriter,
	estado int,
	codigo string,
	mensaje string,
) {
	vista.EscribirError(w, vista.RespuestaError{
		Estado:  estado,
		Codigo:  codigo,
		Mensaje: mensaje,
	})
}

func responderErrorRepositorio(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, modelo.ErrSesionNoEncontrada):
		responderError(w, http.StatusNotFound, "sesion_no_encontrada", err.Error())
	case errors.Is(err, modelo.ErrNoAutorizado):
		responderError(w, http.StatusForbidden, "no_autorizado", err.Error())
	case errors.Is(err, modelo.ErrAsistenciaDuplicada):
		responderError(w, http.StatusConflict, "asistencia_duplicada", err.Error())
	case errors.Is(err, modelo.ErrSesionInvalida):
		responderError(w, http.StatusUnprocessableEntity, "sesion_invalida", err.Error())
	default:
		responderError(w, http.StatusInternalServerError, "error_interno", "Ocurrio un error inesperado.")
	}
}
