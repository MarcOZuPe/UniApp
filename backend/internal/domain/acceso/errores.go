package acceso

import "errors"

var (
	ErrEmailEnUso         = errors.New("el correo ya está registrado")
	ErrCredenciales       = errors.New("correo o contraseña incorrectos")
	ErrNoEncontrado       = errors.New("no encontrado")
	ErrTokenInvalido      = errors.New("enlace inválido o vencido")
	ErrRolNoPermitido     = errors.New("rol no permitido para registro público")
	ErrDemasiadosIntentos = errors.New("demasiados intentos; espera unos minutos")
)
