package acceso

import (
	"net/http"
	"strings"
	"testing"

	modelo "uniapp/internal/models/acceso"
)

func TestVinculacionApoderadoYEscolar(t *testing.T) {
	h, bus := servidor(t)
	var token string
	bus.Registrar(modelo.VinculacionSolicitada, modelo.NuevoObservadorFunc("captura-vinculacion-http", func(e modelo.Event) {
		token, _ = e.Payload["token"].(string)
	}))

	apoderado := ingresar(t, h, "apoderado.prueba@example.test")
	escolar := ingresar(t, h, "escolar.prueba@example.test")

	w := llamar(h, "POST", "/api/apoderado/vinculaciones", `{"email_escolar":"escolar.prueba@example.test"}`, apoderado)
	if w.Code != http.StatusAccepted || !strings.Contains(w.Body.String(), `"estado":"pendiente"`) {
		t.Fatalf("solicitud: %d %s", w.Code, w.Body)
	}
	if token == "" {
		t.Fatal("debía emitirse el token de vinculación")
	}

	w = llamar(h, "POST", "/api/escolar/vinculaciones/consultar", `{"token":"`+token+`"}`, escolar)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Apoderado Prueba") {
		t.Fatalf("consulta: %d %s", w.Code, w.Body)
	}

	w = llamar(h, "POST", "/api/escolar/vinculaciones/aceptar", `{"token":"`+token+`"}`, escolar)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"estado":"aceptada"`) {
		t.Fatalf("aceptar: %d %s", w.Code, w.Body)
	}

	w = llamar(h, "POST", "/api/escolar/vinculaciones/aceptar", `{"token":"`+token+`"}`, escolar)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("token de un solo uso: %d %s", w.Code, w.Body)
	}

	w = llamar(h, "GET", "/api/escolar/apoderado", "", escolar)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "apoderado.prueba@example.test") {
		t.Fatalf("mi apoderado: %d %s", w.Code, w.Body)
	}

	w = llamar(h, "GET", "/api/apoderado/vinculaciones", "", apoderado)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "escolar.prueba@example.test") {
		t.Fatalf("listado: %d %s", w.Code, w.Body)
	}
}

func TestVinculacionRechazaRolesYFaltasDeSesion(t *testing.T) {
	h, _ := servidor(t)
	apoderado := ingresar(t, h, "apoderado.prueba@example.test")
	escolar := ingresar(t, h, "escolar.prueba@example.test")

	if w := llamar(h, "POST", "/api/apoderado/vinculaciones", `{"email_escolar":"escolar.prueba@example.test"}`, escolar); w.Code != http.StatusForbidden {
		t.Fatalf("escolar no solicita: %d %s", w.Code, w.Body)
	}
	if w := llamar(h, "POST", "/api/escolar/vinculaciones/aceptar", `{"token":"x"}`, apoderado); w.Code != http.StatusForbidden {
		t.Fatalf("apoderado no acepta: %d %s", w.Code, w.Body)
	}
	if w := llamar(h, "GET", "/api/apoderado/vinculaciones", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d %s", w.Code, w.Body)
	}
	if w := llamar(h, "POST", "/api/apoderado/vinculaciones", `{"email_escolar":"tutor.prueba@uni.edu.pe"}`, apoderado); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("destino que no es escolar: %d %s", w.Code, w.Body)
	}
}
