package acceso

import (
	"errors"
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

func registrarErroresDeCuenta(errorMapper *MapeadorDeErrores) {
	registrarErrorDeValidacion(errorMapper)
	registrarErrorDeCorreoEnUso(errorMapper)
	registrarErrorDeRolNoPermitido(errorMapper)
	registrarErrorDeCredenciales(errorMapper)
	registrarErrorDeDemasiadosIntentos(errorMapper)
	registrarErrorDeTokenInvalido(errorMapper)
}

func registrarErrorDeValidacion(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			var validationError modelo.ErrValidacion
			return errors.As(err, &validationError)
		},
		func(err error) vista.RespuestaError {
			var validationError modelo.ErrValidacion
			_ = errors.As(err, &validationError)

			return vista.RespuestaError{
				Estado:  http.StatusUnprocessableEntity,
				Codigo:  "validacion",
				Mensaje: "Revisa los datos ingresados.",
				Campos:  validationError.Campos,
			}
		},
	)
}

func registrarErrorDeCorreoEnUso(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrEmailEnUso)
		},
		respuestaConCampos(http.StatusConflict, "correo_en_uso", "email"),
	)
}

func registrarErrorDeRolNoPermitido(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrRolNoPermitido)
		},
		respuestaConEstado(http.StatusForbidden, "rol_no_permitido"),
	)
}

func registrarErrorDeCredenciales(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrCredenciales)
		},
		respuestaConEstado(http.StatusUnauthorized, "credenciales"),
	)
}

func registrarErrorDeDemasiadosIntentos(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrDemasiadosIntentos)
		},
		respuestaConEstado(http.StatusTooManyRequests, "demasiados_intentos"),
	)
}

func registrarErrorDeTokenInvalido(errorMapper *MapeadorDeErrores) {
	errorMapper.Registrar(
		func(err error) bool {
			return errors.Is(err, modelo.ErrTokenInvalido)
		},
		respuestaConEstado(http.StatusBadRequest, "token_invalido"),
	)
}
