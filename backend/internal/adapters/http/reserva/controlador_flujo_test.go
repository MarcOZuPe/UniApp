package reserva

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoriaReserva "uniapp/internal/adapters/repositories/memory/reserva"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/reserva"
	infra "uniapp/internal/infrastructure"
)

type autorizacionSiempre struct{}

func (autorizacionSiempre) AutorizacionVigente(context.Context, string) (bool, error) {
	return true, nil
}

type almacenReserva struct{ datos map[string][]byte }

func (a *almacenReserva) Guardar(_ context.Context, clave string, contenido []byte) error {
	if a.datos == nil {
		a.datos = map[string][]byte{}
	}
	a.datos[clave] = contenido
	return nil
}
func (a *almacenReserva) Leer(_ context.Context, clave string) ([]byte, error) {
	return a.datos[clave], nil
}
func (a *almacenReserva) Eliminar(_ context.Context, clave string) error {
	delete(a.datos, clave)
	return nil
}

func entornoReserva(t *testing.T) (http.Handler, *http.Cookie) {
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
		Email: "escolar.prueba@example.test", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	servicio := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio:  memoriaReserva.NuevoRepositorioMemoria(),
		Oferta:       modelo.NuevaOfertaSintetica(time.Now().UTC()),
		Autorizacion: autorizacionSiempre{},
		Almacen:      &almacenReserva{},
		Analizador:   modelo.AnalizadorBasico{},
	})
	controlador := Nuevo(servicio, ctrlAcceso.Nuevo(acceso, false).Autenticar())
	return controlador.Rutas(), &http.Cookie{Name: "uniapp_sesion", Value: token}
}

func TestControladorReservaFlujo(t *testing.T) {
	manejador, cookie := entornoReserva(t)
	peticion := httptest.NewRequest(http.MethodGet, "/api/reserva/slots", nil)
	peticion.AddCookie(cookie)
	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("slots: %d %s", respuesta.Code, respuesta.Body)
	}
	var slots struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	if err := json.Unmarshal(respuesta.Body.Bytes(), &slots); err != nil || len(slots.Slots) == 0 {
		t.Fatalf("slots inesperados: %s err=%v", respuesta.Body, err)
	}

	cuerpo := `{"slot_id":"` + slots.Slots[0].ID + `","tema":"Ecuaciones lineales","objetivo":"Resolver ejercicios de ecuaciones lineales","grado":"5to de secundaria"}`
	peticion = httptest.NewRequest(http.MethodPost, "/api/reserva", strings.NewReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusCreated {
		t.Fatalf("crear: %d %s", respuesta.Code, respuesta.Body)
	}
	var creada struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respuesta.Body.Bytes(), &creada); err != nil || creada.ID == "" {
		t.Fatalf("reserva creada inesperada: %s", respuesta.Body)
	}

	peticion = httptest.NewRequest(http.MethodGet, "/api/reserva/"+creada.ID, nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("detalle: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/reserva/"+creada.ID+"/adjuntos", nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("adjuntos: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/reserva/autorizacion", nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("autorización: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/reserva/"+creada.ID, nil)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", respuesta.Code)
	}
}
