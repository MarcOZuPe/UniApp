package incidente

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

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoria "uniapp/internal/adapters/repositories/memory/incidente"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/incidente"
	infra "uniapp/internal/infrastructure"
)

const clave = "Clave-Segura-2026!"

type almacenPrueba struct {
	archivos map[string][]byte
}

func (a *almacenPrueba) Guardar(_ context.Context, clave string, contenido []byte) error {
	a.archivos[clave] = append([]byte(nil), contenido...)
	return nil
}

func (a *almacenPrueba) Leer(_ context.Context, clave string) ([]byte, error) {
	contenido, existe := a.archivos[clave]
	if !existe {
		return nil, modelo.ErrAlmacen
	}
	return append([]byte(nil), contenido...), nil
}

func (a *almacenPrueba) Eliminar(_ context.Context, clave string) error {
	delete(a.archivos, clave)
	return nil
}

type entorno struct {
	manejador   http.Handler
	escolarID   string
	escolar     *http.Cookie
	apoderado   *http.Cookie
	tutor       *http.Cookie
	coordinador *http.Cookie
}

func nuevoEntorno(t *testing.T) *entorno {
	t.Helper()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: memoriaAcceso.NuevoRepositorioMemoria(),
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(context.Background(), servicioAcceso, modeloAcceso.DatosSiembra{DominioTutor: "uni.edu.pe", Clave: clave}); err != nil {
		t.Fatal(err)
	}
	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)

	escolarID := idDe(t, servicioAcceso, modeloAcceso.EmailEscolarSintetico)
	repositorio := memoria.NuevoRepositorioMemoria()
	ahora := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	repositorio.SembrarTutoria(modelo.Tutoria{
		ReservaID: "reserva-1", EscolarID: escolarID,
		EscolarNombre: "Escolar Prueba", TutorID: "tutor-1", TutorNombre: "Tutor Prueba",
		Curso: "Álgebra", Inicio: ahora.Add(-2 * time.Hour), Fin: ahora.Add(-time.Hour), Asistio: true,
	})
	repositorio.SembrarVinculacion(
		idDe(t, servicioAcceso, modeloAcceso.EmailApoderadoSintetico),
		modelo.EscolarVinculado{ID: escolarID, Nombre: "Escolar Prueba"},
	)
	servicio := modelo.NuevoServicio(repositorio, &almacenPrueba{archivos: map[string][]byte{}})
	servicio.FijarReloj(func() time.Time { return ahora })

	manejadorIncidente := Nuevo(servicio, controladorAcceso.Autenticar()).Rutas()
	raiz := http.NewServeMux()
	raiz.Handle("/api/incidentes", manejadorIncidente)
	raiz.Handle("/api/incidentes/", manejadorIncidente)
	raiz.Handle("/", controladorAcceso.Rutas())
	manejador := ctrlAcceso.Encadenar(raiz, ctrlAcceso.OrigenPermitido("http://localhost:5173"))

	return &entorno{
		manejador:   manejador,
		escolarID:   escolarID,
		escolar:     ingresar(t, manejador, modeloAcceso.EmailEscolarSintetico),
		apoderado:   ingresar(t, manejador, modeloAcceso.EmailApoderadoSintetico),
		tutor:       ingresar(t, manejador, "tutor.prueba@uni.edu.pe"),
		coordinador: ingresar(t, manejador, modeloAcceso.EmailCoordinadorSintetico),
	}
}

func idDe(t *testing.T, servicio *modeloAcceso.Servicio, email string) string {
	t.Helper()
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorEmail(context.Background(), email)
	if err != nil {
		t.Fatal(err)
	}
	return usuario.ID
}

func llamar(manejador http.Handler, metodo, ruta, cuerpo string, cookie *http.Cookie) *httptest.ResponseRecorder {
	peticion := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	grabador := httptest.NewRecorder()
	manejador.ServeHTTP(grabador, peticion)
	return grabador
}

