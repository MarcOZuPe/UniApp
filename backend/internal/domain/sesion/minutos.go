package sesion

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// I17: minutos verificados y ajustes auditados.
//
// Un RegistroMinutos nace cuando el tutor propone los minutos realmente
// dictados. Nada suma al total hasta que un coordinador lo verifica: no hay
// autoaprobacion. Una disputa congela el registro (suma 0) y una reserva
// cancelada lo anula (suma 0). Cada accion deja una entrada de auditoria
// inmutable.

// MaxMinutosSesion es el tope de minutos de una sesion (60).
const MaxMinutosSesion = int(DuracionSesion / time.Minute)

// MaxCaracteresMotivo acota los motivos de ajuste, disputa y resolucion.
const MaxCaracteresMotivo = 500

// ActorSistema identifica las acciones automaticas (anulacion por cancelacion).
const ActorSistema = "sistema"

type EstadoMinutos string

const (
	MinutosPropuestos  EstadoMinutos = "propuesto"
	MinutosEnDisputa   EstadoMinutos = "en_disputa"
	MinutosVerificados EstadoMinutos = "verificado"
	MinutosAnulados    EstadoMinutos = "anulado"
)

type AccionMinutos string

const (
	AccionProponer AccionMinutos = "proponer"
	AccionAprobar  AccionMinutos = "aprobar"
	AccionAjustar  AccionMinutos = "ajustar"
	AccionDisputar AccionMinutos = "disputar"
	AccionResolver AccionMinutos = "resolver"
	AccionAnular   AccionMinutos = "anular"
)

var (
	ErrMinutosInvalidos       = errors.New("los minutos deben estar entre 1 y 60 al proponer, y entre 0 y 60 al verificar")
	ErrAsistenciaIncompleta   = errors.New("tutor y escolar deben haber declarado asistencia para proponer minutos")
	ErrSesionNoFinalizada     = errors.New("los minutos solo se proponen despues del fin de la sesion")
	ErrMinutosDuplicados      = errors.New("los minutos de esta sesion ya fueron propuestos")
	ErrMinutosNoEncontrados   = errors.New("la sesion no tiene minutos propuestos")
	ErrTransicionMinutos      = errors.New("la accion no es valida en el estado actual de los minutos")
	ErrAutoaprobacion         = errors.New("un participante de la sesion no puede verificar sus propios minutos")
	ErrMotivoInvalido         = errors.New("el motivo debe contener entre 1 y 500 caracteres")
	ErrSesionCancelada        = errors.New("la reserva de la sesion esta cancelada; sus minutos suman cero")
	ErrDisputaNoPermitida     = errors.New("esta disputa no esta permitida")
	ErrAuditoriaInconsistente = errors.New("auditoria de minutos inconsistente")
)

// EstadoReservas es el puerto con el que el flujo E pregunta si una reserva
// fue cancelada. El dato pertenece al flujo D: hoy lo responde un fixture y,
// cuando exista I16, un adaptador de ese flujo, sin cambiar estas reglas.
type EstadoReservas interface {
	Cancelada(ctx context.Context, reservaID string) (bool, error)
}

// EntradaAuditoria es un hecho inmutable sobre el registro de minutos.
type EntradaAuditoria struct {
	accion        AccionMinutos
	actorID       string
	valorAnterior *int
	valorNuevo    *int
	motivo        string
	registradaEn  time.Time
}

func (e EntradaAuditoria) Accion() AccionMinutos   { return e.accion }
func (e EntradaAuditoria) ActorID() string         { return e.actorID }
func (e EntradaAuditoria) Motivo() string          { return e.motivo }
func (e EntradaAuditoria) RegistradaEn() time.Time { return e.registradaEn }

func (e EntradaAuditoria) ValorAnterior() (int, bool) { return valorDe(e.valorAnterior) }
func (e EntradaAuditoria) ValorNuevo() (int, bool)    { return valorDe(e.valorNuevo) }

func valorDe(p *int) (int, bool) {
	if p == nil {
		return 0, false
	}
	return *p, true
}

func entero(v int) *int { return &v }

