package acceso

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
	modelo "uniapp/internal/application/acceso"
	infra "uniapp/internal/infrastructure"
)

const origen = "http://localhost:5173"
const clave = "Clave-Segura-2026!"

func TestRegistroNoPermiteCoordinadorNiCampos(t *testing.T) {
	entorno := inicializarServidorPruebas(t)

	respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/registro",
		cuerpo:    `{"email":"x@example.com","nombre_uso":"Xi","clave":"` + clave + `","rol":"coordinador"}`,
	})
	if respuesta.Code != http.StatusForbidden {
		t.Fatalf("autoasignarse coordinador: %d", respuesta.Code)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/registro",
		cuerpo:    `{"email":"x@example.com","nombre_uso":"Xi","clave":"` + clave + `","rol":"escolar","admin":true}`,
	})
	if respuesta.Code != http.StatusBadRequest {
		t.Fatalf("campos desconocidos deben rechazarse: %d", respuesta.Code)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/registro",
		cuerpo:    `{"email":"x@example.com","nombre_uso":"Xi","clave":"corta","rol":"escolar"}`,
	})
	if respuesta.Code != http.StatusUnprocessableEntity || !strings.Contains(respuesta.Body.String(), `"clave"`) {
		t.Fatalf("validación: %d %s", respuesta.Code, respuesta.Body)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/registro",
		cuerpo:    `{"email":"nuevo@example.com","nombre_uso":"Nuevo","clave":"clave-larga-ok","rol":"escolar"}`,
	})
	if respuesta.Code != http.StatusCreated || strings.Contains(respuesta.Body.String(), "clave") {
		t.Fatalf("registro válido no debe devolver la clave: %d %s", respuesta.Code, respuesta.Body)
	}
}

func TestRegistroConfirmaCorreoAntesDeIngresar(t *testing.T) {
	entorno := inicializarServidorPruebas(t)

	var token string
	entorno.bus.Registrar(modelo.ConfirmacionSolicitada, infra.NuevoObservadorFunc("captura-confirmacion", func(evento modelo.Event) {
		token, _ = evento.Payload["token"].(string)
	}))

	cuerpo := `{"email":"nuevo@example.com","nombre_uso":"Nuevo","clave":"` + clave + `","rol":"escolar"}`
	respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/registro",
		cuerpo:    cuerpo,
	})
	if respuesta.Code != http.StatusCreated || !strings.Contains(respuesta.Body.String(), `"correo_confirmado":false`) {
		t.Fatalf("registro: %d %s", respuesta.Code, respuesta.Body)
	}
	if token == "" {
		t.Fatal("debía emitirse un enlace de confirmación")
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/registro",
		cuerpo:    cuerpo,
	})
	if respuesta.Code != http.StatusConflict {
		t.Fatalf("un correo admite un solo registro: %d %s", respuesta.Code, respuesta.Body)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/ingreso",
		cuerpo:    `{"email":"nuevo@example.com","clave":"` + clave + `"}`,
	})
	if respuesta.Code != http.StatusUnauthorized || !strings.Contains(respuesta.Body.String(), "credenciales") {
		t.Fatalf("sin confirmar debe ser invisible (401 genérico): %d %s", respuesta.Code, respuesta.Body)
	}

	respuestaExistente := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/reenviar-confirmacion",
		cuerpo:    `{"email":"nuevo@example.com"}`,
	})
	respuestaInexistente := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/reenviar-confirmacion",
		cuerpo:    `{"email":"no.existe@example.test"}`,
	})
	if respuestaExistente.Code != http.StatusAccepted || respuestaInexistente.Code != http.StatusAccepted ||
		respuestaExistente.Body.String() != respuestaInexistente.Body.String() {
		t.Fatalf("reenviar no debe revelar existencia: %d/%d %s/%s",
			respuestaExistente.Code, respuestaInexistente.Code, respuestaExistente.Body, respuestaInexistente.Body)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/confirmar",
		cuerpo:    `{"token":"` + token + `"}`,
	})
	if respuesta.Code != http.StatusOK {
		t.Fatalf("confirmar: %d %s", respuesta.Code, respuesta.Body)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/confirmar",
		cuerpo:    `{"token":"` + token + `"}`,
	})
	if respuesta.Code != http.StatusBadRequest {
		t.Fatalf("el enlace es de un solo uso: %d %s", respuesta.Code, respuesta.Body)
	}

	_ = ingresar(t, entorno.manejador, "nuevo@example.com")
}