func subirIncidente(
	t *testing.T,
	manejador http.Handler,
	cookie *http.Cookie,
	escolarID string,
	archivos []struct {
		nombre    string
		contenido []byte
	},
) *httptest.ResponseRecorder {
	t.Helper()
	var cuerpo bytes.Buffer
	escritor := multipart.NewWriter(&cuerpo)
	_ = escritor.WriteField("reserva_id", "reserva-1")
	_ = escritor.WriteField("tipo", string(modelo.TipoTratoInadecuado))
	_ = escritor.WriteField("descripcion", "El tutor hizo un comentario ofensivo durante la sesión.")
	if escolarID != "" {
		_ = escritor.WriteField("escolar_id", escolarID)
	}
	for _, archivo := range archivos {
		parte, err := escritor.CreateFormFile("evidencias", archivo.nombre)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = parte.Write(archivo.contenido)
	}
	if err := escritor.Close(); err != nil {
		t.Fatal(err)
	}
	peticion := httptest.NewRequest("POST", "/api/incidentes", &cuerpo)
	peticion.Header.Set("Content-Type", escritor.FormDataContentType())
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	grabador := httptest.NewRecorder()
	manejador.ServeHTTP(grabador, peticion)
	return grabador
}

func ingresar(t *testing.T, manejador http.Handler, email string) *http.Cookie {
	t.Helper()
	respuesta := llamar(manejador, "POST", "/api/auth/ingreso", `{"email":"`+email+`","clave":"`+clave+`"}`, nil)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("ingreso %s: %d %s", email, respuesta.Code, respuesta.Body)
	}
	for _, cookie := range respuesta.Result().Cookies() {
		if cookie.Name == "uniapp_sesion" {
			return cookie
		}
	}
	t.Fatal("sin cookie de sesión")
	return nil
}

func decodificar(t *testing.T, respuesta *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var cuerpo map[string]any
	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("JSON inválido (%d): %s", respuesta.Code, respuesta.Body)
	}
	return cuerpo
}

func pdf() []byte {
	return []byte("%PDF-1.4\n1 0 obj<</Type/Catalog>>endobj\ntrailer<</Root 1 0 R>>\n%%EOF\n")
}

func png() []byte {
	return append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, []byte("datos")...)
}

