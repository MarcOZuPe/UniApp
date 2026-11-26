package acceso

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// B-09: si el payload no se puede codificar como JSON no debe emitirse una
// respuesta 2xx con cuerpo vacío; el fallo debe traducirse a un 500.
func TestB09EscribirJSONNoDebeConfirmarRespuestaSiNoPuedeCodificar(t *testing.T) {
	respuesta := httptest.NewRecorder()

	EscribirJSON(respuesta, http.StatusOK, make(chan int))

	if respuesta.Code != http.StatusInternalServerError {
		t.Fatalf(
			"B-09: un payload no codificable debe producir 500, se obtuvo %d con cuerpo %q",
			respuesta.Code, respuesta.Body.String(),
		)
	}
}
