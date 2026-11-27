package acceso

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEscribirJSONCacheablePublicaETagYResponde304(t *testing.T) {
	payload := map[string]any{"facultades": []string{"Ingeniería"}}

	primera := httptest.NewRecorder()
	EscribirJSONCacheable(primera, httptest.NewRequest(http.MethodGet, "/api/facultades", nil), http.StatusOK, payload, 60)

	if primera.Code != http.StatusOK {
		t.Fatalf("estado=%d, se esperaba 200", primera.Code)
	}
	if etag := primera.Header().Get("ETag"); etag == "" {
		t.Fatal("la respuesta cacheable debe incluir ETag")
	}
	if cache := primera.Header().Get("Cache-Control"); cache != "public, max-age=60, must-revalidate" {
		t.Fatalf("Cache-Control=%q", cache)
	}
	var decodificado map[string]any
	if err := json.Unmarshal(primera.Body.Bytes(), &decodificado); err != nil {
		t.Fatalf("el cuerpo no es JSON válido: %v", err)
	}

	segunda := httptest.NewRecorder()
	peticion := httptest.NewRequest(http.MethodGet, "/api/facultades", nil)
	peticion.Header.Set("If-None-Match", primera.Header().Get("ETag"))
	EscribirJSONCacheable(segunda, peticion, http.StatusOK, payload, 60)

	if segunda.Code != http.StatusNotModified {
		t.Fatalf("estado=%d, se esperaba 304", segunda.Code)
	}
	if segunda.Body.Len() != 0 {
		t.Fatalf("304 no debe incluir cuerpo, se obtuvo %q", segunda.Body.String())
	}

	tercera := httptest.NewRecorder()
	peticionComodin := httptest.NewRequest(http.MethodGet, "/api/facultades", nil)
	peticionComodin.Header.Set("If-None-Match", "*")
	EscribirJSONCacheable(tercera, peticionComodin, http.StatusOK, payload, 60)
	if tercera.Code != http.StatusNotModified {
		t.Fatalf("If-None-Match=* debe producir 304, se obtuvo %d", tercera.Code)
	}
}

// B-20: If-None-Match debe aceptar comparación débil para GET.
func TestB20ETagDebilDebeCoincidirEnIfNoneMatch(t *testing.T) {
	payload := map[string]any{"facultades": []string{"Ingeniería"}}
	primera := httptest.NewRecorder()
	EscribirJSONCacheable(primera, httptest.NewRequest(http.MethodGet, "/api/facultades", nil), http.StatusOK, payload, 60)

	segunda := httptest.NewRecorder()
	peticion := httptest.NewRequest(http.MethodGet, "/api/facultades", nil)
	peticion.Header.Set("If-None-Match", "W/"+primera.Header().Get("ETag"))
	EscribirJSONCacheable(segunda, peticion, http.StatusOK, payload, 60)

	if segunda.Code != http.StatusNotModified {
		t.Fatalf("B-20: un ETag débil debe producir 304 para GET, se obtuvo %d", segunda.Code)
	}
}

func TestEscribirJSONCacheableEscapaUnPayloadNoSerializable(t *testing.T) {
	respuesta := httptest.NewRecorder()
	EscribirJSONCacheable(
		respuesta,
		httptest.NewRequest(http.MethodGet, "/api/facultades", nil),
		http.StatusOK,
		map[string]any{"invalido": make(chan int)},
		60,
	)
	if respuesta.Code != http.StatusInternalServerError {
		t.Fatalf("estado=%d, se esperaba 500", respuesta.Code)
	}
}