func TestSalidaInvalidaLaSesion(t *testing.T) {
	entorno := inicializarServidorPruebas(t)
	cookie := ingresar(t, entorno.manejador, "escolar.prueba@example.test")

	if respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "GET",
		ruta:      "/api/auth/yo",
		cookie:    cookie,
	}); respuesta.Code != http.StatusOK {
		t.Fatal(respuesta.Code)
	}

	if respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/salida",
		cookie:    cookie,
	}); respuesta.Code != http.StatusNoContent {
		t.Fatal(respuesta.Code)
	}

	if respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "GET",
		ruta:      "/api/auth/yo",
		cookie:    cookie,
	}); respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("tras salir: %d", respuesta.Code)
	}
}

func TestOrigenCruzadoRechazado(t *testing.T) {
	entorno := inicializarServidorPruebas(t)

	respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/ingreso",
		cuerpo:    `{"email":"a","clave":"b"}`,
		cabeceras: map[string]string{"Origin": "https://malo.example"},
	})
	if respuesta.Code != http.StatusForbidden {
		t.Fatalf("CSRF: %d", respuesta.Code)
	}
}

func TestOrigenPermitidoAceptaVarios(t *testing.T) {
	var manejador http.Handler = http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})
	manejador = Encadenar(manejador, OrigenPermitido("http://localhost:5173", "http://203.0.113.10:5173"))

	for _, origenPermitido := range []string{"http://localhost:5173", "http://203.0.113.10:5173"} {
		respuesta := ejecutarPeticion(opcionesPeticion{
			manejador: manejador,
			metodo:    "POST",
			ruta:      "/x",
			cabeceras: map[string]string{"Origin": origenPermitido},
		})
		if respuesta.Code != http.StatusNoContent {
			t.Fatalf("origen permitido %s: %d", origenPermitido, respuesta.Code)
		}
	}

	respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: manejador,
		metodo:    "POST",
		ruta:      "/x",
		cabeceras: map[string]string{"Origin": "http://otro.example"},
	})
	if respuesta.Code != http.StatusForbidden {
		t.Fatalf("origen ajeno: %d", respuesta.Code)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: manejador,
		metodo:    "GET",
		ruta:      "/x",
		cabeceras: map[string]string{"Origin": "http://otro.example"},
	})
	if respuesta.Code != http.StatusNoContent {
		t.Fatalf("las lecturas no comprueban origen: %d", respuesta.Code)
	}
}

func TestRecuperarNoRevelaExistencia(t *testing.T) {
	entorno := inicializarServidorPruebas(t)

	respuestaExistente := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/recuperar",
		cuerpo:    `{"email":"escolar.prueba@example.test"}`,
	})
	respuestaInexistente := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/recuperar",
		cuerpo:    `{"email":"no.existe@example.test"}`,
	})
	if respuestaExistente.Code != http.StatusAccepted || respuestaInexistente.Code != http.StatusAccepted ||
		respuestaExistente.Body.String() != respuestaInexistente.Body.String() {
		t.Fatalf("respuestas distintas: %d/%d", respuestaExistente.Code, respuestaInexistente.Code)
	}
}

func TestIngresoFallidoMensajeGenerico(t *testing.T) {
	entorno := inicializarServidorPruebas(t)

	respuestaExistente := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/ingreso",
		cuerpo:    `{"email":"escolar.prueba@example.test","clave":"mala"}`,
	})
	respuestaInexistente := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "POST",
		ruta:      "/api/auth/ingreso",
		cuerpo:    `{"email":"no.existe@example.test","clave":"mala"}`,
	})
	if respuestaExistente.Code != http.StatusUnauthorized ||
		respuestaExistente.Body.String() != respuestaInexistente.Body.String() {
		t.Fatal("el mensaje no debe distinguir correo inexistente de clave errónea")
	}
}