// RegistroMinutos es el agregado de I17. Sus campos son privados para que
// solo cambie a traves de los comandos que aplican las reglas.
type RegistroMinutos struct {
	sesionID   string
	reservaID  string
	tutorID    string
	escolarID  string
	finSesion  time.Time
	estado     EstadoMinutos
	propuestos int
	// verificados vale solo en estado verificado; en disputa conserva el
	// ultimo valor verificado (si lo hubo) para que el tutor vea el ajuste.
	verificados  *int
	disputadoPor string
	auditoria    []EntradaAuditoria
}

// ProponerMinutos crea el registro: lo propone el tutor asignado, despues del
// fin de la sesion, dentro del plazo de resultado y solo si ambos
// participantes declararon asistencia (un enlace abierto no acredita).
func ProponerMinutos(s *Sesion, tutorID string, minutos int, ahora time.Time) (*RegistroMinutos, error) {
	if err := s.autorizarTutor(tutorID, ahora); err != nil {
		return nil, err
	}
	if ahora.Before(s.fin) {
		return nil, ErrSesionNoFinalizada
	}
	if err := exigirDentroDe(s.VentanaResultado(), ahora); err != nil {
		return nil, err
	}
	if !s.tieneAsistenciaDe(s.tutorID) || !s.tieneAsistenciaDe(s.escolarID) {
		return nil, ErrAsistenciaIncompleta
	}
	if minutos < 1 || minutos > MaxMinutosSesion {
		return nil, ErrMinutosInvalidos
	}
	r := &RegistroMinutos{
		sesionID:   s.id,
		reservaID:  s.reservaID,
		tutorID:    s.tutorID,
		escolarID:  s.escolarID,
		finSesion:  s.fin,
		estado:     MinutosPropuestos,
		propuestos: minutos,
	}
	r.auditar(AccionProponer, tutorID, nil, entero(minutos), "", ahora)
	return r, nil
}

// Aprobar verifica los minutos propuestos tal como estan.
func (r *RegistroMinutos) Aprobar(coordinadorID string, ahora time.Time) error {
	if err := r.exigirCoordinador(coordinadorID, ahora); err != nil {
		return err
	}
	if r.estado != MinutosPropuestos {
		return ErrTransicionMinutos
	}
	r.verificar(AccionAprobar, coordinadorID, r.propuestos, "", ahora)
	return nil
}

// Ajustar fija un valor distinto (0-60) sobre un registro propuesto o ya
// verificado. Exige motivo y queda auditado con el valor anterior.
func (r *RegistroMinutos) Ajustar(coordinadorID string, minutos int, motivo string, ahora time.Time) error {
	if err := r.exigirCoordinador(coordinadorID, ahora); err != nil {
		return err
	}
	if r.estado != MinutosPropuestos && r.estado != MinutosVerificados {
		return ErrTransicionMinutos
	}
	if minutos < 0 || minutos > MaxMinutosSesion {
		return ErrMinutosInvalidos
	}
	motivo, err := normalizarMotivo(motivo)
	if err != nil {
		return err
	}
	r.verificar(AccionAjustar, coordinadorID, minutos, motivo, ahora)
	return nil
}

// Disputar congela el registro. El escolar puede disputar lo propuesto o lo
// verificado; el tutor solo un verificado menor que lo que propuso. Rige el
// mismo plazo que el resultado de la sesion.
func (r *RegistroMinutos) Disputar(usuarioID, motivo string, ahora time.Time) error {
	if ahora.IsZero() {
		return ErrSesionInvalida
	}
	if r.estado == MinutosAnulados {
		return ErrSesionCancelada
	}
	if !r.Participa(usuarioID) {
		return ErrNoAutorizado
	}
	// La resolucion de una disputa es final: evita disputas en bucle.
	if r.fueResuelto() {
		return ErrDisputaNoPermitida
	}
	switch usuarioID {
	case r.escolarID:
		if r.estado != MinutosPropuestos && r.estado != MinutosVerificados {
			return ErrDisputaNoPermitida
		}
	case r.tutorID:
		v, ok := valorDe(r.verificados)
		if r.estado != MinutosVerificados || !ok || v >= r.propuestos {
			return ErrDisputaNoPermitida
		}
	default:
		return ErrNoAutorizado
	}
	if ahora.After(r.finSesion.Add(PlazoResultado)) {
		return ErrFueraDePlazo
	}
	motivo, err := normalizarMotivo(motivo)
	if err != nil {
		return err
	}
	r.auditar(AccionDisputar, usuarioID, r.verificados, nil, motivo, ahora)
	r.estado = MinutosEnDisputa
	r.disputadoPor = usuarioID
	return nil
}

