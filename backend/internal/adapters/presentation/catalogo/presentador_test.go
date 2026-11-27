package catalogo

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// B-09 (repetido en catálogo): EscribirJSON confirma un 2xx antes de codificar
// y descarta el error de codificación.
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
