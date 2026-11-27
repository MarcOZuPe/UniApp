package sesion

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	comun "uniapp/internal/adapters/presentation/acceso"
	vista "uniapp/internal/adapters/presentation/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/sesion"
)

const maxBytesMinutos = 4 * 1024

// ServicioMinutos son los casos de uso de I17 que usa el controlador.
type ServicioMinutos interface {
	Proponer(ctx context.Context, sesionID, tutorID string, minutos int, ahora time.Time) (*modelo.RegistroMinutos, error)
	Consultar(ctx context.Context, sesionID string, ahora time.Time) (*modelo.RegistroMinutos, error)
	Aprobar(ctx context.Context, sesionID, coordinadorID string, ahora time.Time) (*modelo.RegistroMinutos, error)
	Ajustar(ctx context.Context, sesionID, coordinadorID string, minutos int, motivo string, ahora time.Time) (*modelo.RegistroMinutos, error)
	Disputar(ctx context.Context, sesionID, usuarioID, motivo string, ahora time.Time) (*modelo.RegistroMinutos, error)
	Resolver(ctx context.Context, sesionID, coordinadorID string, minutos int, motivo string, ahora time.Time) (*modelo.RegistroMinutos, error)
	Pendientes(ctx context.Context, ahora time.Time) ([]*modelo.RegistroMinutos, error)
	Total(ctx context.Context, tutorID string, ahora time.Time) (modelo.TotalMinutos, error)
}

// ControladorMinutos atiende I17. La identidad y la hora siempre las fija el
// servidor; el cliente solo envia minutos y motivos.
type ControladorMinutos struct {
	servicio ServicioMinutos
	reloj    func() time.Time
}

// RutasMinutos registra las rutas de I17:
//   - participantes: /api/sesion/{id}/minutos[...]
//   - coordinacion (permiso coordinacion:acceder): /api/minutos/...
func RutasMinutos(
	servicio ServicioMinutos,
	autenticar func(http.Handler) http.Handler,
	reloj func() time.Time,
) http.Handler {
	if reloj == nil {
		reloj = time.Now
	}
	c := &ControladorMinutos{servicio: servicio, reloj: reloj}
	coordinacion := func(h http.HandlerFunc) http.Handler {
		return autenticar(ctrlAcceso.RequierePermiso(modeloAcceso.PermisoCoordinacionAcceder)(h))
	}

	mux := http.NewServeMux()
	mux.Handle("GET /api/sesion/{id}/minutos", autenticar(http.HandlerFunc(c.consultar)))
	mux.Handle("POST /api/sesion/{id}/minutos", autenticar(http.HandlerFunc(c.proponer)))
	mux.Handle("POST /api/sesion/{id}/minutos/disputa", autenticar(http.HandlerFunc(c.disputar)))
	mux.Handle("GET /api/minutos/total", autenticar(http.HandlerFunc(c.total)))
	mux.Handle("GET /api/minutos/pendientes", coordinacion(c.pendientes))
	mux.Handle("POST /api/minutos/{id}/aprobacion", coordinacion(c.aprobar))
	mux.Handle("POST /api/minutos/{id}/ajustes", coordinacion(c.ajustar))
	mux.Handle("POST /api/minutos/{id}/resolucion", coordinacion(c.resolver))
	mux.Handle("GET /api/minutos/{id}/auditoria", coordinacion(c.auditoria))
	return mux
}

type solicitudMinutos struct {
	Minutos *int   `json:"minutos"`
	Motivo  string `json:"motivo"`
}

func (c *ControladorMinutos) ahora() time.Time { return c.reloj().UTC() }

func (c *ControladorMinutos) proponer(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	entrada, ok := decodificarMinutos(w, r, true, false)
	if !ok {
		return
	}
	registro, err := c.servicio.Proponer(r.Context(), r.PathValue("id"), usuario.ID, *entrada.Minutos, c.ahora())
	c.responder(w, http.StatusCreated, registro, err)
}

