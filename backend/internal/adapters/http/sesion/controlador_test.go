package sesion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	memoria "uniapp/internal/adapters/repositories/memory/sesion"

	modelo "uniapp/internal/application/sesion"
)

const origenPrueba = "http://localhost:5173"

// Simula un middleware que no incorpora usuario al contexto.
// Permite comprobar que el controlador rechaza el acceso.
func autenticarPrueba(siguiente http.Handler) http.Handler {
	return siguiente
}

func nuevoManejadorPrueba() http.Handler {
	var repositorio modelo.Repositorio = memoria.NuevoRepositorioMemoria()

	return Rutas(
		repositorio,
		autenticarPrueba,
		origenPrueba,
	)
}

func TestRutaSalud(t *testing.T) {
	manejador := nuevoManejadorPrueba()

	peticion := httptest.NewRequest(
		http.MethodGet,
		"/api/sesion/salud",
		nil,
	)
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba 200, se obtuvo %d",
			respuesta.Code,
		)
	}

	var cuerpo struct {
		Flujo string `json:"flujo"`
	}

	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil {
		t.Fatal(err)
	}

	if cuerpo.Flujo != "sesion" {
		t.Fatalf("flujo inesperado: %s", cuerpo.Flujo)
	}
}

func TestConsultarSinUsuario(t *testing.T) {
	manejador := nuevoManejadorPrueba()

	peticion := httptest.NewRequest(
		http.MethodGet,
		"/api/sesion/sesion-001",
		nil,
	)
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf(
			"se esperaba 401, se obtuvo %d",
			respuesta.Code,
		)
	}
}

func TestDeclararSinUsuario(t *testing.T) {
	manejador := nuevoManejadorPrueba()

	peticion := httptest.NewRequest(
		http.MethodPost,
		"/api/sesion/sesion-001/asistencias",
		nil,
	)
	peticion.Header.Set("Origin", origenPrueba)

	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf(
			"se esperaba 401, se obtuvo %d",
			respuesta.Code,
		)
	}
}
