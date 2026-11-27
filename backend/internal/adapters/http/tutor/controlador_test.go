package tutor

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// lectorQueSuperaLimite devuelve el mismo error que produce
// http.MaxBytesReader cuando el cuerpo excede el límite permitido.
type lectorQueSuperaLimite struct{ err error }

func (lector lectorQueSuperaLimite) Read([]byte) (int, error) { return 0, lector.err }

// B-13: una solicitud multipart que supera el máximo debe responder 413, no 422.
func TestB13SolicitudMultipartDemasiadoGrandeResponde413(t *testing.T) {
	const frontera = "limitePrueba"
	prefijo := "--" + frontera + "\r\n" +
		"Content-Disposition: form-data; name=\"documentos\"; filename=\"carta.pdf\"\r\n" +
		"Content-Type: application/pdf\r\n\r\n"

	peticion := httptest.NewRequest(
		http.MethodPost,
		"/api/tutor/expediente",
		io.MultiReader(
			strings.NewReader(prefijo),
			lectorQueSuperaLimite{err: &http.MaxBytesError{Limit: maximoSolicitud}},
		),
	)
	peticion.Header.Set("Content-Type", "multipart/form-data; boundary="+frontera)

	respuesta := httptest.NewRecorder()
	(&Controlador{}).presentarExpediente(respuesta, peticion)

	if respuesta.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(
			"B-13: se esperaba 413 para una solicitud demasiado grande, se obtuvo %d con cuerpo %q",
			respuesta.Code, respuesta.Body.String(),
		)
	}
}

// B-12: el decodificador de tutor acepta un segundo objeto JSON y solo usa el
// primero, lo que permite enviar datos ambiguos.
func TestB12TutorRechazaContenidoJSONAdicional(t *testing.T) {
	peticion := httptest.NewRequest(
		http.MethodPost,
		"/api/tutor/1/revision",
		strings.NewReader(`{"estado":"aprobado"} {"estado":"rechazado"}`),
	)
	peticion.Header.Set("Content-Type", "application/json")
	respuesta := httptest.NewRecorder()

	var destino struct {
		Estado string `json:"estado"`
	}
	if decodificarJSON(respuesta, peticion, &destino) {
		t.Fatalf("B-12: el tutor no debe aceptar JSON adicional; cuerpo=%q", respuesta.Body.String())
	}
	if respuesta.Code != http.StatusBadRequest {
		t.Fatalf("B-12: se esperaba 400, se obtuvo %d", respuesta.Code)
	}
}
