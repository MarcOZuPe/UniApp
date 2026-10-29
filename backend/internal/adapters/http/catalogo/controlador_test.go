package catalogo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoria "uniapp/internal/adapters/repositories/memory/catalogo"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/catalogo"
	"uniapp/internal/infrastructure"
)

const (
	origenCatalogo = "http://localhost:5173"
	claveCatalogo  = "Clave-Segura-2026!"
)

func TestCatalogoImportacionAprobacionYPermisos(t *testing.T) {
	handler := nuevoManejadorCatalogo(t)
	coord := iniciarSesionCatalogo(t, handler, "coord.prueba@example.test")
	escolar := iniciarSesionCatalogo(t, handler, "escolar.prueba@example.test")
	tutor := iniciarSesionCatalogo(t, handler, "tutor.prueba@uni.edu.pe")
	apoderado := iniciarSesionCatalogo(t, handler, "apoderado.prueba@example.test")

	importacionJSON := `{
		"fuente":"sintetico: prueba I07",
		"cursos":[
			{"nombre":"Álgebra","grado":"1.° de secundaria"},
			{"nombre":"Ciencias","grado":"2.° de secundaria"}
		],
		"colegios":[{"nombre":"Colegio Sintético","distrito":"Distrito Sintético"}]
	}`
	if respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/importaciones", nil, importacionJSON); respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("importación sin sesión: estado %d, cuerpo %s", respuesta.Code, respuesta.Body.String())
	}
	if respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/importaciones", escolar, importacionJSON); respuesta.Code != http.StatusForbidden {
		t.Fatalf("un escolar no puede importar: estado %d", respuesta.Code)
	}
	respuestaImportacion := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/importaciones", coord, importacionJSON)
	if respuestaImportacion.Code != http.StatusCreated {
		t.Fatalf("importación de coordinación: estado %d, cuerpo %s", respuestaImportacion.Code, respuestaImportacion.Body.String())
	}
	var importada struct {
		Fuente string `json:"fuente"`
		Cursos []struct {
			ID int64 `json:"id"`
		} `json:"cursos"`
	}
	if err := json.Unmarshal(respuestaImportacion.Body.Bytes(), &importada); err != nil {
		t.Fatal(err)
	}
	if importada.Fuente != "sintetico: prueba I07" || len(importada.Cursos) != 2 {
		t.Fatalf("respuesta de importación: %+v", importada)
	}

	respuestaPublica := solicitudCatalogo(handler, http.MethodGet, "/api/catalogo", nil, "")
	if respuestaPublica.Code != http.StatusOK || !strings.Contains(respuestaPublica.Body.String(), "importado_en") {
		t.Fatalf("catálogo público debe mostrar origen y fecha: estado %d, cuerpo %s", respuestaPublica.Code, respuestaPublica.Body.String())
	}
	if strings.Contains(respuestaPublica.Body.String(), "Álgebra") {
		t.Fatal("los cursos pendientes no deben aparecer en el catálogo público")
	}
	if respuesta := solicitudCatalogo(handler, http.MethodGet, "/api/catalogo/cursos/pendientes", escolar, ""); respuesta.Code != http.StatusForbidden {
		t.Fatalf("un escolar no puede revisar cursos: estado %d", respuesta.Code)
	}
	respuestaPendientes := solicitudCatalogo(handler, http.MethodGet, "/api/catalogo/cursos/pendientes", coord, "")
	if respuestaPendientes.Code != http.StatusOK || !strings.Contains(respuestaPendientes.Body.String(), "Álgebra") {
		t.Fatalf("coordinación debe obtener los cursos pendientes: estado %d, cuerpo %s", respuestaPendientes.Code, respuestaPendientes.Body.String())
	}
	decision := `{"estado":"aprobado"}`
	rutaRevision := "/api/catalogo/cursos/" + strconv.FormatInt(importada.Cursos[0].ID, 10) + "/revision"
	if respuesta := solicitudCatalogo(handler, http.MethodPost, rutaRevision, escolar, decision); respuesta.Code != http.StatusForbidden {
		t.Fatalf("un escolar no puede aprobar cursos: estado %d", respuesta.Code)
	}
	if respuesta := solicitudCatalogo(handler, http.MethodPost, rutaRevision, coord, decision); respuesta.Code != http.StatusOK {
		t.Fatalf("coordinación debe aprobar el curso por grado: estado %d, cuerpo %s", respuesta.Code, respuesta.Body.String())
	}
	if respuesta := solicitudCatalogo(handler, http.MethodPost, rutaRevision, coord, decision); respuesta.Code != http.StatusConflict {
		t.Fatalf("el curso solo debe revisarse una vez: estado %d, cuerpo %s", respuesta.Code, respuesta.Body.String())
	}
	respuestaPublica = solicitudCatalogo(handler, http.MethodGet, "/api/catalogo", nil, "")
	if !strings.Contains(respuestaPublica.Body.String(), "Álgebra") || strings.Contains(respuestaPublica.Body.String(), "Ciencias") {
		t.Fatalf("solo el curso aprobado debe ser público: %s", respuestaPublica.Body.String())
	}
	if respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/importaciones", coord,
		`{"fuente":"test","cursos":[{"nombre":"x","grado":"1"},{"nombre":"X","grado":"1"}],"colegios":[{"nombre":"y","distrito":"z"}]}`); respuesta.Code != http.StatusUnprocessableEntity {
		t.Fatalf("debe validarse una importación duplicada: estado %d, cuerpo %s", respuesta.Code, respuesta.Body.String())
	}

	solicitudColegio := `{"nombre":"Colegio Faltante","distrito":"Distrito Nuevo"}`
	if respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/colegios/solicitudes", tutor, solicitudColegio); respuesta.Code != http.StatusForbidden {
		t.Fatalf("un tutor no puede solicitar un colegio: estado %d", respuesta.Code)
	}
	if respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/colegios/solicitudes", escolar, solicitudColegio); respuesta.Code != http.StatusCreated {
		t.Fatalf("un escolar puede solicitar un colegio: estado %d, cuerpo %s", respuesta.Code, respuesta.Body.String())
	}
	if respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/colegios/solicitudes", apoderado, solicitudColegio); respuesta.Code != http.StatusConflict {
		t.Fatalf("la solicitud duplicada debe rechazarse: estado %d", respuesta.Code)
	}
	respuestaSolicitudes := solicitudCatalogo(handler, http.MethodGet, "/api/catalogo/colegios/solicitudes", coord, "")
	if respuestaSolicitudes.Code != http.StatusOK ||
		strings.Contains(respuestaSolicitudes.Body.String(), "solicitante_id") ||
		!strings.Contains(respuestaSolicitudes.Body.String(), "Colegio Faltante") {
		t.Fatalf("coordinación ve solicitud sin exponer quién la hizo: estado %d, cuerpo %s", respuestaSolicitudes.Code, respuestaSolicitudes.Body.String())
	}
}

