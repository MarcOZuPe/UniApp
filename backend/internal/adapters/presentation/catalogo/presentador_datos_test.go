package catalogo

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	modelo "uniapp/internal/application/catalogo"
)

func TestPresentarImportacionYSolicitudes(t *testing.T) {
	importacion := &modelo.Importacion{
		ID: 1, Fuente: "origen", ImportadoEn: time.Now(),
		Cursos:   []modelo.Curso{{ID: 1, Nombre: "Álgebra", Grado: "5to", Estado: modelo.EstadoPendiente, Motivo: "motivo"}},
		Colegios: []modelo.Colegio{{Nombre: "Colegio", Distrito: "Lima"}},
	}
	presentada := PresentarImportacion(importacion)
	if presentada.ID != 1 || presentada.Cursos[0].Estado != "pendiente" || presentada.Colegios[0].Distrito != "Lima" {
		t.Fatalf("importación presentada inesperada: %+v", presentada)
	}
	solicitudes := PresentarSolicitudes([]modelo.SolicitudColegio{{ID: 2, Nombre: "Colegio", Distrito: "Lima", SolicitadoEn: time.Now()}})
	if len(solicitudes) != 1 || solicitudes[0].ID != 2 {
		t.Fatalf("solicitudes presentadas inesperadas: %+v", solicitudes)
	}
}

func TestEscribirJSONYError(t *testing.T) {
	respuesta := httptest.NewRecorder()
	EscribirJSON(respuesta, 200, map[string]string{"estado": "ok"})
	if respuesta.Code != 200 || respuesta.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("respuesta inesperada: %d %q", respuesta.Code, respuesta.Header().Get("Content-Type"))
	}
	var cuerpo map[string]string
	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil || cuerpo["estado"] != "ok" {
		t.Fatalf("cuerpo inesperado: %q err=%v", respuesta.Body.String(), err)
	}
	errorRespuesta := httptest.NewRecorder()
	EscribirError(errorRespuesta, 422, "validacion", "datos inválidos", map[string]string{"campo": "error"})
	if errorRespuesta.Code != 422 {
		t.Fatalf("estado inesperado: %d", errorRespuesta.Code)
	}
}
