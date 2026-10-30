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

func TestRutaSalud(testingT *testing.T) {
	manejador := nuevoManejadorPrueba()

	respuesta := ejecutarPeticion(opcionesPeticionSesion{
		manejador: manejador,
		metodo:    http.MethodGet,
		ruta:      "/api/sesion/salud",
	})
	if respuesta.Code != http.StatusOK {
		testingT.Fatalf("se esperaba 200, se obtuvo %d", respuesta.Code)
	}

	var cuerpo struct {
		Flujo string `json:"flujo"`
	}
	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil {
		testingT.Fatal(err)
	}
	if cuerpo.Flujo != "sesion" {
		testingT.Fatalf("flujo inesperado: %s", cuerpo.Flujo)
	}
}

func TestConsultarSinUsuario(testingT *testing.T) {
	manejador := nuevoManejadorPrueba()

	respuesta := ejecutarPeticion(opcionesPeticionSesion{
		manejador: manejador,
		metodo:    http.MethodGet,
		ruta:      "/api/sesion/sesion-001",
	})
	if respuesta.Code != http.StatusUnauthorized {
		testingT.Fatalf("se esperaba 401, se obtuvo %d", respuesta.Code)
	}
}

func TestDeclararSinUsuario(testingT *testing.T) {
	manejador := nuevoManejadorPrueba()

	respuesta := ejecutarPeticion(opcionesPeticionSesion{
		manejador: manejador,
		metodo:    http.MethodPost,
		ruta:      "/api/sesion/sesion-001/asistencias",
		cabeceras: map[string]string{"Origin": origenPrueba},
	})
	if respuesta.Code != http.StatusUnauthorized {
		testingT.Fatalf("se esperaba 401, se obtuvo %d", respuesta.Code)
	}
}

type opcionesPeticionSesion struct {
	manejador http.Handler
	metodo    string
	ruta      string
	cookie    *http.Cookie
	cabeceras map[string]string
}

func ejecutarPeticion(opciones opcionesPeticionSesion) *httptest.ResponseRecorder {
	peticion := httptest.NewRequest(opciones.metodo, opciones.ruta, nil)
	if opciones.cookie != nil {
		peticion.AddCookie(opciones.cookie)
	}
	for nombre, valor := range opciones.cabeceras {
		peticion.Header.Set(nombre, valor)
	}

	respuesta := httptest.NewRecorder()
	opciones.manejador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func autenticarPrueba(siguiente http.Handler) http.Handler {
	return siguiente
}

func nuevoManejadorPrueba() http.Handler {
	var repositorio modelo.Repositorio = memoria.NuevoRepositorioMemoria()
	return Rutas(repositorio, autenticarPrueba, origenPrueba)
}
