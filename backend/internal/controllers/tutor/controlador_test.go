package tutor

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/controllers/acceso"
	"uniapp/internal/infra"
	modeloAcceso "uniapp/internal/models/acceso"
	modelo "uniapp/internal/models/tutor"
)

const (
	origenTutor = "http://localhost:5173"
	claveTutor  = "Clave-Segura-2026!"
)

type almacenFalso struct {
	documentos map[string][]byte
}

func (a *almacenFalso) Guardar(_ context.Context, id string, contenido []byte) error {
	a.documentos[id] = append([]byte(nil), contenido...)
	return nil
}

func (a *almacenFalso) Leer(_ context.Context, id string) ([]byte, error) {
	contenido, existe := a.documentos[id]
	if !existe {
		return nil, modelo.ErrNoEncontrado
	}
	return append([]byte(nil), contenido...), nil
}

func (a *almacenFalso) Eliminar(_ context.Context, id string) error {
	delete(a.documentos, id)
	return nil
}

func TestExpedienteRevisionYDocumentosPrivados(t *testing.T) {
	ctx := context.Background()
	repoAcceso := modeloAcceso.NuevoRepositorioMemoria()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: repoAcceso,
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(ctx, servicioAcceso, modeloAcceso.OpcionesSiembra{
		DominioTutor: "uni.edu.pe", Clave: claveTutor,
	}); err != nil {
		t.Fatal(err)
	}
	tutor, err := repoAcceso.UsuarioPorEmail(ctx, "tutor.prueba@uni.edu.pe")
	if err != nil {
		t.Fatal(err)
	}

	almacen := &almacenFalso{documentos: map[string][]byte{}}
	servicioTutor := modelo.NuevoServicio(modelo.NuevoRepositorioMemoria(), almacen)
	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	controladorTutor := Nuevo(servicioTutor, controladorAcceso.Autenticar())
	mux := http.NewServeMux()
	mux.Handle("/api/tutor/", controladorTutor.Rutas())
	mux.Handle("/", controladorAcceso.Rutas())
	handler := ctrlAcceso.Encadenar(mux, ctrlAcceso.OrigenPermitido(origenTutor))

	if respuesta := hacerSolicitud(handler, http.MethodGet, "/api/tutor/revision", nil, nil); respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("revisión sin sesión: estado %d, cuerpo %s", respuesta.Code, respuesta.Body.String())
	}
	if respuesta := hacerSolicitud(handler, http.MethodGet, "/api/tutor/perfiles/"+tutor.ID, nil, nil); respuesta.Code != http.StatusNotFound {
		t.Fatalf("perfil pendiente no debe publicarse: estado %d", respuesta.Code)
	}

	cookieTutor := iniciarSesionTutor(t, handler, "tutor.prueba@uni.edu.pe")
	respuestaPresentacion := presentarFormularioTutor(t, handler, cookieTutor)
	if respuestaPresentacion.Code != http.StatusCreated {
		t.Fatalf("presentar expediente: estado %d, cuerpo %s", respuestaPresentacion.Code, respuestaPresentacion.Body.String())
	}
	var presentado struct {
		Documentos []struct {
			ID string `json:"id"`
		} `json:"documentos"`
	}
	if err := json.Unmarshal(respuestaPresentacion.Body.Bytes(), &presentado); err != nil {
		t.Fatal(err)
	}
	if len(presentado.Documentos) != 1 {
		t.Fatalf("documentos devueltos: %d", len(presentado.Documentos))
	}

	cookieApoderado := iniciarSesionTutor(t, handler, "apoderado.prueba@example.test")
	if respuesta := hacerSolicitud(handler, http.MethodGet, "/api/tutor/documentos/"+presentado.Documentos[0].ID, cookieApoderado, nil); respuesta.Code != http.StatusForbidden {
		t.Fatalf("persona ajena no debe descargar el documento: estado %d", respuesta.Code)
	}
	if respuesta := hacerSolicitud(handler, http.MethodGet, "/api/tutor/revision", cookieTutor, nil); respuesta.Code != http.StatusForbidden {
		t.Fatalf("tutor no debe acceder a revisión: estado %d", respuesta.Code)
	}

	cookieCoord := iniciarSesionTutor(t, handler, "coord.prueba@example.test")
	respuestaRevision := hacerSolicitud(handler, http.MethodGet, "/api/tutor/revision", cookieCoord, nil)
	if respuestaRevision.Code != http.StatusOK || !strings.Contains(respuestaRevision.Body.String(), tutor.ID) {
		t.Fatalf("coordinación no obtiene la cola: estado %d, cuerpo %s", respuestaRevision.Code, respuestaRevision.Body.String())
	}
	venceEn := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	cuerpo, err := json.Marshal(map[string]string{"estado": "aprobado", "vence_en": venceEn})
	if err != nil {
		t.Fatal(err)
	}
	respuestaAprobacion := hacerSolicitud(handler, http.MethodPost, "/api/tutor/"+tutor.ID+"/revision", cookieCoord, bytes.NewReader(cuerpo))
	if respuestaAprobacion.Code != http.StatusOK {
		t.Fatalf("aprobar expediente: estado %d, cuerpo %s", respuestaAprobacion.Code, respuestaAprobacion.Body.String())
	}

	respuestaPublica := hacerSolicitud(handler, http.MethodGet, "/api/tutor/perfiles/"+tutor.ID, nil, nil)
	if respuestaPublica.Code != http.StatusOK || strings.Contains(respuestaPublica.Body.String(), presentado.Documentos[0].ID) {
		t.Fatalf("el perfil aprobado debe ser público pero sin documentos: estado %d, cuerpo %s", respuestaPublica.Code, respuestaPublica.Body.String())
	}
	respuestaDocumento := hacerSolicitud(handler, http.MethodGet, "/api/tutor/documentos/"+presentado.Documentos[0].ID, cookieCoord, nil)
	if respuestaDocumento.Code != http.StatusOK || respuestaDocumento.Header().Get("Content-Type") != "application/pdf" {
		t.Fatalf("coordinación debe descargar el PDF: estado %d", respuestaDocumento.Code)
	}
}

