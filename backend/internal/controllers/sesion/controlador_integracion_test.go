package sesion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/controllers/acceso"
	infra "uniapp/internal/infra"
	modeloAcceso "uniapp/internal/models/acceso"
	modeloSesion "uniapp/internal/models/sesion"
)

const (
	origenI05 = "http://localhost:5173"
	claveI05  = "Clave-Segura-2026!"
)

type entornoI05 struct {
	handler   http.Handler
	tutorID   string
	escolarID string
}

func prepararEntornoI05(t *testing.T) entornoI05 {
	t.Helper()

	ctx := context.Background()

	repoAcceso := modeloAcceso.NuevoRepositorioMemoria()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}

	servicioAcceso := modeloAcceso.NuevoServicio(
		modeloAcceso.Dependencias{
			Repositorio: repoAcceso,
			Fabrica: modeloAcceso.NuevaFabrica(
				hasher,
				"uni.edu.pe",
			),
			Hasher: hasher,
			Bus:    infra.NuevoBus(),
		},
	)

	err := modeloAcceso.Sembrar(
		ctx,
		servicioAcceso,
		modeloAcceso.OpcionesSiembra{
			DominioTutor: "uni.edu.pe",
			Clave:        claveI05,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	tutor, err := repoAcceso.UsuarioPorEmail(
		ctx,
		"tutor.prueba@uni.edu.pe",
	)
	if err != nil {
		t.Fatal(err)
	}

	escolar, err := repoAcceso.UsuarioPorEmail(
		ctx,
		"escolar.prueba@example.test",
	)
	if err != nil {
		t.Fatal(err)
	}

	repoSesion := modeloSesion.NuevoRepositorioMemoria()

	inicio := time.Date(
		2026, 10, 20, 15, 0, 0, 0, time.UTC,
	)

	s, err := modeloSesion.NuevaSesion(
		"sesion-i05",
		"reserva-i05",
		tutor.ID,
		escolar.ID,
		inicio,
		inicio.Add(modeloSesion.DuracionSesion),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := repoSesion.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}

	controladorAcceso := ctrlAcceso.Nuevo(
		servicioAcceso,
		false,
	)

	mux := http.NewServeMux()

	mux.Handle(
		"/api/sesion/",
		Rutas(
			repoSesion,
			controladorAcceso.Autenticar(),
			origenI05,
		),
	)

	mux.Handle("/", controladorAcceso.Rutas())

	handler := ctrlAcceso.Encadenar(
		mux,
		ctrlAcceso.OrigenPermitido(origenI05),
	)

	return entornoI05{
		handler:   handler,
		tutorID:   tutor.ID,
		escolarID: escolar.ID,
	}
}

func iniciarSesionI05(
	t *testing.T,
	handler http.Handler,
	email string,
) *http.Cookie {
	t.Helper()

	cuerpo, err := json.Marshal(map[string]string{
		"email": email,
		"clave": claveI05,
	})
	if err != nil {
		t.Fatal(err)
	}

	peticion := httptest.NewRequest(
		http.MethodPost,
		"/api/auth/ingreso",
		strings.NewReader(string(cuerpo)),
	)
	peticion.Header.Set("Content-Type", "application/json")
	peticion.Header.Set("Origin", origenI05)

	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusOK {
		t.Fatalf(
			"ingreso de %s: %d %s",
			email,
			respuesta.Code,
			respuesta.Body.String(),
		)
	}

	for _, cookie := range respuesta.Result().Cookies() {
		if cookie.Name == "uniapp_sesion" {
			return cookie
		}
	}

	t.Fatal("no se recibio cookie de autenticacion")
	return nil
}

func solicitarI05(
	handler http.Handler,
	metodo string,
	ruta string,
	cookie *http.Cookie,
	origen string,
) *httptest.ResponseRecorder {
	peticion := httptest.NewRequest(metodo, ruta, nil)

	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	if origen != "" {
		peticion.Header.Set("Origin", origen)
	}

	respuesta := httptest.NewRecorder()
	handler.ServeHTTP(respuesta, peticion)

	return respuesta
}

func TestI05AutenticacionYConsulta(t *testing.T) {
	entorno := prepararEntornoI05(t)

	tutor := iniciarSesionI05(
		t, entorno.handler, "tutor.prueba@uni.edu.pe",
	)
	escolar := iniciarSesionI05(
		t, entorno.handler, "escolar.prueba@example.test",
	)
	apoderado := iniciarSesionI05(
		t, entorno.handler, "apoderado.prueba@example.test",
	)

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
		t.Run(caso.nombre, func(t *testing.T) {
			respuesta := solicitarI05(
				entorno.handler,
				http.MethodGet,
				ruta,
				caso.cookie,
				"",
			)

			if respuesta.Code != caso.estado {
				t.Fatalf(
					"esperado %d, obtenido %d: %s",
					caso.estado,
					respuesta.Code,
					respuesta.Body.String(),
				)
			}
		})
	}

	respuesta := solicitarI05(
		entorno.handler,
		http.MethodGet,
		"/api/sesion/inexistente",
		tutor,
		"",
	)

	if respuesta.Code != http.StatusNotFound {
		t.Fatalf(
			"sesion inexistente: %d",
			respuesta.Code,
		)
	}
}

