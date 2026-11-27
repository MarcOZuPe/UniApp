package sesion

import (
	"errors"
	"testing"
	"time"
)

func sesionDePrueba(t *testing.T) *Sesion {
	t.Helper()
	inicio := time.Now().UTC().Add(-time.Hour)
	sesion, err := NuevaSesion("s-1", "r-1", "tutor-1", "escolar-1", inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatalf("no se creó la sesión: %v", err)
	}
	return sesion
}

func TestNuevaSesionValida(t *testing.T) {
	sesion := sesionDePrueba(t)
	if sesion.ID() != "s-1" || sesion.TutorID() != "tutor-1" || sesion.EscolarID() != "escolar-1" {
		t.Fatalf("sesión inesperada: %+v", sesion)
	}
	casos := []struct {
		id, reserva, tutor, escolar string
		inicio, fin                 time.Time
	}{
		{"", "r", "t", "e", time.Now(), time.Now().Add(time.Hour)},
		{"s", "", "t", "e", time.Now(), time.Now().Add(time.Hour)},
		{"s", "r", "", "e", time.Now(), time.Now().Add(time.Hour)},
		{"s", "r", "t", "", time.Now(), time.Now().Add(time.Hour)},
		{"s", "r", "t", "t", time.Now(), time.Now().Add(time.Hour)},
		{"s", "r", "t", "e", time.Time{}, time.Now()},
		{"s", "r", "t", "e", time.Now(), time.Now().Add(30 * time.Minute)},
	}
	for _, caso := range casos {
		if _, err := NuevaSesion(caso.id, caso.reserva, caso.tutor, caso.escolar, caso.inicio, caso.fin); err == nil {
			t.Fatalf("se esperaba sesión inválida: %+v", caso)
		}
	}
}

func TestAsistenciasYDentroDePlazo(t *testing.T) {
	sesion := sesionDePrueba(t)
	ahora := sesion.Inicio().Add(30 * time.Minute)
	if err := sesion.DeclararAsistencia("escolar-1", ahora); err != nil {
		t.Fatalf("asistencia del escolar falló: %v", err)
	}
	if err := sesion.DeclararAsistencia("tutor-1", ahora); err != nil {
		t.Fatalf("asistencia del tutor falló: %v", err)
	}
	if len(sesion.Asistencias()) != 2 {
		t.Fatalf("se esperaban 2 asistencias, se obtuvieron %d", len(sesion.Asistencias()))
	}
	if err := sesion.DeclararAsistencia("escolar-1", ahora); !errors.Is(err, ErrAsistenciaDuplicada) {
		t.Fatalf("se esperaba asistencia duplicada, se obtuvo %v", err)
	}
	if err := sesion.DeclararAsistencia("otro", ahora); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("se esperaba no autorizado, se obtuvo %v", err)
	}
	fuera := sesion.Fin().Add(PlazoDeclaracion + time.Hour)
	if err := sesion.DeclararAsistencia("escolar-1", fuera); !errors.Is(err, ErrFueraDePlazo) {
		t.Fatalf("se esperaba fuera de plazo, se obtuvo %v", err)
	}
}

func TestAusencias(t *testing.T) {
	sesion := sesionDePrueba(t)
	ahora := sesion.Inicio().Add(30 * time.Minute)
	if err := sesion.ReportarAusencia("escolar-1", ahora); err != nil {
		t.Fatalf("reportar ausencia falló: %v", err)
	}
	if len(sesion.Ausencias()) != 1 || sesion.Ausencias()[0].UsuarioID() != "escolar-1" {
		t.Fatalf("ausencia inesperada: %+v", sesion.Ausencias())
	}
	if err := sesion.ReportarAusencia("escolar-1", ahora); err == nil {
		t.Fatal("no se debe reportar dos veces la misma ausencia")
	}
	if err := sesion.ReportarAusencia("otro", ahora); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("se esperaba no autorizado, se obtuvo %v", err)
	}
	if err := sesion.ReportarAusencia("tutor-1", sesion.Fin().Add(PlazoDeclaracion+time.Hour)); !errors.Is(err, ErrFueraDePlazo) {
		t.Fatalf("se esperaba fuera de plazo, se obtuvo %v", err)
	}
}

func TestVentanasYResultado(t *testing.T) {
	sesion := sesionDePrueba(t)
	declaracion := sesion.VentanaDeclaracion()
	if !declaracion.Contiene(sesion.Inicio()) || declaracion.Contiene(sesion.Fin().Add(PlazoDeclaracion+time.Hour)) {
		t.Fatal("ventana de declaración incorrecta")
	}
	resultadoVentana := sesion.VentanaResultado()
	if !resultadoVentana.Contiene(sesion.Fin()) || resultadoVentana.Contiene(sesion.Inicio().Add(-time.Minute)) {
		t.Fatal("ventana de resultado incorrecta")
	}

	ahora := sesion.Fin().Add(time.Hour)
	if err := sesion.RegistrarResultado("escolar-1", "resumen", ahora); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("solo el tutor debe registrar el resultado: %v", err)
	}
	if err := sesion.RegistrarResultado("tutor-1", "", ahora); !errors.Is(err, ErrResultadoInvalido) {
		t.Fatalf("un resumen vacío debe rechazarse: %v", err)
	}
	if err := sesion.RegistrarResultado("tutor-1", "Se resolvieron los ejercicios.", ahora); err != nil {
		t.Fatalf("resultado válido falló: %v", err)
	}
	if sesion.Resultado() == nil || sesion.Resultado().UsuarioID() != "tutor-1" {
		t.Fatal("el resultado no quedó registrado")
	}
	if err := sesion.RegistrarResultado("tutor-1", "otro", ahora); !errors.Is(err, ErrResultadoDuplicado) {
		t.Fatalf("se esperaba resultado duplicado, se obtuvo %v", err)
	}
	copia := sesion.Clonar()
	if copia == nil || copia.Resultado() == nil {
		t.Fatal("el clon debe conservar el resultado")
	}
	copia.RegistrarResultado("tutor-1", "otro", ahora)
	if sesion.Resultado().Resumen() != "Se resolvieron los ejercicios." {
		t.Fatal("el clon no debe compartir el resultado")
	}
}

func TestRestaurarEstadoPersistido(t *testing.T) {
	sesion := sesionDePrueba(t)
	ahora := sesion.Fin().Add(time.Hour)
	if err := sesion.RestaurarAsistencia("escolar-1", ahora); err != nil {
		t.Fatalf("restaurar asistencia falló: %v", err)
	}
	if err := sesion.RestaurarAusencia("tutor-1", ahora); err != nil {
		t.Fatalf("restaurar ausencia falló: %v", err)
	}
	if err := sesion.RestaurarResultado("tutor-1", "Resultado persistido", ahora); err != nil {
		t.Fatalf("restaurar resultado falló: %v", err)
	}
	if len(sesion.Asistencias()) != 1 || len(sesion.Ausencias()) != 1 || sesion.Resultado() == nil {
		t.Fatal("el estado restaurado no quedó completo")
	}
	if err := sesion.RestaurarResultado("otro", "x", ahora); err == nil {
		t.Fatal("un resultado de otro usuario debe rechazarse")
	}
}