func TestIncidentesHTTPVisibilidadPorRolYRevision(t *testing.T) {
	entorno := nuevoEntorno(t)

	tutorias := llamar(entorno.manejador, "GET", "/api/incidentes/tutorias", "", entorno.escolar)
	if tutorias.Code != http.StatusOK || len(decodificar(t, tutorias)["tutorias"].([]any)) != 1 {
		t.Fatalf("tutorías del escolar: %d %s", tutorias.Code, tutorias.Body)
	}
	tutoriasApoderado := llamar(entorno.manejador, "GET", "/api/incidentes/tutorias", "", entorno.apoderado)
	if tutoriasApoderado.Code != http.StatusOK || len(decodificar(t, tutoriasApoderado)["tutorias"].([]any)) != 1 {
		t.Fatalf("tutorías del apoderado: %d %s", tutoriasApoderado.Code, tutoriasApoderado.Body)
	}

	sinEvidencia := subirIncidente(t, entorno.manejador, entorno.escolar, "", nil)
	if sinEvidencia.Code != http.StatusUnprocessableEntity {
		t.Fatalf("sin evidencia: %d %s", sinEvidencia.Code, sinEvidencia.Body)
	}
	invalida := subirIncidente(t, entorno.manejador, entorno.escolar, "", []struct {
		nombre    string
		contenido []byte
	}{{"notas.txt", []byte("texto")}})
	if invalida.Code != http.StatusUnprocessableEntity {
		t.Fatalf("evidencia inválida: %d %s", invalida.Code, invalida.Body)
	}
	prohibido := subirIncidente(t, entorno.manejador, entorno.tutor, "", []struct {
		nombre    string
		contenido []byte
	}{{"captura.pdf", pdf()}})
	if prohibido.Code != http.StatusForbidden {
		t.Fatalf("el tutor no debe reportar: %d %s", prohibido.Code, prohibido.Body)
	}
	coordinaReporte := subirIncidente(t, entorno.manejador, entorno.coordinador, "", []struct {
		nombre    string
		contenido []byte
	}{{"captura.pdf", pdf()}})
	if coordinaReporte.Code != http.StatusForbidden {
		t.Fatalf("coordinación no debe reportar: %d %s", coordinaReporte.Code, coordinaReporte.Body)
	}

	creado := subirIncidente(t, entorno.manejador, entorno.escolar, "", []struct {
		nombre    string
		contenido []byte
	}{{"captura.pdf", pdf()}, {"foto.png", png()}})
	if creado.Code != http.StatusCreated {
		t.Fatalf("crear incidente: %d %s", creado.Code, creado.Body)
	}
	incidente := decodificar(t, creado)
	id, _ := incidente["id"].(string)
	evidencias, _ := incidente["evidencias"].([]any)
	if id == "" || len(evidencias) != 2 || incidente["estado"] != "pendiente" {
		t.Fatalf("incidente creado: %v", incidente)
	}

	creadoApoderado := subirIncidente(
		t, entorno.manejador, entorno.apoderado, entorno.escolarID,
		[]struct {
			nombre    string
			contenido []byte
		}{{"captura.pdf", pdf()}},
	)
	if creadoApoderado.Code != http.StatusCreated {
		t.Fatalf("reporte del apoderado: %d %s", creadoApoderado.Code, creadoApoderado.Body)
	}

	listado := llamar(entorno.manejador, "GET", "/api/incidentes", "", entorno.coordinador)
	if listado.Code != http.StatusOK || len(decodificar(t, listado)["incidentes"].([]any)) != 2 {
		t.Fatalf("listado de coordinación: %d %s", listado.Code, listado.Body)
	}
	if listado.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("el listado confidencial debe evitar caché: %q", listado.Header().Get("Cache-Control"))
	}
	if respuesta := llamar(entorno.manejador, "GET", "/api/incidentes", "", entorno.tutor); respuesta.Code != http.StatusForbidden {
		t.Fatalf("el tutor no debe listar: %d %s", respuesta.Code, respuesta.Body)
	}

	evidenciaID := evidencias[0].(map[string]any)["id"].(string)
	descargaTutor := llamar(entorno.manejador, "GET", "/api/incidentes/"+id+"/evidencias/"+evidenciaID, "", entorno.tutor)
	if descargaTutor.Code != http.StatusForbidden {
		t.Fatalf("el tutor no debe descargar: %d %s", descargaTutor.Code, descargaTutor.Body)
	}
	descargaEscolar := llamar(entorno.manejador, "GET", "/api/incidentes/"+id+"/evidencias/"+evidenciaID, "", entorno.escolar)
	if descargaEscolar.Code != http.StatusOK || !bytes.Equal(descargaEscolar.Body.Bytes(), pdf()) {
		t.Fatalf("descarga del escolar: %d %s", descargaEscolar.Code, descargaEscolar.Body)
	}
	descargaApoderado := llamar(entorno.manejador, "GET", "/api/incidentes/"+id+"/evidencias/"+evidenciaID, "", entorno.apoderado)
	if descargaApoderado.Code != http.StatusOK {
		t.Fatalf("descarga del apoderado vinculado: %d %s", descargaApoderado.Code, descargaApoderado.Body)
	}
	descargaCoordinacion := llamar(entorno.manejador, "GET", "/api/incidentes/"+id+"/evidencias/"+evidenciaID, "", entorno.coordinador)
	if descargaCoordinacion.Code != http.StatusOK {
		t.Fatalf("descarga de coordinación: %d %s", descargaCoordinacion.Code, descargaCoordinacion.Body)
	}

	if respuesta := llamar(entorno.manejador, "POST", "/api/incidentes/"+id+"/revision", `{"resultado":"procede","comentario":"Se revisó la evidencia."}`, entorno.escolar); respuesta.Code != http.StatusForbidden {
		t.Fatalf("el escolar no debe revisar: %d %s", respuesta.Code, respuesta.Body)
	}
	revision := llamar(entorno.manejador, "POST", "/api/incidentes/"+id+"/revision", `{"resultado":"procede","comentario":"Se revisó la evidencia."}`, entorno.coordinador)
	if revision.Code != http.StatusOK {
		t.Fatalf("revisión: %d %s", revision.Code, revision.Body)
	}
	if decodificar(t, revision)["estado"] != "revisado" {
		t.Fatalf("estado revisado: %s", revision.Body)
	}
	segunda := llamar(entorno.manejador, "POST", "/api/incidentes/"+id+"/revision", `{"resultado":"no_procede","comentario":"Otra revisión."}`, entorno.coordinador)
	if segunda.Code != http.StatusConflict {
		t.Fatalf("segunda revisión: %d %s", segunda.Code, segunda.Body)
	}
}