func iniciarSesionTutor(t *testing.T, handler http.Handler, email string) *http.Cookie {
	t.Helper()
	cuerpo, err := json.Marshal(map[string]string{"email": email, "clave": claveTutor})
	if err != nil {
		t.Fatal(err)
	}
	respuesta := hacerSolicitud(handler, http.MethodPost, "/api/auth/ingreso", nil, bytes.NewReader(cuerpo))
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

func presentarFormularioTutor(t *testing.T, handler http.Handler, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	var cuerpo bytes.Buffer
	formulario := multipart.NewWriter(&cuerpo)
	_ = formulario.WriteField("presentacion", "Tutor sintético con presentación breve")
	_ = formulario.WriteField("cursos", "Álgebra")
	_ = formulario.WriteField("experiencia", "Experiencia sintética de tutoría")
	archivo, err := formulario.CreateFormFile("documentos", "constancia.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := archivo.Write([]byte("%PDF-1.4\ncontenido sintético")); err != nil {
		t.Fatal(err)
	}
	if err := formulario.Close(); err != nil {
		t.Fatal(err)
	}
	peticion := httptest.NewRequest(http.MethodPost, "/api/tutor/expediente", &cuerpo)
	peticion.Header.Set("Content-Type", formulario.FormDataContentType())
	peticion.Header.Set("Origin", origenTutor)
	peticion.AddCookie(cookie)
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)
	return respuesta
}

func hacerSolicitud(handler http.Handler, metodo, ruta string, cookie *http.Cookie, cuerpo *bytes.Reader) *httptest.ResponseRecorder {
	var lector *strings.Reader
	var peticion *http.Request
	if cuerpo != nil {
		peticion = httptest.NewRequest(metodo, ruta, cuerpo)
	} else {
		lector = strings.NewReader("")
		peticion = httptest.NewRequest(metodo, ruta, lector)
	}
	peticion.Header.Set("Origin", origenTutor)
	if metodo == http.MethodPost && strings.Contains(ruta, "/api/auth/ingreso") {
		peticion.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)
	return respuesta
}
