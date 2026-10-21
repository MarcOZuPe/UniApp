package acceso

import (
	"errors"
	"net/http"

	vista "uniapp/internal/views/acceso"
)

type reglaDeError struct {
	coincide  func(error) bool
	respuesta func(error) vista.RespuestaError
}

// MapeadorDeErrores permite extender las respuestas HTTP sin modificar el
// controlador: cada flujo registra sus propios errores.
type MapeadorDeErrores struct {
	reglas []reglaDeError
}

func NuevoMapeadorDeErrores() *MapeadorDeErrores {
	return &MapeadorDeErrores{}
}

func (m *MapeadorDeErrores) Registrar(coincide func(error) bool, respuesta func(error) vista.RespuestaError) {
	m.reglas = append(m.reglas, reglaDeError{coincide: coincide, respuesta: respuesta})
}

func (m *MapeadorDeErrores) Escribir(w http.ResponseWriter, err error) {
	for _, regla := range m.reglas {
		if regla.coincide(err) {
			vista.EscribirError(w, regla.respuesta(err))
			return
		}
	}
	vista.EscribirError(w, vista.RespuestaError{
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
