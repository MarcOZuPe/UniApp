package acceso

import (
	"errors"
	"time"
)

// DatosRegistro es la entrada del caso de uso de registro. Vive en el dominio
// porque los perfiles de rol la usan para validar campos adicionales.
type DatosRegistro struct {
	Email, NombreUso, Clave string
	Rol                     Rol
}

type Usuario struct {
	ID              string
	Email           string
	NombreUso       string
	ClaveHash       string
	Rol             Rol
	Activo          bool
	EmailConfirmado bool
	CreadoEn        time.Time
	UltimoIngreso   *time.Time
}

type Sesion struct {
	TokenHash []byte
	UsuarioID string
	ExpiraEn  time.Time
}

type TokenRecuperacion struct {
	TokenHash []byte
	UsuarioID string
	ExpiraEn  time.Time
}

type TokenConfirmacion struct {
	TokenHash []byte
	UsuarioID string
	ExpiraEn  time.Time
}

var (
	ErrEmailEnUso         = errors.New("el correo ya está registrado")
	ErrCredenciales       = errors.New("correo o contraseña incorrectos")
	ErrNoEncontrado       = errors.New("no encontrado")
	ErrTokenInvalido      = errors.New("enlace inválido o vencido")
	ErrRolNoPermitido     = errors.New("rol no permitido para registro público")
	ErrDemasiadosIntentos = errors.New("demasiados intentos; espera unos minutos")
)

type ErrValidacion struct{ Campos map[string]string }

func (e ErrValidacion) Error() string { return "datos inválidos" }
