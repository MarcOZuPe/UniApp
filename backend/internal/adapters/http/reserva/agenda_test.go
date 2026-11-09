package reserva

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloReserva "uniapp/internal/application/reserva"
	"uniapp/internal/infrastructure"
)

// Repositorio espia: permite verificar que la agenda
// consulta las reservas del usuario autenticado.
type repositorioAgendaPrueba struct {
	escolarConsultado string
	tutorConsultado   string
	errorConsulta     error
}

func (r *repositorioAgendaPrueba) PorEscolar(
	_ context.Context,
	escolarID string,
) ([]modeloReserva.Reserva, error) {
	r.escolarConsultado = escolarID

	if r.errorConsulta != nil {
		return nil, r.errorConsulta
	}

	return []modeloReserva.Reserva{
		{
			ID:        "reserva-escolar",
			EscolarID: escolarID,
			Curso:     "Matematicas",
			Estado:    modeloReserva.EstadoConfirmada,
			Inicio:    time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC),
			Fin:       time.Date(2026, 10, 30, 16, 0, 0, 0, time.UTC),
		},
	}, nil
}

func (r *repositorioAgendaPrueba) PorTutor(
	_ context.Context,
	tutorID string,
) ([]modeloReserva.Reserva, error) {
	r.tutorConsultado = tutorID

	if r.errorConsulta != nil {
		return nil, r.errorConsulta
	}

	return []modeloReserva.Reserva{
		{
			ID:      "reserva-tutor",
			TutorID: tutorID,
			Curso:   "Fisica",
			Estado:  modeloReserva.EstadoConfirmada,
			Inicio:  time.Date(2026, 10, 31, 15, 0, 0, 0, time.UTC),
			Fin:     time.Date(2026, 10, 31, 16, 0, 0, 0, time.UTC),
		},
	}, nil
}

