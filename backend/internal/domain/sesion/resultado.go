package sesion

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const MaxCaracteresResultado = 2000

var (
	ErrResultadoInvalido  = errors.New("el resultado debe contener entre 1 y 2000 caracteres")
	ErrResultadoDuplicado = errors.New("resultado ya registrado")
)

// Resultado describe lo ocurrido; no acredita asistencia ni minutos.
type Resultado struct {
	usuarioID    string
	resumen      string
	registradoEn time.Time
}

func (r Resultado) UsuarioID() string       { return r.usuarioID }
func (r Resultado) Resumen() string         { return r.resumen }
func (r Resultado) RegistradoEn() time.Time { return r.registradoEn }

func (s *Sesion) Resultado() *Resultado {
	if s.resultado == nil {
		return nil
	}
	copia := *s.resultado
	return &copia
}

// RegistrarResultado permite un unico registro del tutor asignado. No altera
// las declaraciones independientes de cada participante ni verifica minutos.
func (s *Sesion) RegistrarResultado(usuarioID, resumen string, ahora time.Time) error {
	if err := s.Validar(); err != nil {
		return err
	}
	if usuarioID != s.tutorID {
		return ErrNoAutorizado
	}
	if s.resultado != nil {
		return ErrResultadoDuplicado
	}
	if ahora.IsZero() {
		return ErrSesionInvalida
	}
	resumen = strings.TrimSpace(resumen)
	if resumen == "" || !utf8.ValidString(resumen) || utf8.RuneCountInString(resumen) > MaxCaracteresResultado {
		return ErrResultadoInvalido
	}
	s.resultado = &Resultado{usuarioID: usuarioID, resumen: resumen, registradoEn: ahora.UTC()}
	return nil
}
