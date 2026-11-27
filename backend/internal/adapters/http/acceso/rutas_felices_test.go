package acceso

import (
	"net/http"
	"testing"
)

func TestRutasDeAdministracionYCoordinacion(t *testing.T) {
	entorno := nuevoEntornoHTTP(t)
	admin := entorno.ingresar(t, modeloEmailAdministrador)

	if respuesta := entorno.solicitar(http.MethodGet, "/api/facultades", "", nil); respuesta.Code != http.StatusOK {
		t.Fatalf("facultades: %d %s", respuesta.Code, respuesta.Body)
	}
	if respuesta := entorno.solicitar(http.MethodGet, "/api/auth/yo", "", admin); respuesta.Code != http.StatusOK {
		t.Fatalf("yo: %d %s", respuesta.Code, respuesta.Body)
	}
	if respuesta := entorno.solicitar(http.MethodGet, "/api/administracion/coordinadores-generales", "", admin); respuesta.Code != http.StatusOK {
		t.Fatalf("listar generales: %d %s", respuesta.Code, respuesta.Body)
	}
	creado := entorno.solicitar(http.MethodPost, "/api/administracion/coordinadores-generales",
		`{"email":"nuevo.general@example.test","nombre_uso":"Nuevo General","clave":"Clave-Segura-2026!","facultad_id":"1"}`, admin)
	if creado.Code != http.StatusCreated {
		t.Fatalf("crear general: %d %s", creado.Code, creado.Body)
	}
	idGeneral, _ := cuerpoJSON(t, creado)["id"].(string)
	if idGeneral == "" {
		t.Fatal("la respuesta de creación no incluye id")
	}
	if respuesta := entorno.solicitar(http.MethodGet, "/api/administracion/logs", "", admin); respuesta.Code != http.StatusOK {
		t.Fatalf("logs: %d %s", respuesta.Code, respuesta.Body)
	}
	if respuesta := entorno.solicitar(http.MethodDelete, "/api/administracion/coordinadores-generales/"+idGeneral, "", admin); respuesta.Code != http.StatusNoContent {
		t.Fatalf("eliminar general: %d %s", respuesta.Code, respuesta.Body)
	}

	general := entorno.ingresar(t, "coord.general.prueba@example.test")
	if respuesta := entorno.solicitar(http.MethodGet, "/api/coordinacion/panel", "", general); respuesta.Code != http.StatusOK {
		t.Fatalf("panel: %d %s", respuesta.Code, respuesta.Body)
	}
	if respuesta := entorno.solicitar(http.MethodGet, "/api/coordinacion/coordinadores", "", general); respuesta.Code != http.StatusOK {
		t.Fatalf("listar coordinadores: %d %s", respuesta.Code, respuesta.Body)
	}
	coordinador := entorno.solicitar(http.MethodPost, "/api/coordinacion/coordinadores",
		`{"email":"nuevo.coordinador@example.test","nombre_uso":"Nuevo Coordinador","clave":"Clave-Segura-2026!"}`, general)
	if coordinador.Code != http.StatusCreated {
		t.Fatalf("crear coordinador: %d %s", coordinador.Code, coordinador.Body)
	}
	idCoordinador, _ := cuerpoJSON(t, coordinador)["id"].(string)
	if respuesta := entorno.solicitar(http.MethodDelete, "/api/coordinacion/coordinadores/"+idCoordinador, "", general); respuesta.Code != http.StatusNoContent {
		t.Fatalf("eliminar coordinador: %d %s", respuesta.Code, respuesta.Body)
	}
}

func TestRutasDeApoderadoYPerfil(t *testing.T) {
	entorno := nuevoEntornoHTTP(t)
	apoderado := entorno.ingresar(t, "apoderado.prueba@example.test")
	if respuesta := entorno.solicitar(http.MethodGet, "/api/apoderado/vinculaciones", "", apoderado); respuesta.Code != http.StatusOK {
		t.Fatalf("vinculaciones: %d %s", respuesta.Code, respuesta.Body)
	}
	escolar := entorno.ingresar(t, "escolar.prueba@example.test")
	if respuesta := entorno.solicitar(http.MethodGet, "/api/escolar/apoderado", "", escolar); respuesta.Code != http.StatusOK {
		t.Fatalf("mi apoderado: %d %s", respuesta.Code, respuesta.Body)
	}
	cambio := entorno.solicitar(http.MethodPut, "/api/perfil/clave",
		`{"clave_actual":"`+clavePruebaHTTP+`","clave_nueva":"Otra-Clave-2026!"}`, escolar)
	if cambio.Code != http.StatusNoContent {
		t.Fatalf("cambio de clave: %d %s", cambio.Code, cambio.Body)
	}
	if respuesta := entorno.solicitar(http.MethodGet, "/api/auth/yo", "", escolar); respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("el token anterior debe invalidarse: %d %s", respuesta.Code, respuesta.Body)
	}
	if respuesta := entorno.solicitar(http.MethodPost, "/api/auth/salida", "", apoderado); respuesta.Code != http.StatusNoContent {
		t.Fatalf("salida: %d %s", respuesta.Code, respuesta.Body)
	}
	if respuesta := entorno.solicitar(http.MethodPost, "/api/auth/recuperar", `{"email":"escolar.prueba@example.test"}`, nil); respuesta.Code != http.StatusAccepted {
		t.Fatalf("recuperar: %d %s", respuesta.Code, respuesta.Body)
	}
}
