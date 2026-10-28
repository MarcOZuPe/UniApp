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

func (s *Sesion) DeclararAsistencia(
	usuarioID string,
	ahora time.Time,
) error {
	if err := s.Validar(); err != nil {
		return err
	}

	if ahora.IsZero() {
		return ErrSesionInvalida
	}

	var rol Rol

	switch usuarioID {
	case s.tutorID:
		rol = RolTutor
	case s.escolarID:
		rol = RolEscolar
	default:
		return ErrNoAutorizado
	}

	for _, asistencia := range s.asistencias {
		if asistencia.usuarioID == usuarioID {
			return ErrAsistenciaDuplicada
		}
	}

	s.asistencias = append(s.asistencias, Asistencia{
		usuarioID:   usuarioID,
		rol:         rol,
		declaradaEn: ahora.UTC(),
	})

	return nil
}

// Clonar devuelve una copia independiente de la sesión y de sus asistencias.
func (s *Sesion) Clonar() *Sesion {
	copia := *s
	copia.asistencias = append([]Asistencia(nil), s.asistencias...)
	return &copia
}