// Resolver cierra una disputa fijando los minutos finales (0-60) con motivo.
func (r *RegistroMinutos) Resolver(coordinadorID string, minutos int, motivo string, ahora time.Time) error {
	if err := r.exigirCoordinador(coordinadorID, ahora); err != nil {
		return err
	}
	if r.estado != MinutosEnDisputa {
		return ErrTransicionMinutos
	}
	if minutos < 0 || minutos > MaxMinutosSesion {
		return ErrMinutosInvalidos
	}
	motivo, err := normalizarMotivo(motivo)
	if err != nil {
		return err
	}
	r.verificar(AccionResolver, coordinadorID, minutos, motivo, ahora)
	r.disputadoPor = ""
	return nil
}

// Anular deja el registro en cero porque la reserva fue cancelada. Es
// idempotente: anular dos veces no duplica la auditoria.
func (r *RegistroMinutos) Anular(ahora time.Time) error {
	if ahora.IsZero() {
		return ErrSesionInvalida
	}
	if r.estado == MinutosAnulados {
		return nil
	}
	r.auditar(AccionAnular, ActorSistema, r.verificados, entero(0), "reserva cancelada", ahora)
	r.estado = MinutosAnulados
	r.disputadoPor = ""
	return nil
}

// MinutosComputables es lo que este registro aporta al total: solo los
// verificados; propuestos, en disputa y anulados suman 0.
func (r *RegistroMinutos) MinutosComputables() int {
	if r.estado != MinutosVerificados {
		return 0
	}
	v, _ := valorDe(r.verificados)
	return v
}

func (r *RegistroMinutos) SesionID() string                { return r.sesionID }
func (r *RegistroMinutos) ReservaID() string               { return r.reservaID }
func (r *RegistroMinutos) TutorID() string                 { return r.tutorID }
func (r *RegistroMinutos) EscolarID() string               { return r.escolarID }
func (r *RegistroMinutos) FinSesion() time.Time            { return r.finSesion }
func (r *RegistroMinutos) Estado() EstadoMinutos           { return r.estado }
func (r *RegistroMinutos) MinutosPropuestos() int          { return r.propuestos }
func (r *RegistroMinutos) MinutosVerificados() (int, bool) { return valorDe(r.verificados) }
func (r *RegistroMinutos) DisputadoPor() string            { return r.disputadoPor }

// Auditoria devuelve una copia: la historia no se puede reescribir desde fuera.
func (r *RegistroMinutos) Auditoria() []EntradaAuditoria {
	copia := make([]EntradaAuditoria, len(r.auditoria))
	copy(copia, r.auditoria)
	return copia
}

// Participa indica si el usuario es el tutor o el escolar de la sesion.
func (r *RegistroMinutos) Participa(usuarioID string) bool {
	return usuarioID == r.tutorID || usuarioID == r.escolarID
}

// Clonar devuelve una copia independiente del registro.
func (r *RegistroMinutos) Clonar() *RegistroMinutos {
	copia := *r
	copia.auditoria = r.Auditoria()
	if r.verificados != nil {
		copia.verificados = entero(*r.verificados)
	}
	return &copia
}

func (r *RegistroMinutos) fueResuelto() bool {
	for _, e := range r.auditoria {
		if e.accion == AccionResolver {
			return true
		}
	}
	return false
}

func (r *RegistroMinutos) exigirCoordinador(coordinadorID string, ahora time.Time) error {
	if ahora.IsZero() || strings.TrimSpace(coordinadorID) == "" {
		return ErrSesionInvalida
	}
	if r.estado == MinutosAnulados {
		return ErrSesionCancelada
	}
	if r.Participa(coordinadorID) {
		return ErrAutoaprobacion
	}
	return nil
}

func (r *RegistroMinutos) verificar(accion AccionMinutos, actorID string, minutos int, motivo string, ahora time.Time) {
	r.auditar(accion, actorID, r.verificados, entero(minutos), motivo, ahora)
	r.verificados = entero(minutos)
	r.estado = MinutosVerificados
}

