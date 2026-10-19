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

func servidor(t *testing.T) (http.Handler, *events.Bus) {
	t.Helper()
	h := adapters.PBKDF2Hasher{Iteraciones: 1000}
	bus := events.NuevoBus()
	svc := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: modelo.NuevoRepositorioMemoria(),
		Fabrica:     modelo.NuevaFabrica(h, "uni.edu.pe"),
		Hasher:      h,
		Bus:         bus,
	})
	if err := modelo.Sembrar(context.Background(), svc, modelo.OpcionesSiembra{DominioTutor: "uni.edu.pe", Clave: clave}); err != nil {
		t.Fatal(err)
	}
	c := Nuevo(svc, false)
	return Encadenar(c.Rutas(), OrigenPermitido(origen)), bus
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
	h, _ := servidor(t)
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

func TestRegistroConfirmaCorreoAntesDeIngresar(t *testing.T) {
	h, bus := servidor(t)
	var token string
	bus.Registrar(events.ConfirmacionSolicitada, events.NuevoObservadorFunc("captura-confirmacion", func(e events.Event) {
		token, _ = e.Payload["token"].(string)
	}))
	cuerpo := `{"email":"nuevo@example.com","nombre_uso":"Nuevo","clave":"` + clave + `","rol":"escolar"}`
	w := llamar(h, "POST", "/api/auth/registro", cuerpo, nil)
	if w.Code != 201 || !strings.Contains(w.Body.String(), `"correo_confirmado":false`) {
		t.Fatalf("registro: %d %s", w.Code, w.Body)
	}
	if token == "" {
		t.Fatal("debía emitirse un enlace de confirmación")
	}
	if w = llamar(h, "POST", "/api/auth/registro", cuerpo, nil); w.Code != 409 {
		t.Fatalf("un correo admite un solo registro: %d %s", w.Code, w.Body)
	}
	if w = llamar(h, "POST", "/api/auth/ingreso", `{"email":"nuevo@example.com","clave":"`+clave+`"}`, nil); w.Code != 401 || !strings.Contains(w.Body.String(), "credenciales") {
		t.Fatalf("sin confirmar debe ser invisible (401 genérico): %d %s", w.Code, w.Body)
	}
	a := llamar(h, "POST", "/api/auth/reenviar-confirmacion", `{"email":"nuevo@example.com"}`, nil)
	b := llamar(h, "POST", "/api/auth/reenviar-confirmacion", `{"email":"no.existe@example.test"}`, nil)
	if a.Code != 202 || b.Code != 202 || a.Body.String() != b.Body.String() {
		t.Fatalf("reenviar no debe revelar existencia: %d/%d %s/%s", a.Code, b.Code, a.Body, b.Body)
	}
	if w = llamar(h, "POST", "/api/auth/confirmar", `{"token":"`+token+`"}`, nil); w.Code != 200 {
		t.Fatalf("confirmar: %d %s", w.Code, w.Body)
	}
	if w = llamar(h, "POST", "/api/auth/confirmar", `{"token":"`+token+`"}`, nil); w.Code != 400 {
		t.Fatalf("el enlace es de un solo uso: %d %s", w.Code, w.Body)
	}
	_ = ingresar(t, h, "nuevo@example.com")
}

func TestSalidaInvalidaLaSesion(t *testing.T) {
	h, _ := servidor(t)
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
	h, _ := servidor(t)
	w := llamar(h, "POST", "/api/auth/ingreso", `{"email":"a","clave":"b"}`, nil, "Origin", "https://malo.example")
	if w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
}

func TestOrigenPermitidoAceptaVarios(t *testing.T) {
	base := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) })
	h := Encadenar(base, OrigenPermitido("http://localhost:5173", "http://159.195.245.124:5173"))
	for _, origen := range []string{"http://localhost:5173", "http://159.195.245.124:5173"} {
		if w := llamar(h, "POST", "/x", "", nil, "Origin", origen); w.Code != http.StatusNoContent {
			t.Fatalf("origen permitido %s: %d", origen, w.Code)
		}
	}
	if w := llamar(h, "POST", "/x", "", nil, "Origin", "http://otro.example"); w.Code != http.StatusForbidden {
		t.Fatalf("origen ajeno: %d", w.Code)
	}
	if w := llamar(h, "GET", "/x", "", nil, "Origin", "http://otro.example"); w.Code != http.StatusNoContent {
		t.Fatalf("las lecturas no comprueban origen: %d", w.Code)
	}
}

func TestRecuperarNoRevelaExistencia(t *testing.T) {
	h, _ := servidor(t)
	a := llamar(h, "POST", "/api/auth/recuperar", `{"email":"escolar.prueba@example.test"}`, nil)
	b := llamar(h, "POST", "/api/auth/recuperar", `{"email":"no.existe@example.test"}`, nil)
	if a.Code != 202 || b.Code != 202 || a.Body.String() != b.Body.String() {
		t.Fatalf("respuestas distintas: %d/%d", a.Code, b.Code)
	}
}

func TestIngresoFallidoMensajeGenerico(t *testing.T) {
	h, _ := servidor(t)
	a := llamar(h, "POST", "/api/auth/ingreso", `{"email":"escolar.prueba@example.test","clave":"mala"}`, nil)
	b := llamar(h, "POST", "/api/auth/ingreso", `{"email":"no.existe@example.test","clave":"mala"}`, nil)
	if a.Code != 401 || a.Body.String() != b.Body.String() {
		t.Fatal("el mensaje no debe distinguir correo inexistente de clave errónea")
	}
}