func TestPA01PanelCoordinacionSoloCoordinador(t *testing.T) {
	entorno := inicializarServidorPruebas(t)

	cookieEscolar := ingresar(t, entorno.manejador, "escolar.prueba@example.test")
	respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "GET",
		ruta:      "/api/coordinacion/panel",
		cookie:    cookieEscolar,
	})
	if respuesta.Code != http.StatusForbidden {
		t.Fatalf("PA-01: escolar debe recibir 403, obtuvo %d %s", respuesta.Code, respuesta.Body)
	}

	cookieCoordinador := ingresar(t, entorno.manejador, "coord.prueba@example.test")
	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "GET",
		ruta:      "/api/coordinacion/panel",
		cookie:    cookieCoordinador,
	})
	if respuesta.Code != http.StatusOK || !strings.Contains(respuesta.Body.String(), `"rol":"coordinador"`) {
		t.Fatalf("PA-01: coordinador debe recibir 200, obtuvo %d %s", respuesta.Code, respuesta.Body)
	}

	respuesta = ejecutarPeticion(opcionesPeticion{
		manejador: entorno.manejador,
		metodo:    "GET",
		ruta:      "/api/coordinacion/panel",
	})
	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("PA-01: sin sesión debe recibir 401, obtuvo %d %s", respuesta.Code, respuesta.Body)
	}
}

type entornoPruebas struct {
	manejador http.Handler
	bus       *infra.Bus
}

func inicializarServidorPruebas(t *testing.T) entornoPruebas {
	t.Helper()

	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	busDeEventos := infra.NuevoBus()
	servicio := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: memoria.NuevoRepositorioMemoria(),
		Fabrica:     modelo.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         busDeEventos,
	})

	if err := modelo.Sembrar(context.Background(), servicio, modelo.DatosSiembra{
		DominioTutor: "uni.edu.pe",
		Clave:        clave,
	}); err != nil {
		t.Fatal(err)
	}

	controlador := Nuevo(servicio, false)
	return entornoPruebas{
		manejador: Encadenar(controlador.Rutas(), OrigenPermitido(origen)),
		bus:       busDeEventos,
	}
}

type opcionesPeticion struct {
	manejador http.Handler
	metodo    string
	ruta      string
	cuerpo    string
	cookie    *http.Cookie
	cabeceras map[string]string
}

func ejecutarPeticion(opciones opcionesPeticion) *httptest.ResponseRecorder {
	peticion := httptest.NewRequest(opciones.metodo, opciones.ruta, strings.NewReader(opciones.cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	for nombre, valor := range opciones.cabeceras {
		peticion.Header.Set(nombre, valor)
	}
	if opciones.cookie != nil {
		peticion.AddCookie(opciones.cookie)
	}

	respuesta := httptest.NewRecorder()
	opciones.manejador.ServeHTTP(respuesta, peticion)
	return respuesta
}

func ingresar(t *testing.T, manejador http.Handler, email string) *http.Cookie {
	t.Helper()

	respuesta := ejecutarPeticion(opcionesPeticion{
		manejador: manejador,
		metodo:    "POST",
		ruta:      "/api/auth/ingreso",
		cuerpo:    `{"email":"` + email + `","clave":"` + clave + `"}`,
	})
	if respuesta.Code != http.StatusOK {
		t.Fatalf("ingreso %s: %d %s", email, respuesta.Code, respuesta.Body)
	}

	for _, cookie := range respuesta.Result().Cookies() {
		if cookie.Name == nombreCookie {
			if !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
				t.Fatal("cookie debe ser HttpOnly y SameSite=Strict")
			}
			return cookie
		}
	}
	t.Fatal("sin cookie de sesión")
	return nil
}
