package sesion

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
	memoriaSesion "uniapp/internal/adapters/repositories/memory/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloSesion "uniapp/internal/application/sesion"
	infra "uniapp/internal/infrastructure"
)

const (
	origenI05 = "http://localhost:5173"
	claveI05  = "Clave-Segura-2026!"
)

func TestI05AutenticacionYConsulta(testingT *testing.T) {
	entorno := prepararEntornoI05(testingT)

	tutor := iniciarSesionI05(testingT, entorno.manejador, "tutor.prueba@uni.edu.pe")
	escolar := iniciarSesionI05(testingT, entorno.manejador, "escolar.prueba@example.test")
	apoderado := iniciarSesionI05(testingT, entorno.manejador, "apoderado.prueba@example.test")

	ruta := "/api/sesion/sesion-i05"

	casos := []struct {
		nombre string
		cookie *http.Cookie
		estado int
	}{
		{"sin autenticacion", nil, http.StatusUnauthorized},
		{"consulta tutor", tutor, http.StatusOK},
		{"consulta escolar", escolar, http.StatusOK},
		{"usuario ajeno", apoderado, http.StatusForbidden},
	}

	for _, caso := range casos {
		testingT.Run(caso.nombre, func(testingT *testing.T) {
			respuesta := solicitarI05(opcionesSolicitudI05{
				manejador: entorno.manejador,
				metodo:    http.MethodGet,
				ruta:      ruta,
				cookie:    caso.cookie,
			})

			if respuesta.Code != caso.estado {
				testingT.Fatalf("esperado %d, obtenido %d: %s",
					caso.estado, respuesta.Code, respuesta.Body.String())
			}
		})
	}

	respuesta := solicitarI05(opcionesSolicitudI05{
		manejador: entorno.manejador,
		metodo:    http.MethodGet,
		ruta:      "/api/sesion/inexistente",
		cookie:    tutor,
	})
	if respuesta.Code != http.StatusNotFound {
		testingT.Fatalf("sesion inexistente: %d", respuesta.Code)
	}
}

func TestI05DeclaracionAsistencia(testingT *testing.T) {
	entorno := prepararEntornoI05(testingT)

	tutor := iniciarSesionI05(testingT, entorno.manejador, "tutor.prueba@uni.edu.pe")
	escolar := iniciarSesionI05(testingT, entorno.manejador, "escolar.prueba@example.test")
	apoderado := iniciarSesionI05(testingT, entorno.manejador, "apoderado.prueba@example.test")

	ruta := "/api/sesion/sesion-i05/asistencias"

	comprobar := func(caso comprobacionAsistencia) {
		testingT.Helper()

		respuesta := solicitarI05(opcionesSolicitudI05{
			manejador: entorno.manejador,
			metodo:    http.MethodPost,
			ruta:      ruta,
			cookie:    caso.cookie,
			origen:    caso.origen,
		})
		if respuesta.Code != caso.esperado {
			testingT.Fatalf("%s: esperado %d, obtenido %d: %s",
				caso.nombre, caso.esperado, respuesta.Code, respuesta.Body.String())
		}
	}

	comprobar(comprobacionAsistencia{
		nombre: "sin autenticacion", origen: origenI05, esperado: http.StatusUnauthorized,
	})
	comprobar(comprobacionAsistencia{
		nombre: "origen no permitido", cookie: tutor, origen: "https://otro.example", esperado: http.StatusForbidden,
	})
	comprobar(comprobacionAsistencia{
		nombre: "usuario ajeno", cookie: apoderado, origen: origenI05, esperado: http.StatusForbidden,
	})
	comprobar(comprobacionAsistencia{
		nombre: "tutor declara", cookie: tutor, origen: origenI05, esperado: http.StatusCreated,
	})
	comprobar(comprobacionAsistencia{
		nombre: "tutor duplicado", cookie: tutor, origen: origenI05, esperado: http.StatusConflict,
	})
	comprobar(comprobacionAsistencia{
		nombre: "escolar declara", cookie: escolar, origen: origenI05, esperado: http.StatusCreated,
	})

	respuesta := solicitarI05(opcionesSolicitudI05{
		manejador: entorno.manejador,
		metodo:    http.MethodGet,
		ruta:      "/api/sesion/sesion-i05",
		cookie:    tutor,
	})
	if respuesta.Code != http.StatusOK {
		testingT.Fatalf("consulta final: %d", respuesta.Code)
	}

	var resultado struct {
		Asistencias []struct {
			UsuarioID string `json:"usuario_id"`
		} `json:"asistencias"`
	}
	if err := json.Unmarshal(respuesta.Body.Bytes(), &resultado); err != nil {
		testingT.Fatal(err)
	}
	if len(resultado.Asistencias) != 2 {
		testingT.Fatalf("se esperaban 2 asistencias, se obtuvieron %d", len(resultado.Asistencias))
	}
}

