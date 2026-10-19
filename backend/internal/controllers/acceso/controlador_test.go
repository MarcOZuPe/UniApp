package acceso

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"uniapp/internal/adapters"
	"uniapp/internal/events"
	modelo "uniapp/internal/models/acceso"
)

const origen = "http://localhost:5173"
const clave = "Clave-Segura-2026!"

func servidor(t *testing.T) http.Handler {
	t.Helper()
	h := adapters.PBKDF2Hasher{Iteraciones: 1000}
	svc := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: modelo.NuevoRepositorioMemoria(),
		Fabrica:     modelo.NuevaFabrica(h, "uni.edu.pe"),
		Hasher:      h,
		Bus:         events.NuevoBus(),
	})
	if err := modelo.Sembrar(context.Background(), svc, modelo.OpcionesSiembra{DominioTutor: "uni.edu.pe", Clave: clave}); err != nil {
		t.Fatal(err)
	}
	c := Nuevo(svc, false)
	return Encadenar(c.Rutas(), OrigenPermitido(origen))
}

func llamar(h http.Handler, metodo, ruta, cuerpo string, cookie *http.Cookie, hdr ...string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	r.Header.Set("Content-Type", "application/json")
	for i := 0; i+1 < len(hdr); i += 2 {
		r.Header.Set(hdr[i], hdr[i+1])
	}
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func ingresar(t *testing.T, h http.Handler, email string) *http.Cookie {
	t.Helper()
	w := llamar(h, "POST", "/api/auth/ingreso", `{"email":"`+email+`","clave":"`+clave+`"}`, nil)
	if w.Code != 200 {
		t.Fatalf("ingreso %s: %d %s", email, w.Code, w.Body)
	}
	for _, c := range w.Result().Cookies() {
		if c.Name == nombreCookie {
			if !c.HttpOnly || c.SameSite != http.SameSiteStrictMode {
				t.Fatal("cookie debe ser HttpOnly y SameSite=Strict")
			}
			return c
		}
	}
	t.Fatal("sin cookie de sesión")
	return nil
}

func TestRegistroNoPermiteCoordinadorNiCampos(t *testing.T) {
	h := servidor(t)
	w := llamar(h, "POST", "/api/auth/registro", `{"email":"x@example.com","nombre_uso":"Xi","clave":"`+clave+`","rol":"coordinador"}`, nil)
	if w.Code != 403 {
		t.Fatalf("autoasignarse coordinador: %d", w.Code)
	}
	w = llamar(h, "POST", "/api/auth/registro", `{"email":"x@example.com","nombre_uso":"Xi","clave":"`+clave+`","rol":"escolar","admin":true}`, nil)
	if w.Code != 400 {
		t.Fatalf("campos desconocidos deben rechazarse: %d", w.Code)
	}
	w = llamar(h, "POST", "/api/auth/registro", `{"email":"x@example.com","nombre_uso":"Xi","clave":"corta","rol":"escolar"}`, nil)
	if w.Code != 422 || !strings.Contains(w.Body.String(), `"clave"`) {
		t.Fatalf("validación: %d %s", w.Code, w.Body)
	}
	w = llamar(h, "POST", "/api/auth/registro", `{"email":"nuevo@example.com","nombre_uso":"Nuevo","clave":"clave-larga-ok","rol":"escolar"}`, nil)
	if w.Code != 201 || strings.Contains(w.Body.String(), "clave") {
		t.Fatalf("registro válido no debe devolver la clave: %d %s", w.Code, w.Body)
	}
}

func TestSalidaInvalidaLaSesion(t *testing.T) {
	h := servidor(t)
	c := ingresar(t, h, "escolar.prueba@example.test")
	if w := llamar(h, "GET", "/api/auth/yo", "", c); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := llamar(h, "POST", "/api/auth/salida", "", c); w.Code != 204 {
		t.Fatal(w.Code)
	}
	if w := llamar(h, "GET", "/api/auth/yo", "", c); w.Code != 401 {
		t.Fatalf("tras salir: %d", w.Code)
	}
}

func TestOrigenCruzadoRechazado(t *testing.T) {
	h := servidor(t)
	w := llamar(h, "POST", "/api/auth/ingreso", `{"email":"a","clave":"b"}`, nil, "Origin", "https://malo.example")
	if w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
}

func TestRecuperarNoRevelaExistencia(t *testing.T) {
	h := servidor(t)
	a := llamar(h, "POST", "/api/auth/recuperar", `{"email":"escolar.prueba@example.test"}`, nil)
	b := llamar(h, "POST", "/api/auth/recuperar", `{"email":"no.existe@example.test"}`, nil)
	if a.Code != 202 || b.Code != 202 || a.Body.String() != b.Body.String() {
		t.Fatalf("respuestas distintas: %d/%d", a.Code, b.Code)
	}
}

func TestIngresoFallidoMensajeGenerico(t *testing.T) {
	h := servidor(t)
	a := llamar(h, "POST", "/api/auth/ingreso", `{"email":"escolar.prueba@example.test","clave":"mala"}`, nil)
	b := llamar(h, "POST", "/api/auth/ingreso", `{"email":"no.existe@example.test","clave":"mala"}`, nil)
	if a.Code != 401 || a.Body.String() != b.Body.String() {
		t.Fatal("el mensaje no debe distinguir correo inexistente de clave errónea")
	}
}
