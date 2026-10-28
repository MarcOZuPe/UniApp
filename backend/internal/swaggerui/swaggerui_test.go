package swaggerui

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const especificacion = `openapi: 3.0.3
info:
  title: Prueba
  version: 1.0.0
paths: {}
`

func TestRutasSirveUIYContrato(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := os.WriteFile(ruta, []byte(especificacion), 0o600); err != nil {
		t.Fatal(err)
	}

	manejador, err := Rutas(ruta)
	if err != nil {
		t.Fatalf("Rutas no debía fallar: %v", err)
	}

	ui := httptest.NewRecorder()
	manejador.ServeHTTP(ui, httptest.NewRequest(http.MethodGet, Prefijo+"/", nil))
	if ui.Code != http.StatusOK || !strings.Contains(ui.Body.String(), "Swagger UI") {
		t.Fatalf("Swagger UI: %d %s", ui.Code, ui.Body.String())
	}

	contrato := httptest.NewRecorder()
	manejador.ServeHTTP(contrato, httptest.NewRequest(http.MethodGet, Prefijo+"/swagger_spec", nil))
	if contrato.Code != http.StatusOK || contrato.Body.String() != especificacion {
		t.Fatalf("contrato: %d %s", contrato.Code, contrato.Body.String())
	}
	if tipo := contrato.Header().Get("Content-Type"); tipo != "application/yaml; charset=utf-8" {
		t.Fatalf("Content-Type del contrato: %q", tipo)
	}
}

func TestRutasFallaSinContrato(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "no-existe.yaml")
	if _, err := Rutas(ruta); err == nil {
		t.Fatal("Rutas debía fallar si el contrato no existe")
	}
	if _, err := Rutas(" "); err == nil {
		t.Fatal("Rutas debía rechazar una ruta vacía")
	}
}
