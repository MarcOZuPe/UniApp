package panel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoria "uniapp/internal/adapters/repositories/memory/acceso"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloPanel "uniapp/internal/application/panel"
	modelo "uniapp/internal/domain/panel"
	infra "uniapp/internal/infrastructure"
)

type repoPanel struct{}

func (repoPanel) Metricas(context.Context, time.Time, string) (*modelo.Metricas, error) {
	return &modelo.Metricas{Reservas: modelo.ReservasMetricas{Totales: 1}}, nil
}
func (repoPanel) Auditoria(context.Context, modelo.FiltroAuditoria) ([]modelo.EventoAuditoria, error) {
	return []modelo.EventoAuditoria{{ID: 1, Accion: "incidente.revisado"}}, nil
}

func entornoPanel(t *testing.T) (http.Handler, *http.Cookie) {
	t.Helper()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: memoria.NuevoRepositorioMemoria(),
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(context.Background(), servicioAcceso, modeloAcceso.DatosSiembra{
		DominioTutor: "uni.edu.pe", Clave: "Clave-Segura-2026!",
	}); err != nil {
		t.Fatal(err)
	}
	token, _, err := servicioAcceso.Sesion.Ingresar(context.Background(), modeloAcceso.CredencialesIngreso{
		Email: "coord.ciencias@example.test", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	controlador := Nuevo(modeloPanel.NuevoServicio(repoPanel{}), controladorAcceso.Autenticar())
	return controlador.Rutas(), &http.Cookie{Name: "uniapp_sesion", Value: token}
}

func TestControladorPanelMetricasYAuditoria(t *testing.T) {
	manejador, cookie := entornoPanel(t)
	peticion := httptest.NewRequest(http.MethodGet, "/api/coordinacion/metricas", nil)
	peticion.AddCookie(cookie)
	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK || respuesta.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("métricas inesperadas: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/coordinacion/auditoria?limite=10", nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("auditoría inesperada: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/coordinacion/auditoria?limite=0", nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusBadRequest {
		t.Fatalf("límite inválido inesperado: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/coordinacion/metricas", nil)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión debe ser 401: %d %s", respuesta.Code, respuesta.Body)
	}
}
