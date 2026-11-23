package acceso

import (
	"errors"
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

func registrarErroresDeCoordinadores(mapeador *MapeadorDeErrores) {
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrNoAutorizado) },
		func(error) vista.RespuestaError {
			return vista.RespuestaError{
				Estado: http.StatusForbidden, Codigo: "sin_permiso",
				Mensaje: "Solo un coordinador general puede gestionar coordinadores.",
			}
		},
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrCoordinadorNoEncontrado) },
		func(error) vista.RespuestaError {
			return vista.RespuestaError{
				Estado: http.StatusNotFound, Codigo: "coordinador_inexistente",
				Mensaje: "El coordinador no existe o ya no está activo.",
			}
		},
	)
}
