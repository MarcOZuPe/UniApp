package catalogo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoria "uniapp/internal/adapters/repositories/memory/catalogo"
	modelo "uniapp/internal/application/catalogo"
)

func TestControladorCatalogoPublico(t *testing.T) {
	repo := memoria.NuevoRepositorioMemoria()
	servicio := modelo.NuevoServicio(repo)
	if _, err := servicio.Importar(context.Background(), modelo.DatosImportacion{
		Fuente: "origen", Cursos: []modelo.CursoEntrada{{Nombre: "Álgebra", Grado: "5to"}},
		Colegios: []modelo.ColegioEntrada{{Nombre: "Colegio", Distrito: "Lima"}},
	}); err != nil {
		t.Fatal(err)
	}
	controlador := Nuevo(servicio, ctrlAcceso.Nuevo(nil, false).Autenticar())
	manejador := controlador.Rutas()

	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, "/api/catalogo", nil))
	if respuesta.Code != http.StatusOK || respuesta.Header().Get("Cache-Control") == "" {
		t.Fatalf("catálogo público inesperado: %d %s", respuesta.Code, respuesta.Body)
	}
}
