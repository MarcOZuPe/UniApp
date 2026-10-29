package reserva

import (
	"context"
	"net/http"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vistaAcceso "uniapp/internal/adapters/presentation/acceso"
	vistaReserva "uniapp/internal/adapters/presentation/reserva"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloReserva "uniapp/internal/application/reserva"
)

// RepositorioAgenda contiene las operaciones necesarias
// para consultar la agenda de escolares y tutores.
type RepositorioAgenda interface {
	PorEscolar(ctx context.Context, escolarID string) ([]modeloReserva.Reserva, error)
	PorTutor(ctx context.Context, tutorID string) ([]modeloReserva.Reserva, error)
}

// ControladorAgenda atiende las solicitudes de agenda.
type ControladorAgenda struct {
	repositorio RepositorioAgenda
}

// RutasAgenda registra las rutas protegidas de la agenda.
func RutasAgenda(
	repositorio RepositorioAgenda,
	autenticar func(http.Handler) http.Handler,
) http.Handler {
	controlador := &ControladorAgenda{
		repositorio: repositorio,
	}

	mux := http.NewServeMux()

	mux.Handle(
		"GET /api/agenda",
		autenticar(http.HandlerFunc(controlador.consultar)),
	)

	return mux
}

// consultar devuelve las reservas del usuario autenticado.
// No recibe IDs de usuario enviados desde el frontend.
func (c *ControladorAgenda) consultar(
	w http.ResponseWriter,
	r *http.Request,
) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())

	if usuario == nil {
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado:  http.StatusUnauthorized,
			Codigo:  "no_autenticado",
			Mensaje: "Debes iniciar sesión.",
		})
		return
	}

	var reservas []modeloReserva.Reserva
	var err error

	switch usuario.Rol {
	case modeloAcceso.RolEscolar:
		reservas, err = c.repositorio.PorEscolar(
			r.Context(),
			usuario.ID,
		)

	case modeloAcceso.RolTutor:
		reservas, err = c.repositorio.PorTutor(
			r.Context(),
			usuario.ID,
		)

	default:
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado:  http.StatusForbidden,
			Codigo:  "rol_no_permitido",
			Mensaje: "Tu rol no tiene acceso a la agenda.",
		})
		return
	}

	if err != nil {
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado:  http.StatusInternalServerError,
			Codigo:  "error_agenda",
			Mensaje: "No se pudo consultar la agenda.",
		})
		return
	}

	vistaAcceso.EscribirJSON(
		w,
		http.StatusOK,
		map[string]any{
			"reservas": vistaReserva.PresentarReservas(reservas),
		},
	)
}
