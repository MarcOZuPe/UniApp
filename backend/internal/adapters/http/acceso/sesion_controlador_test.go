package acceso

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
	modelo "uniapp/internal/application/acceso"
	infra "uniapp/internal/infrastructure"
)

// repositorioConFalloAlSalir delega en el repositorio en memoria y solo
// intercepta el cierre de sesión para simular un fallo de persistencia.
type repositorioConFalloAlSalir struct {
	*memoria.RepositorioMemoria
	err error
}

func (repositorio repositorioConFalloAlSalir) EliminarSesion(context.Context, []byte) error {
	return repositorio.err
}

// B-10: el logout ignora el error de eliminar la sesión y responde 204,
// dejando la sesión activa en el servidor sin avisar al cliente.
func TestB10SalidaNoDebeOcultarElErrorAlCerrarSesion(t *testing.T) {
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	base := memoria.NuevoRepositorioMemoria()
	servicio := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: repositorioConFalloAlSalir{
			RepositorioMemoria: base,
			err:                errors.New("postgres no disponible"),
		},
		Fabrica: modelo.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:  hasher,
		Bus:     infra.NuevoBus(),
	})
	if err := modelo.Sembrar(context.Background(), servicio, modelo.DatosSiembra{
		DominioTutor: "uni.edu.pe", Clave: clavePruebaHTTP,
	}); err != nil {
		t.Fatal(err)
	}

	manejador := Encadenar(Nuevo(servicio, false).Rutas(), OrigenPermitido(origenPruebaHTTP))

	ingreso := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/ingreso",
		strings.NewReader(`{"email":"`+modeloEmailAdministrador+`","clave":"`+clavePruebaHTTP+`"}`),
	)
	ingreso.Header.Set("Content-Type", "application/json")
	ingreso.Header.Set("Origin", origenPruebaHTTP)
	respuestaIngreso := httptest.NewRecorder()
	manejador.ServeHTTP(respuestaIngreso, ingreso)
	if respuestaIngreso.Code != http.StatusOK {
		t.Fatalf("ingreso previo: %d %s", respuestaIngreso.Code, respuestaIngreso.Body)
	}
	var cookieSesion *http.Cookie
	for _, cookie := range respuestaIngreso.Result().Cookies() {
		if cookie.Name == nombreCookie {
			cookieSesion = cookie
		}
	}
	if cookieSesion == nil {
		t.Fatal("el ingreso no devolvió la cookie de sesión")
	}

	salida := httptest.NewRequest(http.MethodPost, "/api/auth/salida", nil)
	salida.Header.Set("Origin", origenPruebaHTTP)
	salida.AddCookie(cookieSesion)
	respuestaSalida := httptest.NewRecorder()
	manejador.ServeHTTP(respuestaSalida, salida)

	if respuestaSalida.Code != http.StatusInternalServerError {
		t.Fatalf(
			"B-10: un fallo al cerrar la sesión debe reportarse, se obtuvo %d con cuerpo %q",
			respuestaSalida.Code, respuestaSalida.Body.String(),
		)
	}
}
