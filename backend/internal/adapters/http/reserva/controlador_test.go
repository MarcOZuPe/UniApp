package reserva

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// B-12: el controlador de reservas decodifica un único objeto pero no verifica
// que el resto del cuerpo esté vacío; un JSON adicional se ignora en silencio.
func TestB12ReservaRechazaContenidoJSONAdicional(t *testing.T) {
	peticion := httptest.NewRequest(
		http.MethodPost,
		"/api/reserva",
		strings.NewReader(`{"slot_id":"abc"} {"slot_id":"otro"}`),
	)
	peticion.Header.Set("Content-Type", "application/json")
	respuesta := httptest.NewRecorder()

	var destino struct {
		SlotID string `json:"slot_id"`
	}
	if decodificarCuerpo(respuesta, peticion, &destino) {
		t.Fatalf("B-12: reservas no debe aceptar JSON adicional; cuerpo=%q", respuesta.Body.String())
	}
	if respuesta.Code != http.StatusBadRequest {
		t.Fatalf("B-12: se esperaba 400, se obtuvo %d", respuesta.Code)
	}
}
