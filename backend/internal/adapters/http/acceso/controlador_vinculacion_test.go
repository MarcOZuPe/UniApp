package acceso

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modelo "uniapp/internal/application/acceso"
	infra "uniapp/internal/infrastructure"
)

func TestVinculacionApoderadoYEscolar(testingT *testing.T) {
	testEnvironment := inicializarServidorPruebas(testingT)

	var linkingToken string

	testEnvironment.bus.Registrar(
		modelo.VinculacionSolicitada,
		infra.NuevoObservadorFunc("captura-vinculacion-http", func(event modelo.Event) {
			linkingToken, _ = event.Payload["token"].(string)
		}),
	)

	guardianCookie := ingresar(testingT, testEnvironment.manejador, "apoderado.prueba@example.test")
	studentCookie := ingresar(testingT, testEnvironment.manejador, "escolar.prueba@example.test")

	response := ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/apoderado/vinculaciones",
		cuerpo:    `{"email_escolar":"escolar.prueba@example.test"}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusAccepted,
		fragmento: `"estado":"pendiente"`,
	})

	if linkingToken == "" {
		testingT.Fatal("debía emitirse el token de vinculación")
	}

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/escolar/vinculaciones/consultar",
		cuerpo:    `{"token":"` + linkingToken + `"}`,
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusOK,
		fragmento: "Apoderado Prueba",
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/escolar/vinculaciones/aceptar",
		cuerpo:    `{"token":"` + linkingToken + `"}`,
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusOK,
		fragmento: `"estado":"aceptada"`,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/escolar/vinculaciones/aceptar",
		cuerpo:    `{"token":"` + linkingToken + `"}`,
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusBadRequest,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      "/api/escolar/apoderado",
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusOK,
		fragmento: "apoderado.prueba@example.test",
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      "/api/apoderado/vinculaciones",
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusOK,
		fragmento: "escolar.prueba@example.test",
	})
}

func TestAutorizacionDeNuevasReservas(testingT *testing.T) {
	testEnvironment := inicializarServidorPruebas(testingT)

	var linkingToken string

	testEnvironment.bus.Registrar(
		modelo.VinculacionSolicitada,
		infra.NuevoObservadorFunc("captura-autorizacion-http", func(event modelo.Event) {
			linkingToken, _ = event.Payload["token"].(string)
		}),
	)

	guardianCookie := ingresar(testingT, testEnvironment.manejador, "apoderado.prueba@example.test")
	studentCookie := ingresar(testingT, testEnvironment.manejador, "escolar.prueba@example.test")
	tutorCookie := ingresar(testingT, testEnvironment.manejador, "tutor.prueba@uni.edu.pe")

	response := ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/apoderado/vinculaciones",
		cuerpo:    `{"email_escolar":"escolar.prueba@example.test"}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusAccepted,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/escolar/vinculaciones/aceptar",
		cuerpo:    `{"token":"` + linkingToken + `"}`,
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusOK,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      "/api/escolar/apoderado",
		cookie:    studentCookie,
	})

	var guardianDetails struct {
		Vinculacion struct {
			Escolar struct {
				ID string `json:"id"`
			} `json:"escolar"`
		} `json:"vinculacion"`
	}

	if err := json.Unmarshal(response.Body.Bytes(), &guardianDetails); err != nil {
		testingT.Fatalf("no se pudo interpretar la vinculación: %v", err)
	}

	studentID := guardianDetails.Vinculacion.Escolar.ID
	if studentID == "" {
		testingT.Fatal("no se encontró el identificador del escolar")
	}

	authorizationRoute := "/api/apoderado/escolares/" + studentID + "/autorizacion"

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      authorizationRoute,
		cuerpo:    `{"autorizado":false}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusOK,
		fragmento: `"estado":"no_autorizado"`,
	})

	if !strings.Contains(response.Body.String(), `"creada_en"`) {
		testingT.Fatalf("el cambio debe exponer fecha y hora: %s", response.Body)
	}

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      "/api/escolar/apoderado",
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusOK,
		fragmento: `"autorizado":false`,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      authorizationRoute,
		cuerpo:    `{"autorizado":true}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusOK,
		fragmento: `"estado":"autorizado"`,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      authorizationRoute,
		cuerpo:    `{}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusUnprocessableEntity,
		fragmento: `"autorizado"`,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      authorizationRoute,
		cuerpo:    `{"autorizado":false}`,
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusForbidden,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      authorizationRoute,
		cuerpo:    `{"autorizado":false}`,
		cookie:    tutorCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusForbidden,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      authorizationRoute,
		cuerpo:    `{"autorizado":false}`,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusUnauthorized,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/apoderado/escolares/00000000-0000-0000-0000-000000000000/autorizacion",
		cuerpo:    `{"autorizado":false}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo:    http.StatusConflict,
		fragmento: `"sin_apoderado"`,
	})

	authorizationHistoryRoute := "/api/apoderado/escolares/" + studentID + "/autorizaciones"

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      authorizationHistoryRoute,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusOK,
	})

	var authorizationHistory respuestaHistorialAutorizaciones
	if err := json.Unmarshal(response.Body.Bytes(), &authorizationHistory); err != nil {
		testingT.Fatalf("no se pudo interpretar el historial: %v", err)
	}

	if len(authorizationHistory.Autorizaciones) != 3 {
		testingT.Fatalf("esperaba aceptación, revocación y autorización: %s", response.Body)
	}

	expectedStates := []string{"autorizado", "no_autorizado", "autorizado"}
	for index, expectedState := range expectedStates {
		authorization := authorizationHistory.Autorizaciones[index]
		if authorization.Estado != expectedState || authorization.CreadaEn == "" {
			testingT.Fatalf("historial[%d]=%+v", index, authorization)
		}
		expectedAuthorization := expectedState == "autorizado"
		if authorization.Autorizado != expectedAuthorization {
			testingT.Fatalf("autorizado incoherente en historial[%d]: %+v", index, authorization)
		}
	}

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      authorizationHistoryRoute,
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusForbidden,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      authorizationHistoryRoute,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusUnauthorized,
	})
}

