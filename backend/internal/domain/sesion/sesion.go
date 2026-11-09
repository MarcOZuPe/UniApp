package sesion

import (
	"errors"
	"time"
)

const DuracionSesion = 60 * time.Minute

var (
	ErrSesionInvalida      = errors.New("sesion invalida")
	ErrNoAutorizado        = errors.New("participante no autorizado")
	ErrAsistenciaDuplicada = errors.New("asistencia ya declarada")
)

type Rol string

const (
	RolTutor   Rol = "tutor"
	RolEscolar Rol = "escolar"
)

type Asistencia struct {
	usuarioID   string
	rol         Rol
	declaradaEn time.Time
}

func (a Asistencia) UsuarioID() string {
	return a.usuarioID
}

func (a Asistencia) Rol() Rol {
	return a.rol
}

func (a Asistencia) DeclaradaEn() time.Time {
	return a.declaradaEn
}

type Sesion struct {
	id          string
	reservaID   string
	tutorID     string
	escolarID   string
	inicio      time.Time
	fin         time.Time
	asistencias []Asistencia
	ausencias   []Ausencia
	resultado   *Resultado
}

func NuevaSesion(
	id string,
	reservaID string,
	tutorID string,
	escolarID string,
	inicio time.Time,
	fin time.Time,
) (*Sesion, error) {
	s := &Sesion{
		id:        id,
		reservaID: reservaID,
		tutorID:   tutorID,
		escolarID: escolarID,
		inicio:    inicio.UTC(),
		fin:       fin.UTC(),
	}

	if err := s.Validar(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sesion) Validar() error {
	if s == nil ||
		s.id == "" ||
		s.reservaID == "" ||
		s.tutorID == "" ||
		s.escolarID == "" ||
		s.tutorID == s.escolarID ||
		s.inicio.IsZero() ||
		s.fin.IsZero() ||
		s.fin.Sub(s.inicio) != DuracionSesion {
		return ErrSesionInvalida
	}

	return nil
}

func (s *Sesion) ID() string {
	return s.id
}

func (s *Sesion) ReservaID() string {
	return s.reservaID
}

func (s *Sesion) TutorID() string {
	return s.tutorID
}

func (s *Sesion) EscolarID() string {
	return s.escolarID
}

func (s *Sesion) Inicio() time.Time {
	return s.inicio
}

func (s *Sesion) Fin() time.Time {
	return s.fin
}

func (s *Sesion) Asistencias() []Asistencia {
	resultado := make([]Asistencia, len(s.asistencias))
	copy(resultado, s.asistencias)
	return resultado
}

// DeclararAsistencia es el comando de un participante: exige que la
// declaracion ocurra dentro de la ventana de la sesion.
func (s *Sesion) DeclararAsistencia(usuarioID string, ahora time.Time) error {
	if err := s.autorizarParticipante(usuarioID, ahora); err != nil {
		return err
	}
	if err := exigirDentroDe(s.VentanaDeclaracion(), ahora); err != nil {
		return err
	}
	return s.agregarAsistencia(usuarioID, ahora)
}

// RestaurarAsistencia reconstituye una asistencia ya persistida. Aplica las
// invariantes del agregado, pero no la politica temporal: esta rige al
// declarar y no debe invalidar datos historicos.
func (s *Sesion) RestaurarAsistencia(usuarioID string, declaradaEn time.Time) error {
	return s.agregarAsistencia(usuarioID, declaradaEn)
}

func (s *Sesion) agregarAsistencia(usuarioID string, momento time.Time) error {
	if err := s.validarMomento(momento); err != nil {
		return err
	}
	rol, err := s.rolDe(usuarioID)
	if err != nil {
		return err
	}
	if s.tieneAusenciaDe(usuarioID) {
		return ErrDeclaracionContradictoria
	}
	if s.tieneAsistenciaDe(usuarioID) {
		return ErrAsistenciaDuplicada
	}

	s.asistencias = append(s.asistencias, Asistencia{
		usuarioID:   usuarioID,
		rol:         rol,
		declaradaEn: momento.UTC(),
	})
	return nil
}

// autorizarParticipante comprueba sesion, momento e identidad antes de
// evaluar plazos, para no revelar detalles de la sesion a terceros.
func (s *Sesion) autorizarParticipante(usuarioID string, momento time.Time) error {
	if err := s.validarMomento(momento); err != nil {
		return err
	}
	_, err := s.rolDe(usuarioID)
	return err
}

func (s *Sesion) validarMomento(momento time.Time) error {
	if err := s.Validar(); err != nil {
		return err
	}
	if momento.IsZero() {
		return ErrSesionInvalida
	}
	return nil
}

func (s *Sesion) rolDe(usuarioID string) (Rol, error) {
	switch usuarioID {
	case s.tutorID:
		return RolTutor, nil
	case s.escolarID:
		return RolEscolar, nil
	default:
		return "", ErrNoAutorizado
	}
}

func (s *Sesion) tieneAsistenciaDe(usuarioID string) bool {
	for _, asistencia := range s.asistencias {
		if asistencia.usuarioID == usuarioID {
			return true
		}
	}
	return false
}

// Clonar devuelve una copia independiente de la sesion y sus declaraciones.
func (s *Sesion) Clonar() *Sesion {
	copia := *s
	copia.asistencias = append([]Asistencia(nil), s.asistencias...)
	copia.ausencias = append([]Ausencia(nil), s.ausencias...)
	copia.resultado = s.Resultado()
	return &copia
}