func TestSolicitudColegioRequiereJSONValido(t *testing.T) {
	handler := nuevoManejadorCatalogo(t)
	escolar := iniciarSesionCatalogo(t, handler, "escolar.prueba@example.test")
	for _, cuerpo := range []string{
		`{"nombre":"Colegio","distrito":"Distrito","campo_no_permitido":true}`,
		`{"nombre":"Colegio","distrito":"Distrito"} {"nombre":"otro"}`,
	} {
		respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/catalogo/colegios/solicitudes", escolar, cuerpo)
		if respuesta.Code != http.StatusBadRequest {
			t.Fatalf("cuerpo inválido debe rechazarse: estado %d, cuerpo %s", respuesta.Code, respuesta.Body.String())
		}
	}
}

func nuevoManejadorCatalogo(t *testing.T) http.Handler {
	t.Helper()
	ctx := context.Background()
	repoAcceso := memoriaAcceso.NuevoRepositorioMemoria()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: repoAcceso,
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(ctx, servicioAcceso, modeloAcceso.OpcionesSiembra{
		DominioTutor: "uni.edu.pe", Clave: claveCatalogo,
	}); err != nil {
		t.Fatal(err)
	}
	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	controladorCatalogo := Nuevo(modelo.NuevoServicio(memoria.NuevoRepositorioMemoria()), controladorAcceso.Autenticar())
	mux := http.NewServeMux()
	mux.Handle("/api/catalogo", controladorCatalogo.Rutas())
	mux.Handle("/api/catalogo/", controladorCatalogo.Rutas())
	mux.Handle("/", controladorAcceso.Rutas())
	return ctrlAcceso.Encadenar(mux, ctrlAcceso.OrigenPermitido(origenCatalogo))
}

func iniciarSesionCatalogo(t *testing.T, handler http.Handler, email string) *http.Cookie {
	t.Helper()
	cuerpo, err := json.Marshal(map[string]string{"email": email, "clave": claveCatalogo})
	if err != nil {
		t.Fatal(err)
	}
	respuesta := solicitudCatalogo(handler, http.MethodPost, "/api/auth/ingreso", nil, string(cuerpo))
	if respuesta.Code != http.StatusOK {
		t.Fatalf("ingreso %s: estado %d, cuerpo %s", email, respuesta.Code, respuesta.Body.String())
	}
	for _, cookie := range respuesta.Result().Cookies() {
		if cookie.Name == "uniapp_sesion" {
			return cookie
		}
	}
	t.Fatal("respuesta de ingreso no incluye cookie")
	return nil
}

func solicitudCatalogo(handler http.Handler, metodo, ruta string, cookie *http.Cookie, cuerpo string) *httptest.ResponseRecorder {
	var body *strings.Reader
	if cuerpo == "" {
		body = strings.NewReader("")
	} else {
		body = strings.NewReader(cuerpo)
	}
	peticion := httptest.NewRequest(metodo, ruta, body)
	peticion.Header.Set("Origin", origenCatalogo)
	if metodo == http.MethodPost {
		peticion.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)
	return respuesta
}
