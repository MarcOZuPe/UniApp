package acceso

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modelo "uniapp/internal/application/acceso"
)

func TestMapeadorDeErroresSeExtiendeSinCambiarElRespondedor(testingT *testing.T) {
	mapeador := NuevoMapeadorDeErrores()
	errorPropio := errors.New("error propio")
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, errorPropio) },
		respuestaConEstado(http.StatusTeapot, "propio"),
	)

	respuesta := httptest.NewRecorder()
	mapeador.Escribir(respuesta, errorPropio)
	verificarRespuestaError(testingT, respuesta, expectativaError{
		codigo:    http.StatusTeapot,
		fragmento: `"codigo":"propio"`,
	})

	respuesta = httptest.NewRecorder()
	mapeador.Escribir(respuesta, errors.New("desconocido"))
	verificarRespuestaError(testingT, respuesta, expectativaError{
		codigo:    http.StatusInternalServerError,
		fragmento: `"codigo":"error_interno"`,
	})
}

func TestMapeadorDeErroresRespetaElOrdenDeLaCadena(testingT *testing.T) {
	mapeador := NuevoMapeadorDeErrores()
	errorPropio := errors.New("error propio")
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, errorPropio) },
		respuestaConEstado(http.StatusTeapot, "primero"),
	)
	mapeador.Registrar(
		func(err error) bool { return errors.Is(err, errorPropio) },
		respuestaConEstado(http.StatusInternalServerError, "segundo"),
	)

	respuesta := httptest.NewRecorder()
	mapeador.Escribir(respuesta, errorPropio)
	verificarRespuestaError(testingT, respuesta, expectativaError{
		codigo:    http.StatusTeapot,
		fragmento: `"codigo":"primero"`,
	})
}

func TestMapeadorDeErroresDeCuentaMantieneValidacion(testingT *testing.T) {
	mapeador := NuevoMapeadorDeErrores()
	registrarErroresDeCuenta(mapeador)

	respuesta := httptest.NewRecorder()
	mapeador.Escribir(respuesta, modelo.ErrValidacion{Campos: map[string]string{"email": "obligatorio"}})
	verificarRespuestaError(testingT, respuesta, expectativaError{
		codigo:    http.StatusUnprocessableEntity,
		fragmento: `"codigo":"validacion"`,
	})
	if !strings.Contains(respuesta.Body.String(), `"email":"obligatorio"`) {
		testingT.Fatalf("validación sin campos: %s", respuesta.Body)
	}
}

type expectativaError struct {
	codigo    int
	fragmento string
}

func verificarRespuestaError(
	testingT *testing.T,
	respuesta *httptest.ResponseRecorder,
	expected expectativaError,
) {
	testingT.Helper()

	if respuesta.Code != expected.codigo {
		testingT.Fatalf("código esperado %d, recibido %d: %s",
			expected.codigo, respuesta.Code, respuesta.Body)
	}
	if expected.fragmento != "" && !strings.Contains(respuesta.Body.String(), expected.fragmento) {
		testingT.Fatalf("la respuesta no contiene %q: %s", expected.fragmento, respuesta.Body)
	}
}