// consultar: lo ven los participantes de la sesion y la coordinacion.
func (c *ControladorMinutos) consultar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	registro, err := c.servicio.Consultar(r.Context(), r.PathValue("id"), c.ahora())
	if err == nil && !registro.Participa(usuario.ID) && !esCoordinacion(usuario) {
		err = modelo.ErrNoAutorizado
	}
	c.responder(w, http.StatusOK, registro, err)
}

func (c *ControladorMinutos) disputar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	entrada, ok := decodificarMinutos(w, r, false, true)
	if !ok {
		return
	}
	registro, err := c.servicio.Disputar(r.Context(), r.PathValue("id"), usuario.ID, entrada.Motivo, c.ahora())
	c.responder(w, http.StatusOK, registro, err)
}

func (c *ControladorMinutos) aprobar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	if r.Body != nil && r.Body != http.NoBody {
		responderError(w, detalleErrorHTTP{estado: http.StatusBadRequest, codigo: "solicitud_invalida", mensaje: "No se admite un cuerpo."})
		return
	}
	registro, err := c.servicio.Aprobar(r.Context(), r.PathValue("id"), usuario.ID, c.ahora())
	c.responder(w, http.StatusOK, registro, err)
}

func (c *ControladorMinutos) ajustar(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	entrada, ok := decodificarMinutos(w, r, true, true)
	if !ok {
		return
	}
	registro, err := c.servicio.Ajustar(r.Context(), r.PathValue("id"), usuario.ID, *entrada.Minutos, entrada.Motivo, c.ahora())
	c.responder(w, http.StatusOK, registro, err)
}

func (c *ControladorMinutos) resolver(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	entrada, ok := decodificarMinutos(w, r, true, true)
	if !ok {
		return
	}
	registro, err := c.servicio.Resolver(r.Context(), r.PathValue("id"), usuario.ID, *entrada.Minutos, entrada.Motivo, c.ahora())
	c.responder(w, http.StatusOK, registro, err)
}

func (c *ControladorMinutos) pendientes(w http.ResponseWriter, r *http.Request) {
	registros, err := c.servicio.Pendientes(r.Context(), c.ahora())
	if err != nil {
		responderErrorMinutos(w, err)
		return
	}
	comun.EscribirJSON(w, http.StatusOK, map[string]any{"pendientes": vista.PresentarListaMinutos(registros)})
}

func (c *ControladorMinutos) auditoria(w http.ResponseWriter, r *http.Request) {
	registro, err := c.servicio.Consultar(r.Context(), r.PathValue("id"), c.ahora())
	if err != nil {
		responderErrorMinutos(w, err)
		return
	}
	comun.EscribirJSON(w, http.StatusOK, map[string]any{
		"minutos":   vista.PresentarMinutos(registro),
		"auditoria": vista.PresentarAuditoria(registro.Auditoria()),
	})
}

// total: el tutor ve solo su total; la coordinacion puede pedir el de
// cualquier tutor con ?tutor_id=.
func (c *ControladorMinutos) total(w http.ResponseWriter, r *http.Request) {
	usuario, ok := exigirUsuario(w, r)
	if !ok {
		return
	}
	tutorID := r.URL.Query().Get("tutor_id")
	switch {
	case esCoordinacion(usuario):
		if tutorID == "" {
			responderError(w, detalleErrorHTTP{estado: http.StatusBadRequest, codigo: "solicitud_invalida", mensaje: "Indica tutor_id."})
			return
		}
	case usuario.Rol == modeloAcceso.RolTutor:
		if tutorID != "" && tutorID != usuario.ID {
			responderError(w, detalleErrorHTTP{estado: http.StatusForbidden, codigo: "no_autorizado", mensaje: "Solo puedes ver tu propio total."})
			return
		}
		tutorID = usuario.ID
	default:
		responderError(w, detalleErrorHTTP{estado: http.StatusForbidden, codigo: "rol_no_permitido", mensaje: "Tu rol no tiene acceso a los totales."})
		return
	}
	total, err := c.servicio.Total(r.Context(), tutorID, c.ahora())
	if err != nil {
		responderErrorMinutos(w, err)
		return
	}
	comun.EscribirJSON(w, http.StatusOK, map[string]any{
		"tutor_id":            total.TutorID,
		"minutos_verificados": total.Verificados,
		"registros":           vista.PresentarListaMinutos(total.Registros),
	})
}