func prepararAgenda(t *testing.T) (
	http.Handler,
	*modeloAcceso.Servicio,
	*repositorioAgendaPrueba,
) {
	t.Helper()

	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}

	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: memoriaAcceso.NuevoRepositorioMemoria(),
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})

	err := modeloAcceso.Sembrar(
		context.Background(),
		servicioAcceso,
		modeloAcceso.DatosSiembra{
			DominioTutor: "uni.edu.pe",
			Clave:        clave,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	repositorio := &repositorioAgendaPrueba{}

	rutasAgenda := RutasAgenda(
		repositorio,
		controladorAcceso.Autenticar(),
	)

	raiz := http.NewServeMux()
	raiz.Handle("/api/agenda", rutasAgenda)
	raiz.Handle("/", controladorAcceso.Rutas())

	return raiz, servicioAcceso, repositorio
}

func TestAgendaEscolarConsultaSuIdentidad(t *testing.T) {
	manejador, servicioAcceso, repositorio := prepararAgenda(t)

	cookie := ingresar(
		t, manejador, "escolar.prueba@example.test",
	)

	usuario, err := servicioAcceso.Sesion.UsuarioDeSesion(
		context.Background(), cookie.Value,
	)
	if err != nil {
		t.Fatal(err)
	}

	respuesta := llamar(
		manejador,
		http.MethodGet,
		"/api/agenda",
		"",
		cookie,
	)

	if respuesta.Code != http.StatusOK {
		t.Fatalf("agenda escolar: %d %s", respuesta.Code, respuesta.Body)
	}

	if repositorio.escolarConsultado != usuario.ID {
		t.Fatalf(
			"Se consulto escolar %q; esperado %q",
			repositorio.escolarConsultado,
			usuario.ID,
		)
	}

	if repositorio.tutorConsultado != "" {
		t.Fatal("La agenda escolar no debe consultar como tutor")
	}

	cuerpo := decodificar(t, respuesta)
	reservas, ok := cuerpo["reservas"].([]any)

	if !ok || len(reservas) != 1 {
		t.Fatalf("Reservas inesperadas: %v", cuerpo)
	}
}

func TestAgendaTutorConsultaSuIdentidad(t *testing.T) {
	manejador, servicioAcceso, repositorio := prepararAgenda(t)

	cookie := ingresar(
		t, manejador, "tutor.prueba@uni.edu.pe",
	)

	usuario, err := servicioAcceso.Sesion.UsuarioDeSesion(
		context.Background(), cookie.Value,
	)
	if err != nil {
		t.Fatal(err)
	}

	respuesta := llamar(
		manejador,
		http.MethodGet,
		"/api/agenda",
		"",
		cookie,
	)

	if respuesta.Code != http.StatusOK {
		t.Fatalf("agenda tutor: %d %s", respuesta.Code, respuesta.Body)
	}

	if repositorio.tutorConsultado != usuario.ID {
		t.Fatalf(
			"Se consulto tutor %q; esperado %q",
			repositorio.tutorConsultado,
			usuario.ID,
		)
	}

	if repositorio.escolarConsultado != "" {
		t.Fatal("La agenda tutor no debe consultar como escolar")
	}

	cuerpo := decodificar(t, respuesta)
	reservas, ok := cuerpo["reservas"].([]any)

	if !ok || len(reservas) != 1 {
		t.Fatalf("Reservas inesperadas: %v", cuerpo)
	}
}

func TestAgendaSinSesionEsRechazada(t *testing.T) {
	manejador, _, repositorio := prepararAgenda(t)

	respuesta := llamar(
		manejador,
		http.MethodGet,
		"/api/agenda",
		"",
		nil,
	)

	if respuesta.Code != http.StatusUnauthorized {
		t.Fatalf(
			"Esperado 401, recibido %d: %s",
			respuesta.Code,
			respuesta.Body,
		)
	}

	if repositorio.escolarConsultado != "" ||
		repositorio.tutorConsultado != "" {
		t.Fatal("Una peticion sin sesion llego al repositorio")
	}
}

func TestAgendaCoordinadorNoAutorizado(t *testing.T) {
	manejador, _, repositorio := prepararAgenda(t)

	cookie := ingresar(
		t, manejador, "coord.prueba@example.test",
	)

	respuesta := llamar(
		manejador,
		http.MethodGet,
		"/api/agenda",
		"",
		cookie,
	)

	if respuesta.Code != http.StatusForbidden {
		t.Fatalf(
			"Esperado 403, recibido %d: %s",
			respuesta.Code,
			respuesta.Body,
		)
	}

	if repositorio.escolarConsultado != "" ||
		repositorio.tutorConsultado != "" {
		t.Fatal("Un rol no autorizado llego al repositorio")
	}
}

func TestAgendaFalloRepositorio(t *testing.T) {
	manejador, _, repositorio := prepararAgenda(t)

	repositorio.errorConsulta = errors.New("fallo simulado")

	cookie := ingresar(
		t, manejador, "escolar.prueba@example.test",
	)

	respuesta := llamar(
		manejador,
		http.MethodGet,
		"/api/agenda",
		"",
		cookie,
	)

	if respuesta.Code != http.StatusInternalServerError {
		t.Fatalf(
			"Esperado 500, recibido %d: %s",
			respuesta.Code,
			respuesta.Body,
		)
	}

	cuerpo := decodificar(t, respuesta)

	if cuerpo["codigo"] != "error_agenda" {
		t.Fatalf("Codigo inesperado: %v", cuerpo)
	}

	if respuesta.Body.String() == "fallo simulado" {
		t.Fatal("No se deben exponer errores internos")
	}
}

// Comprobacion adicional del metodo HTTP.
func TestAgendaMetodoNoPermitido(t *testing.T) {
	manejador, _, _ := prepararAgenda(t)

	peticion := httptest.NewRequest(
		http.MethodPost,
		"/api/agenda",
		nil,
	)
	respuesta := httptest.NewRecorder()

	manejador.ServeHTTP(respuesta, peticion)

	if respuesta.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"Esperado 405, recibido %d",
			respuesta.Code,
		)
	}
}