func TestI05DeclaracionAsistencia(t *testing.T) {
	entorno := prepararEntornoI05(t)

	tutor := iniciarSesionI05(
		t, entorno.handler, "tutor.prueba@uni.edu.pe",
	)
	escolar := iniciarSesionI05(
		t, entorno.handler, "escolar.prueba@example.test",
	)
	apoderado := iniciarSesionI05(
		t, entorno.handler, "apoderado.prueba@example.test",
	)

	ruta := "/api/sesion/sesion-i05/asistencias"

	comprobar := func(
		nombre string,
		cookie *http.Cookie,
		origen string,
		esperado int,
	) {
		t.Helper()

		respuesta := solicitarI05(
			entorno.handler,
			http.MethodPost,
			ruta,
			cookie,
			origen,
		)

		if respuesta.Code != esperado {
			t.Fatalf(
				"%s: esperado %d, obtenido %d: %s",
				nombre,
				esperado,
				respuesta.Code,
				respuesta.Body.String(),
			)
		}
	}

	comprobar(
		"sin autenticacion",
		nil,
		origenI05,
		http.StatusUnauthorized,
	)

	comprobar(
		"origen no permitido",
		tutor,
		"https://otro.example",
		http.StatusForbidden,
	)

	comprobar(
		"usuario ajeno",
		apoderado,
		origenI05,
		http.StatusForbidden,
	)

	comprobar(
		"tutor declara",
		tutor,
		origenI05,
		http.StatusCreated,
	)

	comprobar(
		"tutor duplicado",
		tutor,
		origenI05,
		http.StatusConflict,
	)

	comprobar(
		"escolar declara",
		escolar,
		origenI05,
		http.StatusCreated,
	)

	respuesta := solicitarI05(
		entorno.handler,
		http.MethodGet,
		"/api/sesion/sesion-i05",
		tutor,
		"",
	)

	if respuesta.Code != http.StatusOK {
		t.Fatalf("consulta final: %d", respuesta.Code)
	}

	var resultado struct {
		Asistencias []struct {
			UsuarioID string `json:"usuario_id"`
		} `json:"asistencias"`
	}

	if err := json.Unmarshal(
		respuesta.Body.Bytes(),
		&resultado,
	); err != nil {
		t.Fatal(err)
	}

	if len(resultado.Asistencias) != 2 {
		t.Fatalf(
			"se esperaban 2 asistencias, se obtuvieron %d",
			len(resultado.Asistencias),
		)
	}
}

func TestI05DeclaracionSinOrigin(t *testing.T) {
	entorno := prepararEntornoI05(t)

	tutor := iniciarSesionI05(
		t,
		entorno.handler,
		"tutor.prueba@uni.edu.pe",
	)

	respuesta := solicitarI05(
		entorno.handler,
		http.MethodPost,
		"/api/sesion/sesion-i05/asistencias",
		tutor,
		"",
	)

	if respuesta.Code != http.StatusCreated {
		t.Fatalf(
			"esperado 201 sin Origin, obtenido %d: %s",
			respuesta.Code,
			respuesta.Body.String(),
		)
	}
}

func TestI05ConsultaSesionPorReserva(t *testing.T) {
	entorno := prepararEntornoI05(t)

	tutor := iniciarSesionI05(
		t,
		entorno.handler,
		"tutor.prueba@uni.edu.pe",
	)

	escolar := iniciarSesionI05(
		t,
		entorno.handler,
		"escolar.prueba@example.test",
	)

	apoderado := iniciarSesionI05(
		t,
		entorno.handler,
		"apoderado.prueba@example.test",
	)

	ruta := "/api/sesion/por-reserva/reserva-i05"

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
		t.Run(caso.nombre, func(t *testing.T) {
			respuesta := solicitarI05(
				entorno.handler,
				http.MethodGet,
				ruta,
				caso.cookie,
				"",
			)

			if respuesta.Code != caso.estado {
				t.Fatalf(
					"esperado %d, obtenido %d: %s",
					caso.estado,
					respuesta.Code,
					respuesta.Body.String(),
				)
			}

			if caso.estado == http.StatusOK {
				var resultado struct {
					ID        string `json:"id"`
					ReservaID string `json:"reserva_id"`
				}

				if err := json.Unmarshal(
					respuesta.Body.Bytes(),
					&resultado,
				); err != nil {
					t.Fatal(err)
				}

				if resultado.ID != "sesion-i05" ||
					resultado.ReservaID != "reserva-i05" {
					t.Fatalf(
						"sesion incorrecta: %+v",
						resultado,
					)
				}
			}
		})
	}

	t.Run("reserva inexistente", func(t *testing.T) {
		respuesta := solicitarI05(
			entorno.handler,
			http.MethodGet,
			"/api/sesion/por-reserva/reserva-inexistente",
			tutor,
			"",
		)

		if respuesta.Code != http.StatusNotFound {
			t.Fatalf(
				"esperado 404, obtenido %d: %s",
				respuesta.Code,
				respuesta.Body.String(),
			)
		}
	})
}
