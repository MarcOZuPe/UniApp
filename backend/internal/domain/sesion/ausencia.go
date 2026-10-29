package sesion

import (
	"errors"
	"time"
)

var (
	ErrAusenciaDuplicada         = errors.New("ausencia ya reportada")
	ErrDeclaracionContradictoria = errors.New("no se puede declarar asistencia y ausencia para el mismo participante")
)

// Ausencia es una declaracion propia: usuarioID identifica tanto al ausente
// como a quien reporta. La identidad debe proceder de la autenticacion.
type Ausencia struct {
	usuarioID   string
	rol         Rol
	reportadaEn time.Time
}

func (a Ausencia) UsuarioID() string      { return a.usuarioID }
func (a Ausencia) Rol() Rol               { return a.rol }
func (a Ausencia) ReportadaEn() time.Time { return a.reportadaEn }

func (s *Sesion) Ausencias() []Ausencia {
	resultado := make([]Ausencia, len(s.ausencias))
	copy(resultado, s.ausencias)
	return resultado
}

// ReportarAusencia solo admite la ausencia propia. I11 no especifica permisos
// para reportar por otra persona; no se amplian los permisos de los participantes.
func (s *Sesion) ReportarAusencia(usuarioID string, ahora time.Time) error {
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

	for _, ausencia := range s.ausencias {
		if ausencia.usuarioID == usuarioID {
			return ErrAusenciaDuplicada
		}
	}
	for _, asistencia := range s.asistencias {
		if asistencia.usuarioID == usuarioID {
			return ErrDeclaracionContradictoria
		}
	}

	s.ausencias = append(s.ausencias, Ausencia{
		usuarioID:   usuarioID,
		rol:         rol,
		reportadaEn: ahora.UTC(),
	})
	return nil
}
