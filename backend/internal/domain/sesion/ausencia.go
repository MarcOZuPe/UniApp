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

// ReportarAusencia es el comando de un participante para declarar su propia
// ausencia dentro de la ventana de la sesion. I11 no especifica permisos para
// reportar por otra persona; no se amplian los permisos de los participantes.
func (s *Sesion) ReportarAusencia(usuarioID string, ahora time.Time) error {
	if err := s.autorizarParticipante(usuarioID, ahora); err != nil {
		return err
	}
	if err := exigirDentroDe(s.VentanaDeclaracion(), ahora); err != nil {
		return err
	}
	return s.agregarAusencia(usuarioID, ahora)
}

// RestaurarAusencia reconstituye una ausencia persistida sin aplicar la
// politica temporal.
func (s *Sesion) RestaurarAusencia(usuarioID string, reportadaEn time.Time) error {
	return s.agregarAusencia(usuarioID, reportadaEn)
}

func (s *Sesion) agregarAusencia(usuarioID string, momento time.Time) error {
	if err := s.validarMomento(momento); err != nil {
		return err
	}
	rol, err := s.rolDe(usuarioID)
	if err != nil {
		return err
	}
	if s.tieneAusenciaDe(usuarioID) {
		return ErrAusenciaDuplicada
	}
	if s.tieneAsistenciaDe(usuarioID) {
		return ErrDeclaracionContradictoria
	}

	s.ausencias = append(s.ausencias, Ausencia{
		usuarioID:   usuarioID,
		rol:         rol,
		reportadaEn: momento.UTC(),
	})
	return nil
}

func (s *Sesion) tieneAusenciaDe(usuarioID string) bool {
	for _, ausencia := range s.ausencias {
		if ausencia.usuarioID == usuarioID {
			return true
		}
	}
	return false
}
