package sesion

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoriaSesion "uniapp/internal/adapters/repositories/memory/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloSesion "uniapp/internal/application/sesion"
	infra "uniapp/internal/infrastructure"
)

// nombresDeAcceso adapta el repositorio de acceso al puerto DirectorioNombres.
type nombresDeAcceso struct {
	repo *memoriaAcceso.RepositorioMemoria
}

func (n nombresDeAcceso) NombreDe(ctx context.Context, id string) (string, error) {
	usuario, err := n.repo.UsuarioPorID(ctx, id)
	if err != nil {
		return "", nil
	}
	return usuario.NombreUso, nil
}

// servicioSinConstancia simula que la constancia no se pudo guardar.
type servicioSinConstancia struct{ ServicioReporte }

func (servicioSinConstancia) RegistrarDescarga(context.Context, *modeloSesion.Reporte, modeloSesion.FormatoReporte, []byte) (modeloSesion.RegistroReporte, error) {
	return modeloSesion.RegistroReporte{}, errors.New("base caida")
}

type entornoI18 struct {
	manejador http.Handler
	escolarID string
}

// prepararEntornoI18: una sesion con 45 minutos verificados (20/10 en Lima).
func prepararEntornoI18(t *testing.T, envolver func(ServicioReporte) ServicioReporte) entornoI18 {
	t.Helper()
	ctx := context.Background()
	repositorioAcceso := memoriaAcceso.NuevoRepositorioMemoria()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: repositorioAcceso,
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(ctx, servicioAcceso, modeloAcceso.DatosSiembra{DominioTutor: "uni.edu.pe", Clave: claveI05}); err != nil {
		t.Fatal(err)
	}
	tutor, _ := repositorioAcceso.UsuarioPorEmail(ctx, "tutor.prueba@uni.edu.pe")
	escolar, _ := repositorioAcceso.UsuarioPorEmail(ctx, "escolar.prueba@example.test")

	sesiones := memoriaSesion.NuevoRepositorioMemoria()
	s, err := modeloSesion.NuevaSesion("sesion-i18", "reserva-i18", tutor.ID, escolar.ID, inicioSesionI05, inicioSesionI05.Add(modeloSesion.DuracionSesion))
	if err != nil {
		t.Fatal(err)
	}
	if err := sesiones.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	for _, u := range []string{tutor.ID, escolar.ID} {
		if _, err := sesiones.DeclararAsistencia(ctx, s.ID(), u, inicioSesionI05); err != nil {
			t.Fatal(err)
		}
	}
	ahora := inicioSesionI05.Add(2 * time.Hour)
	minutos := modeloSesion.ServicioMinutos{Sesiones: sesiones, Minutos: memoriaSesion.NuevoRepositorioMinutosMemoria(), Reservas: memoriaSesion.NuevasCancelacionesFijas()}
	if _, err := minutos.Proponer(ctx, s.ID(), tutor.ID, 45, ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := minutos.Aprobar(ctx, s.ID(), "coord-x", ahora); err != nil {
		t.Fatal(err)
	}
	seudonimos, _ := modeloSesion.NuevoSeudonimizador([]byte("clave-de-prueba-i18-suficiente"))
	var servicio ServicioReporte = modeloSesion.ServicioReporte{
		Reportes:   memoriaSesion.NuevoRepositorioReportesMemoria(sesiones),
		Minutos:    minutos,
		Nombres:    nombresDeAcceso{repo: repositorioAcceso},
		Seudonimos: seudonimos,
	}
	if envolver != nil {
		servicio = envolver(servicio)
	}

	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	mux := http.NewServeMux()
	mux.Handle("/api/reportes/", RutasReporte(servicio, controladorAcceso.Autenticar(), func() time.Time { return ahora }))
	mux.Handle("/", controladorAcceso.Rutas())
	return entornoI18{manejador: ctrlAcceso.Encadenar(mux, ctrlAcceso.OrigenPermitido(origenI05)), escolarID: escolar.ID}
}

func pedirI18(t *testing.T, e entornoI18, cookie *http.Cookie, ruta string, esperado int) *httptest.ResponseRecorder {
	t.Helper()
	peticion := httptest.NewRequest("GET", ruta, nil)
	peticion.Header.Set("Origin", origenI05)
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	respuesta := httptest.NewRecorder()
	e.manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != esperado {
		t.Fatalf("GET %s: esperado %d, recibido %d: %s", ruta, esperado, respuesta.Code, respuesta.Body.String())
	}
	return respuesta
}

const rutaReporteI18 = "/api/reportes/minutos?desde=2026-10-01&hasta=2026-10-31"

func TestI18SoloCoordinacionDescarga(t *testing.T) {
	e := prepararEntornoI18(t, nil)
	pedirI18(t, e, nil, rutaReporteI18+"&formato=csv", 401)
	for _, email := range []string{"tutor.prueba@uni.edu.pe", "escolar.prueba@example.test", "apoderado.prueba@example.test"} {
		cookie := iniciarSesionI05(t, e.manejador, email)
		pedirI18(t, e, cookie, rutaReporteI18+"&formato=csv", 403)
		pedirI18(t, e, cookie, "/api/reportes/registro", 403)
	}
}

func TestI18ValidaParametros(t *testing.T) {
	e := prepararEntornoI18(t, nil)
	coord := iniciarSesionI05(t, e.manejador, "coord.prueba@example.test")
	casos := map[string]int{
		rutaReporteI18:                                                        400, // falta formato
		rutaReporteI18 + "&formato=xlsx":                                      400,
		rutaReporteI18 + "&formato=csv&extra=1":                               400,
		rutaReporteI18 + "&formato=csv&formato=pdf":                           400,
		"/api/reportes/minutos?desde=01-10-2026&hasta=2026-10-31&formato=csv": 400,
		"/api/reportes/minutos?desde=2026-10-31&hasta=2026-10-01&formato=csv": 422,
		"/api/reportes/minutos?desde=2026-10-01&hasta=2026-11-01&formato=csv": 422, // 32 dias
	}
	for ruta, esperado := range casos {
		pedirI18(t, e, coord, ruta, esperado)
	}
	// Un solo dia (desde == hasta) es valido.
	pedirI18(t, e, coord, "/api/reportes/minutos?desde=2026-10-20&hasta=2026-10-20&formato=csv", 200)
}

func TestI18DescargaCSVConConstancia(t *testing.T) {
	e := prepararEntornoI18(t, nil)
	coord := iniciarSesionI05(t, e.manejador, "coord.prueba@example.test")
	respuesta := pedirI18(t, e, coord, rutaReporteI18+"&formato=csv", 200)

	cuerpo := respuesta.Body.Bytes()
	huella := sha256.Sum256(cuerpo)
	cabecera := respuesta.Header()
	if cabecera.Get("X-Reporte-Sha256") != hex.EncodeToString(huella[:]) {
		t.Fatal("la huella de la cabecera debe ser la del archivo entregado")
	}
	if !strings.HasPrefix(cabecera.Get("Content-Type"), "text/csv") || cabecera.Get("Cache-Control") != "no-store" {
		t.Fatalf("cabeceras inesperadas: %v", cabecera)
	}
	if cabecera.Get("Content-Disposition") != `attachment; filename="reporte-minutos-20261001-20261031.csv"` {
		t.Fatalf("nombre de archivo inesperado: %s", cabecera.Get("Content-Disposition"))
	}
	texto := string(cuerpo)
	if strings.Contains(texto, e.escolarID) {
		t.Fatal("el reporte no puede incluir el ID del escolar")
	}
	if !strings.Contains(texto, "sesion-i18") || !strings.Contains(texto, ",45,45,") || !strings.Contains(texto, "Sin validez oficial") {
		t.Fatalf("contenido inesperado:\n%s", texto)
	}

	registro := pedirI18(t, e, coord, "/api/reportes/registro", 200)
	var datos struct {
		Reportes []struct {
			ID      string `json:"id"`
			Formato string `json:"formato"`
			Filas   int    `json:"filas"`
			SHA256  string `json:"sha256"`
		} `json:"reportes"`
	}
	if err := json.Unmarshal(registro.Body.Bytes(), &datos); err != nil {
		t.Fatal(err)
	}
	if len(datos.Reportes) != 1 || datos.Reportes[0].SHA256 != cabecera.Get("X-Reporte-Sha256") ||
		datos.Reportes[0].ID != cabecera.Get("X-Reporte-Id") || datos.Reportes[0].Filas != 1 || datos.Reportes[0].Formato != "csv" {
		t.Fatalf("la constancia no coincide con la descarga: %+v", datos.Reportes)
	}
}

func TestI18DescargaPDF(t *testing.T) {
	e := prepararEntornoI18(t, nil)
	coord := iniciarSesionI05(t, e.manejador, "coord.prueba@example.test")
	respuesta := pedirI18(t, e, coord, rutaReporteI18+"&formato=pdf", 200)
	if respuesta.Header().Get("Content-Type") != "application/pdf" || !strings.HasPrefix(respuesta.Body.String(), "%PDF-1.4") {
		t.Fatal("se esperaba un PDF")
	}
	if strings.Contains(respuesta.Body.String(), e.escolarID) {
		t.Fatal("el PDF no puede incluir el ID del escolar")
	}
}

func TestI18SinConstanciaNoSeEntrega(t *testing.T) {
	e := prepararEntornoI18(t, func(s ServicioReporte) ServicioReporte { return servicioSinConstancia{s} })
	coord := iniciarSesionI05(t, e.manejador, "coord.prueba@example.test")
	respuesta := pedirI18(t, e, coord, rutaReporteI18+"&formato=csv", 500)
	if strings.Contains(respuesta.Body.String(), "sesion-i18") || respuesta.Header().Get("X-Reporte-Sha256") != "" {
		t.Fatal("si la constancia falla no debe entregarse el archivo")
	}
}
