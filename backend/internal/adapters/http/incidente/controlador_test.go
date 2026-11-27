package incidente

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoriaIncidente "uniapp/internal/adapters/repositories/memory/incidente"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/incidente"
	infra "uniapp/internal/infrastructure"
)

type almacenIncidente struct{ datos map[string][]byte }

func (a *almacenIncidente) Guardar(_ context.Context, clave string, contenido []byte) error {
	if a.datos == nil {
		a.datos = map[string][]byte{}
	}
	a.datos[clave] = contenido
	return nil
}
func (a *almacenIncidente) Leer(_ context.Context, clave string) ([]byte, error) {
	return a.datos[clave], nil
}
func (a *almacenIncidente) Eliminar(_ context.Context, clave string) error {
	delete(a.datos, clave)
	return nil
}

func entornoIncidente(t *testing.T) (http.Handler, *http.Cookie, string) {
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
	escolar, err := acceso.RepositorioUsuarios.UsuarioPorEmail(context.Background(), "escolar.prueba@example.test")
	if err != nil {
		t.Fatal(err)
	}
	repo := memoriaIncidente.NuevoRepositorioMemoria()
	inicio := time.Now().UTC().Add(-2 * time.Hour)
	repo.SembrarTutoria(modelo.Tutoria{
		ReservaID: "reserva-1", EscolarID: escolar.ID, EscolarNombre: escolar.NombreUso,
		TutorID: "tutor-1", TutorNombre: "Tutor", TutorFacultadID: "1", Curso: "Álgebra",
		Inicio: inicio, Fin: inicio.Add(time.Hour),
	})
	servicio := modelo.NuevoServicio(repo, &almacenIncidente{})
	servicio.FijarReloj(func() time.Time { return inicio.Add(3 * time.Hour) })
	if _, err := servicio.Registrar(context.Background(), modelo.Solicitante{ID: escolar.ID, Rol: modeloAcceso.RolEscolar}, modelo.DatosIncidente{
		EscolarID: escolar.ID, ReservaID: "reserva-1", Tipo: modelo.TipoTratoInadecuado,
		Descripcion: "Descripción válida del incidente",
		Evidencias:  []modelo.Archivo{{Nombre: "a.pdf", Contenido: []byte("%PDF-1.4 contenido")}},
	}); err != nil {
		t.Fatal(err)
	}
	token, _, err := acceso.Sesion.Ingresar(context.Background(), modeloAcceso.CredencialesIngreso{
		Email: "escolar.prueba@example.test", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	controlador := Nuevo(servicio, ctrlAcceso.Nuevo(acceso, false).Autenticar())
	return controlador.Rutas(), &http.Cookie{Name: "uniapp_sesion", Value: token}, escolar.ID
}

func TestControladorIncidenteListarDetalleYEvidencia(t *testing.T) {
	manejador, cookie, _ := entornoIncidente(t)
	for _, ruta := range []string{"/api/incidentes/salud", "/api/incidentes"} {
		peticion := httptest.NewRequest(http.MethodGet, ruta, nil)
		peticion.AddCookie(cookie)
		respuesta := httptest.NewRecorder()
		manejador.ServeHTTP(respuesta, peticion)
		if respuesta.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", ruta, respuesta.Code, respuesta.Body)
		}
	}
	peticion := httptest.NewRequest(http.MethodGet, "/api/incidentes", nil)
	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión debe ser 401: %d", respuesta.Code)
	}

	var listado struct {
		Incidentes []struct {
			ID         string `json:"id"`
			Evidencias []struct {
				ID string `json:"id"`
			} `json:"evidencias"`
		} `json:"incidentes"`
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/incidentes", nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if err := json.Unmarshal(respuesta.Body.Bytes(), &listado); err != nil || len(listado.Incidentes) != 1 || len(listado.Incidentes[0].Evidencias) != 1 {
		t.Fatalf("listado inesperado: %s err=%v", respuesta.Body, err)
	}
	id := listado.Incidentes[0].ID
	evidenciaID := listado.Incidentes[0].Evidencias[0].ID
	peticion = httptest.NewRequest(http.MethodGet, "/api/incidentes/"+id, nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("detalle: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/incidentes/"+id+"/evidencias/"+evidenciaID, nil)
	peticion.AddCookie(cookie)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK || respuesta.Body.Len() == 0 {
		t.Fatalf("evidencia: %d %s", respuesta.Code, respuesta.Body)
	}
}