func (r *RegistroMinutos) auditar(accion AccionMinutos, actorID string, anterior, nuevo *int, motivo string, ahora time.Time) {
	var a, n *int
	if anterior != nil {
		a = entero(*anterior)
	}
	if nuevo != nil {
		n = entero(*nuevo)
	}
	r.auditoria = append(r.auditoria, EntradaAuditoria{
		accion: accion, actorID: actorID, valorAnterior: a, valorNuevo: n,
		motivo: motivo, registradaEn: ahora.UTC(),
	})
}

func normalizarMotivo(motivo string) (string, error) {
	motivo = strings.TrimSpace(motivo)
	if motivo == "" || !utf8.ValidString(motivo) || utf8.RuneCountInString(motivo) > MaxCaracteresMotivo {
		return "", ErrMotivoInvalido
	}
	return motivo, nil
}

// DatosRegistroMinutos permite reconstituir un registro persistido.
type DatosRegistroMinutos struct {
	SesionID, ReservaID, TutorID, EscolarID string
	FinSesion                               time.Time
	Estado                                  EstadoMinutos
	Propuestos                              int
	Verificados                             *int
	DisputadoPor                            string
	Auditoria                               []DatosEntradaAuditoria
}

type DatosEntradaAuditoria struct {
	Accion        AccionMinutos
	ActorID       string
	ValorAnterior *int
	ValorNuevo    *int
	Motivo        string
	RegistradaEn  time.Time
}

// RestaurarRegistroMinutos reconstituye un registro desde la persistencia y
// valida sus invariantes (sin aplicar plazos, que rigen al actuar).
func RestaurarRegistroMinutos(d DatosRegistroMinutos) (*RegistroMinutos, error) {
	if d.SesionID == "" || d.ReservaID == "" || d.TutorID == "" || d.EscolarID == "" || d.FinSesion.IsZero() {
		return nil, ErrSesionInvalida
	}
	switch d.Estado {
	case MinutosPropuestos, MinutosEnDisputa, MinutosVerificados, MinutosAnulados:
	default:
		return nil, ErrTransicionMinutos
	}
	if d.Propuestos < 1 || d.Propuestos > MaxMinutosSesion {
		return nil, ErrMinutosInvalidos
	}
	if d.Verificados != nil && (*d.Verificados < 0 || *d.Verificados > MaxMinutosSesion) {
		return nil, ErrMinutosInvalidos
	}
	if d.Estado == MinutosVerificados && d.Verificados == nil {
		return nil, ErrMinutosInvalidos
	}
	if len(d.Auditoria) == 0 || d.Auditoria[0].Accion != AccionProponer {
		return nil, ErrAuditoriaInconsistente
	}
	r := &RegistroMinutos{
		sesionID: d.SesionID, reservaID: d.ReservaID, tutorID: d.TutorID, escolarID: d.EscolarID,
		finSesion: d.FinSesion.UTC(), estado: d.Estado, propuestos: d.Propuestos,
		disputadoPor: d.DisputadoPor,
	}
	if d.Verificados != nil {
		r.verificados = entero(*d.Verificados)
	}
	for _, e := range d.Auditoria {
		r.auditar(e.Accion, e.ActorID, e.ValorAnterior, e.ValorNuevo, e.Motivo, e.RegistradaEn)
	}
	return r, nil
}

// RepositorioMinutos persiste los registros de I17.
type RepositorioMinutos interface {
	// CrearMinutos guarda un registro nuevo; ErrMinutosDuplicados si ya existe.
	CrearMinutos(ctx context.Context, registro *RegistroMinutos) error
	// MinutosPorSesion devuelve ErrMinutosNoEncontrados si no hay registro.
	MinutosPorSesion(ctx context.Context, sesionID string) (*RegistroMinutos, error)
	// ActualizarMinutos aplica cambio de forma atomica (con bloqueo) y
	// persiste el estado y las entradas de auditoria nuevas.
	ActualizarMinutos(ctx context.Context, sesionID string, cambio func(*RegistroMinutos) error) (*RegistroMinutos, error)
	// MinutosPendientes devuelve los registros propuestos o en disputa.
	MinutosPendientes(ctx context.Context) ([]*RegistroMinutos, error)
	// MinutosDeTutor devuelve todos los registros de un tutor.
	MinutosDeTutor(ctx context.Context, tutorID string) ([]*RegistroMinutos, error)
}
