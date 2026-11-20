package incidente

import (
	"errors"
	"net/http"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vistaAcceso "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/incidente"
)

func registrarErroresDeIncidente(mapeador *ctrlAcceso.MapeadorDeErrores) {
	mapeador.Registrar(func(err error) bool {
		var validacion modelo.ErrorValidacion
		return errors.As(err, &validacion)
	}, func(err error) vistaAcceso.RespuestaError {
		var validacion modelo.ErrorValidacion
		_ = errors.As(err, &validacion)
		return vistaAcceso.RespuestaError{
			Estado: http.StatusUnprocessableEntity, Codigo: "validacion",
			Mensaje: "Revisa los datos del incidente.", Campos: validacion.Campos,
		}
	})
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrNoAutorizado) },
		respuestaConEstado(http.StatusForbidden, "no_autorizado", "No puedes acceder a este incidente."),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrEscolarNoVinculado) },
		respuestaConEstado(http.StatusForbidden, "escolar_no_vinculado", "El escolar no está vinculado a tu cuenta."),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrTutoriaNoEncontrada) },
		func(error) vistaAcceso.RespuestaError {
			return vistaAcceso.RespuestaError{
				Estado: http.StatusUnprocessableEntity, Codigo: "tutoria_invalida",
				Mensaje: "La tutoría debe haber finalizado y pertenecer al escolar.",
				Campos:  map[string]string{"reserva_id": "Elige una tutoría previa del escolar."},
			}
		},
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrIncidenteNoEncontrado) },
		respuestaConEstado(http.StatusNotFound, "incidente_inexistente", "El incidente no existe."),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrEvidenciaNoEncontrada) },
		respuestaConEstado(http.StatusNotFound, "evidencia_inexistente", "La evidencia no existe."),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrYaRevisado) },
		respuestaConEstado(http.StatusConflict, "incidente_ya_revisado", "El incidente ya fue revisado."),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrEvidenciaInvalida) },
		func(error) vistaAcceso.RespuestaError {
			return vistaAcceso.RespuestaError{
				Estado: http.StatusUnprocessableEntity, Codigo: "evidencia_invalida",
				Mensaje: "La evidencia debe ser PDF, JPG o PNG de hasta 10 MB.",
				Campos:  map[string]string{"evidencias": "Adjunta archivos PDF, JPG o PNG válidos."},
			}
		},
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrAlmacen) },
		respuestaConEstado(http.StatusInternalServerError, "almacen", "No se pudo guardar la evidencia. Intenta de nuevo."),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrIncidenteInvalido) },
		respuestaConEstado(http.StatusUnprocessableEntity, "incidente_invalido", "El incidente no es válido."),
	)
}

func respuestaConEstado(estado int, codigo, mensaje string) func(error) vistaAcceso.RespuestaError {
	return func(error) vistaAcceso.RespuestaError {
		return vistaAcceso.RespuestaError{Estado: estado, Codigo: codigo, Mensaje: mensaje}
	}
}
