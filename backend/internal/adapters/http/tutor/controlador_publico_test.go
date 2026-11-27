package tutor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoria "uniapp/internal/adapters/repositories/memory/tutor"
	modelo "uniapp/internal/application/tutor"
)

type almacenTutor struct{}

func (almacenTutor) Guardar(context.Context, string, []byte) error { return nil }
func (almacenTutor) Leer(context.Context, string) ([]byte, error)  { return nil, nil }
func (almacenTutor) Eliminar(context.Context, string) error        { return nil }

func TestControladorTutorPerfilesPublicos(t *testing.T) {
	repo := memoria.NuevoRepositorioMemoria()
	vencimiento := time.Now().Add(24 * time.Hour)
	expediente := &modelo.Expediente{
		TutorID: "t-1", FacultadID: "1", Presentacion: "Presentación", Cursos: "Álgebra",
		Experiencia: "Experiencia", Estado: modelo.EstadoAprobado, VenceEn: &vencimiento,
		Documentos: []modelo.Documento{{ID: "d-1", Nombre: "a.pdf"}},
	}
	if err := repo.Guardar(context.Background(), expediente); err != nil {
		t.Fatal(err)
	}
	servicio := modelo.NuevoServicio(repo, almacenTutor{})
	controlador := Nuevo(servicio, ctrlAcceso.Nuevo(nil, false).Autenticar())
	manejador := controlador.Rutas()

	for _, ruta := range []string{"/api/tutor/perfiles", "/api/tutor/perfiles/t-1"} {
		respuesta := httptest.NewRecorder()
		manejador.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, ruta, nil))
		if respuesta.Code != http.StatusOK || respuesta.Header().Get("Cache-Control") == "" {
			t.Fatalf("%s inesperado: %d %s", ruta, respuesta.Code, respuesta.Body)
		}
	}
	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, httptest.NewRequest(http.MethodGet, "/api/tutor/perfiles/inexistente", nil))
	if respuesta.Code != http.StatusNotFound {
		t.Fatalf("perfil inexistente inesperado: %d %s", respuesta.Code, respuesta.Body)
	}
}