func TestI05DeclaracionSinOrigin(testingT *testing.T) {
	entorno := prepararEntornoI05(testingT)
	tutor := iniciarSesionI05(testingT, entorno.manejador, "tutor.prueba@uni.edu.pe")

	respuesta := solicitarI05(opcionesSolicitudI05{
		manejador: entorno.manejador,
		metodo:    http.MethodPost,
		ruta:      "/api/sesion/sesion-i05/asistencias",
		cookie:    tutor,
	})
	if respuesta.Code != http.StatusCreated {
		testingT.Fatalf("esperado 201 sin Origin, obtenido %d: %s", respuesta.Code, respuesta.Body.String())
	}
}

type entornoI05 struct {
	manejador http.Handler
	tutorID   string
	escolarID string
}

type opcionesSolicitudI05 struct {
	manejador http.Handler
	metodo    string
	ruta      string
	cookie    *http.Cookie
	origen    string
}

type comprobacionAsistencia struct {
	nombre   string
	cookie   *http.Cookie
	origen   string
	esperado int
}

func prepararEntornoI05(testingT *testing.T) entornoI05 {
	testingT.Helper()

	contexto := context.Background()
	repositorioAcceso := memoriaAcceso.NuevoRepositorioMemoria()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}

	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: repositorioAcceso,
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(contexto, servicioAcceso, modeloAcceso.DatosSiembra{
		DominioTutor: "uni.edu.pe",
		Clave:        claveI05,
	}); err != nil {
		testingT.Fatal(err)
	}

	tutor, err := repositorioAcceso.UsuarioPorEmail(contexto, "tutor.prueba@uni.edu.pe")
	if err != nil {
		testingT.Fatal(err)
	}
	escolar, err := repositorioAcceso.UsuarioPorEmail(contexto, "escolar.prueba@example.test")
	if err != nil {
		testingT.Fatal(err)
	}

	repositorioSesion := memoriaSesion.NuevoRepositorioMemoria()
	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	sesion, err := modeloSesion.NuevaSesion(
		"sesion-i05", "reserva-i05", tutor.ID, escolar.ID,
		inicio, inicio.Add(modeloSesion.DuracionSesion),
	)
	if err != nil {
		testingT.Fatal(err)
	}
	if err := repositorioSesion.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatal(err)
	}

	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	multiplexor := http.NewServeMux()
	multiplexor.Handle("/api/sesion/", Rutas(
		repositorioSesion, controladorAcceso.Autenticar(), origenI05,
	))
	multiplexor.Handle("/", controladorAcceso.Rutas())

	manejador := ctrlAcceso.Encadenar(
		multiplexor,
		ctrlAcceso.OrigenPermitido(origenI05),
	)

	return entornoI05{manejador: manejador, tutorID: tutor.ID, escolarID: escolar.ID}
}

func iniciarSesionI05(testingT *testing.T, manejador http.Handler, email string) *http.Cookie {
	testingT.Helper()

	cuerpo, err := json.Marshal(map[string]string{"email": email, "clave": claveI05})
	if err != nil {
		testingT.Fatal(err)
	}

	peticion := httptest.NewRequest(http.MethodPost, "/api/auth/ingreso", strings.NewReader(string(cuerpo)))
	peticion.Header.Set("Content-Type", "application/json")
	peticion.Header.Set("Origin", origenI05)

	respuesta := httptest.NewRecorder()
	manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		testingT.Fatalf("ingreso de %s: %d %s", email, respuesta.Code, respuesta.Body.String())
	}

	for _, cookie := range respuesta.Result().Cookies() {
		if cookie.Name == "uniapp_sesion" {
			return cookie
		}
	}
	testingT.Fatal("no se recibio cookie de autenticacion")
	return nil
}

func solicitarI05(opciones opcionesSolicitudI05) *httptest.ResponseRecorder {
	peticion := httptest.NewRequest(opciones.metodo, opciones.ruta, nil)
	if opciones.cookie != nil {
		peticion.AddCookie(opciones.cookie)
	}
	if opciones.origen != "" {
		peticion.Header.Set("Origin", opciones.origen)
	}

	respuesta := httptest.NewRecorder()
	opciones.manejador.ServeHTTP(respuesta, peticion)
	return respuesta
}
