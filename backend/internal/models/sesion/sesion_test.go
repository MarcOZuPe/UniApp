package sesion

import (
	"errors"
	"testing"
	"time"
)

func nuevaSesionPrueba(t *testing.T) *Sesion {
	t.Helper()

	inicio := time.Date(
		2026, 10, 20, 15, 0, 0, 0, time.UTC,
	)

	s, err := NuevaSesion(
		"sesion-001",
		"reserva-sintetica-001",
		"tutor-001",
		"escolar-001",
		inicio,
		inicio.Add(time.Hour),
	)

	if err != nil {
		t.Fatalf("No se pudo crear la sesion: %v", err)
	}

	return s
}

func TestNuevaSesion(t *testing.T) {
	sesion := nuevaSesionPrueba(t)

	if sesion.ID() != "sesion-001" {
		t.Fatal("Identificador incorrecto")
	}

	if sesion.ReservaID() != "reserva-sintetica-001" {
		t.Fatal("Reserva incorrecta")
	}
}

func TestDeclararAsistencia(t *testing.T) {
	sesion := nuevaSesionPrueba(t)

	err := sesion.DeclararAsistencia(
		"escolar-001",
		sesion.Inicio(),
	)

	if err != nil {
		t.Fatalf("No deberia fallar: %v", err)
	}

	asistencias := sesion.Asistencias()

	if len(asistencias) != 1 {
		t.Fatal("Se esperaba una declaracion")
	}

	if asistencias[0].Rol() != RolEscolar {
		t.Fatal("Rol incorrecto")
	}
}

func TestRechazarParticipanteAjeno(t *testing.T) {
	sesion := nuevaSesionPrueba(t)

	err := sesion.DeclararAsistencia(
		"usuario-ajeno",
		sesion.Inicio(),
	)

	if !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("Se esperaba acceso denegado: %v", err)
	}

	if len(sesion.Asistencias()) != 0 {
		t.Fatal("No debe registrarse el usuario ajeno")
	}
}

func TestRechazarDuplicado(t *testing.T) {
	sesion := nuevaSesionPrueba(t)

	if err := sesion.DeclararAsistencia(
		"tutor-001",
		sesion.Inicio(),
	); err != nil {
		t.Fatal(err)
	}

	err := sesion.DeclararAsistencia(
		"tutor-001",
		sesion.Inicio().Add(time.Minute),
	)

	if !errors.Is(err, ErrAsistenciaDuplicada) {
		t.Fatalf("Se esperaba duplicado: %v", err)
	}

	if len(sesion.Asistencias()) != 1 {
		t.Fatal("No debe duplicar declaraciones")
	}
}

func TestAmbosParticipantes(t *testing.T) {
	sesion := nuevaSesionPrueba(t)

	if err := sesion.DeclararAsistencia(
		"tutor-001", sesion.Inicio(),
	); err != nil {
		t.Fatal(err)
	}

	if err := sesion.DeclararAsistencia(
		"escolar-001", sesion.Inicio(),
	); err != nil {
		t.Fatal(err)
	}

	if len(sesion.Asistencias()) != 2 {
		t.Fatal("Se esperaban dos declaraciones independientes")
	}
}

func TestAsistenciasDevuelveCopia(t *testing.T) {
	sesion := nuevaSesionPrueba(t)

	if err := sesion.DeclararAsistencia(
		"tutor-001", sesion.Inicio(),
	); err != nil {
		t.Fatal(err)
	}

	copia := sesion.Asistencias()
	copia[0] = Asistencia{}

	if sesion.Asistencias()[0].UsuarioID() != "tutor-001" {
		t.Fatal("La coleccion interna fue modificada")
	}
}

func TestRechazarSesionInvalida(t *testing.T) {
	inicio := time.Date(
		2026, 10, 20, 15, 0, 0, 0, time.UTC,
	)

	_, err := NuevaSesion(
		"sesion-001",
		"reserva-001",
		"tutor-001",
		"escolar-001",
		inicio,
		inicio,
	)

	if !errors.Is(err, ErrSesionInvalida) {
		t.Fatalf("Se esperaba sesion invalida: %v", err)
	}
}
func TestRechazarSesionDe30Minutos(t *testing.T) {
	inicio := time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	fin := inicio.Add(30 * time.Minute)

	_, err := NuevaSesion(
		"sesion-30",
		"reserva-30",
		"tutor-1",
		"escolar-1",
		inicio,
		fin,
	)

	if !errors.Is(err, ErrSesionInvalida) {
		t.Fatalf("se esperaba ErrSesionInvalida, se obtuvo %v", err)
	}
}

func TestRechazarSesionDe90Minutos(t *testing.T) {
	inicio := time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	fin := inicio.Add(90 * time.Minute)

	_, err := NuevaSesion(
		"sesion-90",
		"reserva-90",
		"tutor-1",
		"escolar-1",
		inicio,
		fin,
	)

	if !errors.Is(err, ErrSesionInvalida) {
		t.Fatalf("se esperaba ErrSesionInvalida, se obtuvo %v", err)
	}
}