func (c *ControladorMinutos) responder(w http.ResponseWriter, estado int, registro *modelo.RegistroMinutos, err error) {
	if err != nil {
		responderErrorMinutos(w, err)
		return
	}
	comun.EscribirJSON(w, estado, vista.PresentarMinutos(registro))
}

func exigirUsuario(w http.ResponseWriter, r *http.Request) (*modeloAcceso.Usuario, bool) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		responderError(w, detalleErrorHTTP{estado: http.StatusUnauthorized, codigo: "no_autenticado", mensaje: "Debes iniciar sesion."})
		return nil, false
	}
	return usuario, true
}

func esCoordinacion(usuario *modeloAcceso.Usuario) bool {
	return usuario.Rol.TienePermiso(modeloAcceso.PermisoCoordinacionAcceder)
}

// decodificarMinutos acepta un unico objeto JSON acotado, sin campos ajenos
// (autor, fecha, estado...). conMinutos/conMotivo marcan lo obligatorio.
func decodificarMinutos(w http.ResponseWriter, r *http.Request, conMinutos, conMotivo bool) (solicitudMinutos, bool) {
	var entrada solicitudMinutos
	invalida := func(mensaje string) (solicitudMinutos, bool) {
		responderError(w, detalleErrorHTTP{estado: http.StatusBadRequest, codigo: "solicitud_invalida", mensaje: mensaje})
		return entrada, false
	}
	decodificador := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBytesMinutos))
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(&entrada); err != nil {
		return invalida("Envia un objeto JSON valido.")
	}
	if err := decodificador.Decode(new(any)); err != io.EOF {
		return invalida("Envia un unico objeto JSON.")
	}
	if conMinutos && entrada.Minutos == nil {
		return invalida("Indica los minutos.")
	}
	if !conMinutos && entrada.Minutos != nil {
		return invalida("Esta accion no recibe minutos.")
	}
	if !conMotivo && entrada.Motivo != "" {
		return invalida("Esta accion no recibe motivo.")
	}
	return entrada, true
}

func responderErrorMinutos(w http.ResponseWriter, err error) {
	tabla := []struct {
		err    error
		estado int
		codigo string
	}{
		{modelo.ErrSesionNoEncontrada, http.StatusNotFound, "sesion_no_encontrada"},
		{modelo.ErrMinutosNoEncontrados, http.StatusNotFound, "minutos_no_encontrados"},
		{modelo.ErrNoAutorizado, http.StatusForbidden, "no_autorizado"},
		{modelo.ErrAutoaprobacion, http.StatusForbidden, "autoaprobacion"},
		{modelo.ErrMinutosDuplicados, http.StatusConflict, "minutos_duplicados"},
		{modelo.ErrTransicionMinutos, http.StatusConflict, "transicion_invalida"},
		{modelo.ErrDisputaNoPermitida, http.StatusConflict, "disputa_no_permitida"},
		{modelo.ErrSesionCancelada, http.StatusConflict, "sesion_cancelada"},
		{modelo.ErrMinutosInvalidos, http.StatusUnprocessableEntity, "minutos_invalidos"},
		{modelo.ErrMotivoInvalido, http.StatusUnprocessableEntity, "motivo_invalido"},
		{modelo.ErrAsistenciaIncompleta, http.StatusUnprocessableEntity, "asistencia_incompleta"},
		{modelo.ErrSesionNoFinalizada, http.StatusUnprocessableEntity, "sesion_no_finalizada"},
		{modelo.ErrFueraDePlazo, http.StatusUnprocessableEntity, "fuera_de_plazo"},
		{modelo.ErrSesionInvalida, http.StatusUnprocessableEntity, "sesion_invalida"},
	}
	for _, fila := range tabla {
		if errors.Is(err, fila.err) {
			responderError(w, detalleErrorHTTP{estado: fila.estado, codigo: fila.codigo, mensaje: err.Error()})
			return
		}
	}
	responderError(w, detalleErrorHTTP{estado: http.StatusInternalServerError, codigo: "error_interno", mensaje: "Ocurrio un error inesperado."})
}