func TestVinculacionRechazaRolesYFaltasDeSesion(testingT *testing.T) {
	testEnvironment := inicializarServidorPruebas(testingT)

	guardianCookie := ingresar(testingT, testEnvironment.manejador, "apoderado.prueba@example.test")
	studentCookie := ingresar(testingT, testEnvironment.manejador, "escolar.prueba@example.test")

	response := ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/apoderado/vinculaciones",
		cuerpo:    `{"email_escolar":"escolar.prueba@example.test"}`,
		cookie:    studentCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusForbidden,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/escolar/vinculaciones/aceptar",
		cuerpo:    `{"token":"x"}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusForbidden,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "GET",
		ruta:      "/api/apoderado/vinculaciones",
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusUnauthorized,
	})

	response = ejecutarPeticion(opcionesPeticion{
		manejador: testEnvironment.manejador,
		metodo:    "POST",
		ruta:      "/api/apoderado/vinculaciones",
		cuerpo:    `{"email_escolar":"tutor.prueba@uni.edu.pe"}`,
		cookie:    guardianCookie,
	})

	verificarRespuestaHTTP(testingT, response, expectativaRespuesta{
		codigo: http.StatusUnprocessableEntity,
	})
}

type expectativaRespuesta struct {
	codigo    int
	fragmento string
}

type registroAutorizacion struct {
	Estado     string `json:"estado"`
	Autorizado bool   `json:"autorizado"`
	CreadaEn   string `json:"creada_en"`
}

type respuestaHistorialAutorizaciones struct {
	Autorizaciones []registroAutorizacion `json:"autorizaciones"`
}

func verificarRespuestaHTTP(
	testingT *testing.T,
	response *httptest.ResponseRecorder,
	expected expectativaRespuesta,
) {
	testingT.Helper()

	if response.Code != expected.codigo {
		testingT.Fatalf("código HTTP esperado: %d, recibido: %d. Respuesta: %s",
			expected.codigo, response.Code, response.Body)
	}

	if expected.fragmento != "" && !strings.Contains(response.Body.String(), expected.fragmento) {
		testingT.Fatalf("la respuesta no contiene %q: %s", expected.fragmento, response.Body)
	}
}
