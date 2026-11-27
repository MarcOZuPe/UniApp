package swaggerui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRutasSwagger(t *testing.T) {
	if _, err := Rutas(""); err == nil {
		t.Fatal("sin ruta de contrato debe fallar")
	}
	if _, err := Rutas(filepath.Join(t.TempDir(), "inexistente.yaml")); err == nil {
		t.Fatal("un contrato inexistente debe fallar")
	}
	ruta := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := os.WriteFile(ruta, []byte("openapi: 3.0.3\n"), 0600); err != nil {
		t.Fatal(err)
	}
	manejador, err := Rutas(ruta)
	if err != nil {
		t.Fatalf("crear rutas falló: %v", err)
	}

	redirect := httptest.NewRecorder()
	manejador.ServeHTTP(redirect, httptest.NewRequest(http.MethodGet, "/swagger?x=1", nil))
	if redirect.Code != http.StatusMovedPermanently || redirect.Header().Get("Location") != "/swagger/?x=1" {
		t.Fatalf("redirección inesperada: %d %q", redirect.Code, redirect.Header().Get("Location"))
	}

	inicializador := httptest.NewRecorder()
	manejador.ServeHTTP(inicializador, httptest.NewRequest(http.MethodGet, "/swagger/swagger-initializer.js", nil))
	if inicializador.Code != http.StatusOK || inicializador.Header().Get("Content-Type") != "application/javascript; charset=utf-8" {
		t.Fatalf("inicializador inesperado: %d %q", inicializador.Code, inicializador.Header().Get("Content-Type"))
	}
	if inicializador.Body.Len() == 0 {
		t.Fatal("el inicializador no debe estar vacío")
	}

	especificacion := httptest.NewRecorder()
	manejador.ServeHTTP(especificacion, httptest.NewRequest(http.MethodGet, "/swagger/swagger_spec", nil))
	if especificacion.Code != http.StatusOK || especificacion.Header().Get("Content-Type") != "application/yaml; charset=utf-8" {
		t.Fatalf("especificación inesperada: %d %q", especificacion.Code, especificacion.Header().Get("Content-Type"))
	}
}
