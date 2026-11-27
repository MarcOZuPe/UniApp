package oferta

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/oferta"
	infra "uniapp/internal/infrastructure"
)

type repoOferta struct{}

func (repoOferta) ObtenerDisponibilidad(context.Context, string) (modelo.CupoSemanal, []modelo.SlotSemanal, error) {
	return modelo.CupoSemanal{Minutos: 180}, []modelo.SlotSemanal{{
		ID: 1, DiaSemana: time.Monday, Inicio: 9 * time.Hour, Fin: 10 * time.Hour,
	}}, nil
}
func (repoOferta) AgregarSlots(_ context.Context, _ string, slots []modelo.SlotSemanal) (modelo.CupoSemanal, []modelo.SlotSemanal, error) {
	return modelo.CupoSemanal{Minutos: 180}, slots, nil
}
func (repoOferta) EliminarSlot(context.Context, string, int64) error { return nil }
func (repoOferta) CambiarCupo(context.Context, string, int) error    { return nil }

func entornoOferta(t *testing.T) (http.Handler, *http.Cookie) {
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
	token, _, err := acceso.Sesion.Ingresar(context.Background(), modeloAcceso.CredencialesIngreso{
		Email: "tutor.prueba@uni.edu.pe", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	return Rutas(repoOferta{}, ctrlAcceso.Nuevo(acceso, false).Autenticar()), &http.Cookie{Name: "uniapp_sesion", Value: token}
}

func TestRutasOferta(t *testing.T) {
	manejador, cookie := entornoOferta(t)
	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, "/api/oferta/salud", nil))
	if respuesta.Code != http.StatusOK {
		t.Fatalf("salud: %d", respuesta.Code)
	}
	peticion := httptest.NewRequest(http.MethodGet, "/api/oferta/slots", nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("slots: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodPost, "/api/oferta/slots", strings.NewReader(
		`{"slots":[{"dia_semana":1,"inicio":"09:00","fin":"10:00"}]}`,
	))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusCreated {
		t.Fatalf("agregar slots: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodPost, "/api/oferta/slots", strings.NewReader(`{"slots":[{"dia_semana":9,"inicio":"09:00","fin":"10:00"}]}`))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusUnprocessableEntity {
		t.Fatalf("slots inválidos: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodDelete, "/api/oferta/slots/1", nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusNoContent {
		t.Fatalf("eliminar slot: %d", respuesta.Code)
	}
	peticion = httptest.NewRequest(http.MethodPut, "/api/oferta/cupo", strings.NewReader(`{"minutos":180}`))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("cambiar cupo: %d %s", respuesta.Code, respuesta.Body)
	}
}
