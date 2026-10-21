package reserva

import (
	"errors"
	"net/http"

	ctrlAcceso "uniapp/internal/controllers/acceso"
	modelo "uniapp/internal/models/reserva"
	vistaAcceso "uniapp/internal/views/acceso"
)

func registrarErroresDeReserva(mapeador *ctrlAcceso.MapeadorDeErrores) {
	mapeador.Registrar(func(err error) bool {
		var validacion modelo.ErrValidacion
		return errors.As(err, &validacion)
	}, func(err error) vistaAcceso.RespuestaError {
		var validacion modelo.ErrValidacion
		_ = errors.As(err, &validacion)
		return vistaAcceso.RespuestaError{
			Estado: http.StatusUnprocessableEntity, Codigo: "validacion",
			Mensaje: "Revisa los datos de la reserva.", Campos: validacion.Campos,
		}
	})
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrAnticipacion) },
		respuestaConCampo(http.StatusUnprocessableEntity, "anticipacion_insuficiente", "slot_id"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrHorizonte) },
		respuestaConCampo(http.StatusUnprocessableEntity, "horizonte_excedido", "slot_id"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrSlotInexistente) },
		respuestaConEstado(http.StatusNotFound, "slot_inexistente"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrSlotVencido) },
		respuestaConEstado(http.StatusConflict, "slot_vencido"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrSlotOcupado) },
		respuestaConCampo(http.StatusConflict, "slot_ocupado", "slot_id"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrReservaNoEncontrada) },
		respuestaConEstado(http.StatusNotFound, "reserva_inexistente"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrSinAutorizacion) },
		respuestaConEstado(http.StatusForbidden, "sin_autorizacion"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, modelo.ErrNoAutorizado) },
		func(error) vistaAcceso.RespuestaError {
			return vistaAcceso.RespuestaError{
				Estado: http.StatusForbidden, Codigo: "no_autorizado",
				Mensaje: "No puedes acceder a esta reserva.",
			}
		},
	)
}

func respuestaConEstado(estado int, codigo string) func(error) vistaAcceso.RespuestaError {
	return func(err error) vistaAcceso.RespuestaError {
		return vistaAcceso.RespuestaError{Estado: estado, Codigo: codigo, Mensaje: err.Error()}
	}
}

func respuestaConCampo(estado int, codigo, campo string) func(error) vistaAcceso.RespuestaError {
	return func(err error) vistaAcceso.RespuestaError {
		return vistaAcceso.RespuestaError{
			Estado: estado, Codigo: codigo, Mensaje: err.Error(),
			Campos: map[string]string{campo: err.Error()},
		}
	}
}
