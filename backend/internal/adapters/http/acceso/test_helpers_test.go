package acceso

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
	modelo "uniapp/internal/application/acceso"
	infra "uniapp/internal/infrastructure"
)

const (
	origenPruebaHTTP         = "http://localhost:5173"
	clavePruebaHTTP          = "Clave-Segura-2026!"
	modeloEmailAdministrador = "admin@uni.edu.pe"
)

type entornoHTTP struct {
	manejador http.Handler
	servicio  *modelo.Servicio
	bus       *infra.Bus
}

func nuevoEntornoHTTP(t *testing.T) *entornoHTTP {
	t.Helper()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	bus := infra.NuevoBus()
	servicio := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: memoria.NuevoRepositorioMemoria(),
		Fabrica:     modelo.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         bus,
	})
	if err := modelo.Sembrar(context.Background(), servicio, modelo.DatosSiembra{
		DominioTutor: "uni.edu.pe", Clave: clavePruebaHTTP,
	}); err != nil {
		t.Fatal(err)
	}
	controlador := Nuevo(servicio, false)
	return &entornoHTTP{
		manejador: Encadenar(controlador.Rutas(), OrigenPermitido(origenPruebaHTTP)),
		servicio:  servicio,
		bus:       bus,
	}
}

func (entorno *entornoHTTP) solicitar(metodo, ruta, cuerpo string, cookie *http.Cookie) *httptest.ResponseRecorder {
	entornoAcceso := entorno
	peticion := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.Header.Set("Origin", origenPruebaHTTP)
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	respuesta := httptest.NewRecorder()
	entornoAcceso.manejador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func (entorno *entornoHTTP) ingresar(t *testing.T, email string) *http.Cookie {
	t.Helper()
	respuesta := entorno.solicitar(http.MethodPost, "/api/auth/ingreso",
		`{"email":"`+email+`","clave":"`+clavePruebaHTTP+`"}`, nil)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("ingreso %s: %d %s", email, respuesta.Code, respuesta.Body)
	}
	for _, cookie := range respuesta.Result().Cookies() {
		if cookie.Name == nombreCookie {
			return cookie
		}
	}
	t.Fatal("la respuesta de ingreso no incluye la cookie de sesión")
	return nil
}

func cuerpoJSON(t *testing.T, respuesta *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var datos map[string]any
	if err := json.Unmarshal(respuesta.Body.Bytes(), &datos); err != nil {
		t.Fatalf("respuesta no es JSON (%d): %s", respuesta.Code, respuesta.Body)
	}
	return datos
}

func idDeEmail(t *testing.T, servicio *modelo.Servicio, email string) string {
	t.Helper()
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorEmail(context.Background(), email)
	if err != nil {
		t.Fatal(err)
	}
	return usuario.ID
}
