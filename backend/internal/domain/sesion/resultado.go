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

// RegistrarResultado es el comando del tutor asignado: un unico registro,
// dentro de la ventana de resultado. No altera las declaraciones de los
// participantes ni verifica minutos.
func (s *Sesion) RegistrarResultado(usuarioID, resumen string, ahora time.Time) error {
	if err := s.autorizarTutor(usuarioID, ahora); err != nil {
		return err
	}
	if err := exigirDentroDe(s.VentanaResultado(), ahora); err != nil {
		return err
	}
	return s.fijarResultado(usuarioID, resumen, ahora)
}

// RestaurarResultado reconstituye un resultado persistido sin aplicar la
// politica temporal.
func (s *Sesion) RestaurarResultado(usuarioID, resumen string, registradoEn time.Time) error {
	return s.fijarResultado(usuarioID, resumen, registradoEn)
}

func (s *Sesion) autorizarTutor(usuarioID string, momento time.Time) error {
	if err := s.validarMomento(momento); err != nil {
		return err
	}
	if usuarioID != s.tutorID {
		return ErrNoAutorizado
	}
	return nil
}

func (s *Sesion) fijarResultado(usuarioID, resumen string, momento time.Time) error {
	if err := s.autorizarTutor(usuarioID, momento); err != nil {
		return err
	}
	if s.resultado != nil {
		return ErrResultadoDuplicado
	}
	resumen = strings.TrimSpace(resumen)
	if !resumenValido(resumen) {
		return ErrResultadoInvalido
	}
	s.resultado = &Resultado{usuarioID: usuarioID, resumen: resumen, registradoEn: momento.UTC()}
	return nil
}

func resumenValido(resumen string) bool {
	return resumen != "" &&
		utf8.ValidString(resumen) &&
		utf8.RuneCountInString(resumen) <= MaxCaracteresResultado
}
