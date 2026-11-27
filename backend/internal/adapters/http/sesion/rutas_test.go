package sesion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoriaSesion "uniapp/internal/adapters/repositories/memory/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/sesion"
	infra "uniapp/internal/infrastructure"
)

func entornoSesion(t *testing.T) (http.Handler, *http.Cookie, *http.Cookie, string) {
	t.Helper()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	acceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: memoriaAcceso.NuevoRepositorioMemoria(),
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(context.Background(), acceso, modeloAcceso.DatosSiembra{
		DominioTutor: "uni.edu.pe", Clave: "Clave-Segura-2026!",
	}); err != nil {
		t.Fatal(err)
	}
	tutor, err := acceso.RepositorioUsuarios.UsuarioPorEmail(context.Background(), "tutor.prueba@uni.edu.pe")
	if err != nil {
		t.Fatal(err)
	}
	escolar, err := acceso.RepositorioUsuarios.UsuarioPorEmail(context.Background(), "escolar.prueba@example.test")
	if err != nil {
		t.Fatal(err)
	}
	repo := memoriaSesion.NuevoRepositorioMemoria()
	inicio := time.Now().UTC().Add(-30 * time.Minute)
	sesion, err := modelo.NuevaSesion("s-1", "r-1", tutor.ID, escolar.ID, inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.CrearSesion(context.Background(), sesion); err != nil {
		t.Fatal(err)
	}
	tokenTutor, _, err := acceso.Sesion.Ingresar(context.Background(), modeloAcceso.CredencialesIngreso{
		Email: "tutor.prueba@uni.edu.pe", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	tokenEscolar, _, err := acceso.Sesion.Ingresar(context.Background(), modeloAcceso.CredencialesIngreso{
		Email: "escolar.prueba@example.test", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	manejador := Rutas(repo, ctrlAcceso.Nuevo(acceso, false).Autenticar(), "http://localhost:5173")
	return manejador, &http.Cookie{Name: "uniapp_sesion", Value: tokenTutor}, &http.Cookie{Name: "uniapp_sesion", Value: tokenEscolar}, escolar.ID
}

func TestRutasSesionYAsistencia(t *testing.T) {
	manejador, tutor, escolar, _ := entornoSesion(t)
	peticion := httptest.NewRequest(http.MethodGet, "/api/sesion/salud", nil)
	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("salud: %d", respuesta.Code)
	}

	peticion = httptest.NewRequest(http.MethodGet, "/api/sesion/s-1", nil)
	peticion.AddCookie(escolar)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("detalle: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/sesion/por-reserva/r-1", nil)
	peticion.AddCookie(tutor)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("por reserva: %d %s", respuesta.Code, respuesta.Body)
	}

	peticion = httptest.NewRequest(http.MethodPost, "/api/sesion/s-1/asistencias", nil)
	peticion.AddCookie(escolar)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusCreated {
		t.Fatalf("asistencia: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodPost, "/api/sesion/s-1/asistencias", nil)
	peticion.AddCookie(escolar)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusConflict {
		t.Fatalf("asistencia duplicada: %d %s", respuesta.Code, respuesta.Body)
	}

	peticion = httptest.NewRequest(http.MethodPost, "/api/sesion/s-1/ausencias", nil)
	peticion.AddCookie(tutor)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusCreated {
		t.Fatalf("ausencia: %d %s", respuesta.Code, respuesta.Body)
	}

	cuerpo := `{"resumen":"Resultado de la tutoría"}`
	peticion = httptest.NewRequest(http.MethodPost, "/api/sesion/s-1/resultado", stringsReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(tutor)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusCreated {
		t.Fatalf("resultado: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodPost, "/api/sesion/s-1/resultado", stringsReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(tutor)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusConflict {
		t.Fatalf("resultado duplicado: %d %s", respuesta.Code, respuesta.Body)
	}

	peticion = httptest.NewRequest(http.MethodGet, "/api/sesion/s-1", nil)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", respuesta.Code)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/sesion/inexistente", nil)
	peticion.AddCookie(tutor)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusNotFound {
		t.Fatalf("inexistente: %d", respuesta.Code)
	}
}

func stringsReader(valor string) *strings.Reader { return strings.NewReader(valor) }
