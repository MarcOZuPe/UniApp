package acceso

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodificarCuerpoRechazaCamposDesconocidosYJSONInvalido(t *testing.T) {
	base := &controladorBase{errores: NuevoMapeadorDeErrores()}

	for _, cuerpo := range []string{`{"email":"a","campo_extra":true}`, `{`, ``, `{"email":"a"} {"email":"b"}`} {
		respuesta := httptest.NewRecorder()
		peticion := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(cuerpo))
		var destino struct {
			Email string `json:"email"`
		}
		if base.decodificarCuerpo(respuesta, peticion, &destino) {
			t.Fatalf("cuerpo %q debía rechazarse", cuerpo)
		}
		if respuesta.Code != http.StatusBadRequest ||
			!strings.Contains(respuesta.Body.String(), `"codigo":"solicitud_invalida"`) {
			t.Fatalf("cuerpo %q: %d %s", cuerpo, respuesta.Code, respuesta.Body)
		}
	}
}
