package catalogo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoriaCatalogo "uniapp/internal/adapters/repositories/memory/catalogo"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/catalogo"
	infra "uniapp/internal/infrastructure"
)

func entornoCatalogo(t *testing.T) (http.Handler, *http.Cookie, *http.Cookie) {
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
	coordToken, _, err := acceso.Sesion.Ingresar(context.Background(), modeloAcceso.CredencialesIngreso{
		Email: "coord.ciencias@example.test", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	escolarToken, _, err := acceso.Sesion.Ingresar(context.Background(), modeloAcceso.CredencialesIngreso{
		Email: "escolar.prueba@example.test", Clave: "Clave-Segura-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	servicio := modelo.NuevoServicio(memoriaCatalogo.NuevoRepositorioMemoria())
	controlador := Nuevo(servicio, ctrlAcceso.Nuevo(acceso, false).Autenticar())
	return controlador.Rutas(),
		&http.Cookie{Name: "uniapp_sesion", Value: coordToken},
		&http.Cookie{Name: "uniapp_sesion", Value: escolarToken}
}

func TestControladorCatalogoGestion(t *testing.T) {
	manejador, coord, escolar := entornoCatalogo(t)
	peticion := httptest.NewRequest(http.MethodPost, "/api/catalogo/importaciones", strings.NewReader(
		`{"fuente":"origen","cursos":[{"nombre":"Álgebra","grado":"5to"}],"colegios":[{"nombre":"Colegio","distrito":"Lima"}]}`,
	))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(coord)
	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusCreated {
		t.Fatalf("importar: %d %s", respuesta.Code, respuesta.Body)
	}

	peticion = httptest.NewRequest(http.MethodGet, "/api/catalogo/cursos/pendientes", nil)
	peticion.AddCookie(coord)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("pendientes: %d %s", respuesta.Code, respuesta.Body)
	}
	var pendientes struct {
		Cursos []struct {
			ID int64 `json:"id"`
		} `json:"cursos"`
	}
	if err := json.Unmarshal(respuesta.Body.Bytes(), &pendientes); err != nil || len(pendientes.Cursos) != 1 {
		t.Fatalf("pendientes inesperados: %s err=%v", respuesta.Body, err)
	}
	rutaRevision := "/api/catalogo/cursos/" + strconv.FormatInt(pendientes.Cursos[0].ID, 10) + "/revision"
	peticion = httptest.NewRequest(http.MethodPost, rutaRevision, strings.NewReader(`{"estado":"aprobado"}`))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(coord)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("revisar: %d %s", respuesta.Code, respuesta.Body)
	}

	peticion = httptest.NewRequest(http.MethodPost, "/api/catalogo/colegios/solicitudes", strings.NewReader(`{"nombre":"Otro","distrito":"Surco"}`))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.AddCookie(escolar)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusCreated {
		t.Fatalf("solicitar colegio: %d %s", respuesta.Code, respuesta.Body)
	}
	peticion = httptest.NewRequest(http.MethodGet, "/api/catalogo/colegios/solicitudes", nil)
	peticion.AddCookie(coord)
	respuesta = httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("solicitudes: %d %s", respuesta.Code, respuesta.Body)
	}
}
