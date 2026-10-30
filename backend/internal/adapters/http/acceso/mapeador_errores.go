package acceso

import (
	"errors"
	"net/http"

	vista "uniapp/internal/adapters/presentation/acceso"
)

// Patrón aplicado: Chain of Responsibility.
// Cada regla decide si maneja el error; si no, lo delega a la siguiente.
type manejadorDeError interface {
	manejar(error) (vista.RespuestaError, bool)
	establecerSiguiente(manejadorDeError)
}

type reglaDeError struct {
	coincide  func(error) bool
	responde  func(error) vista.RespuestaError
	siguiente manejadorDeError
}

func (regla *reglaDeError) manejar(err error) (vista.RespuestaError, bool) {
	if regla.coincide(err) {
		return regla.responde(err), true
	}
	if regla.siguiente != nil {
		return regla.siguiente.manejar(err)
	}
	return vista.RespuestaError{}, false
}

func (regla *reglaDeError) establecerSiguiente(siguiente manejadorDeError) {
	regla.siguiente = siguiente
}

type MapeadorDeErrores struct {
	primeraRegla manejadorDeError
	ultimaRegla  *reglaDeError
}

func NuevoMapeadorDeErrores() *MapeadorDeErrores {
	return &MapeadorDeErrores{}
}

func (mapeador *MapeadorDeErrores) Registrar(coincide func(error) bool, responde func(error) vista.RespuestaError) {
	regla := &reglaDeError{coincide: coincide, responde: responde}
	if mapeador.primeraRegla == nil {
		mapeador.primeraRegla = regla
	} else {
		mapeador.ultimaRegla.establecerSiguiente(regla)
	}
	mapeador.ultimaRegla = regla
}

func (mapeador *MapeadorDeErrores) Escribir(respuesta http.ResponseWriter, err error) {
	if mapeador.primeraRegla != nil {
		if respuestaError, manejado := mapeador.primeraRegla.manejar(err); manejado {
			vista.EscribirError(respuesta, respuestaError)
			return
		}
	}
	vista.EscribirError(respuesta, vista.RespuestaError{
		Estado:  http.StatusInternalServerError,
		Codigo:  "error_interno",
		Mensaje: "Ocurrió un error inesperado.",
	})
}

func esAlguno(err error, objetivos ...error) bool {
	for _, objetivo := range objetivos {
		if errors.Is(err, objetivo) {
			return true
		}
	}
	return false
}

func respuestaConEstado(estado int, codigo string) func(error) vista.RespuestaError {
	return func(err error) vista.RespuestaError {
		return vista.RespuestaError{Estado: estado, Codigo: codigo, Mensaje: err.Error()}
	}
}

func respuestaConCampos(estado int, codigo, campo string) func(error) vista.RespuestaError {
	return func(err error) vista.RespuestaError {
		return vista.RespuestaError{
			Estado: estado, Codigo: codigo, Mensaje: err.Error(),
			Campos: map[string]string{campo: err.Error()},
		}
	}
}
