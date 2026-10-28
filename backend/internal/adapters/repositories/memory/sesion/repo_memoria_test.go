package sesion

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func nuevaSesionDePrueba(t *testing.T) *Sesion {
	t.Helper()

	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)

	s, err := NuevaSesion(
		"sesion-001",
		"reserva-001",
		"tutor-001",
		"escolar-001",
		inicio,
		inicio.Add(DuracionSesion),
	)
	if err != nil {
		t.Fatalf("no se pudo crear la sesion: %v", err)
	}

	return s
}

func TestRepositorioMemoriaCrearYConsultar(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)

	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatalf("crear sesion: %v", err)
	}

	obtenida, err := repo.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatalf("consultar sesion: %v", err)
	}

	if obtenida.ID() != s.ID() ||
		obtenida.ReservaID() != s.ReservaID() {
		t.Fatal("la sesion recuperada no coincide")
	}
}

func TestRepositorioMemoriaSesionNoEncontrada(t *testing.T) {
	repo := NuevoRepositorioMemoria()

	_, err := repo.SesionPorID(context.Background(), "inexistente")
	if !errors.Is(err, ErrSesionNoEncontrada) {
		t.Fatalf("se esperaba ErrSesionNoEncontrada, se obtuvo %v", err)
	}
}

func TestRepositorioMemoriaRechazaDuplicados(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)

	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}

	if err := repo.CrearSesion(ctx, s); !errors.Is(err, ErrSesionDuplicada) {
		t.Fatalf("debe rechazar ID duplicado: %v", err)
	}

	otra, err := NuevaSesion(
		"sesion-002",
		s.ReservaID(),
		"tutor-001",
		"escolar-001",
		s.Inicio(),
		s.Fin(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.CrearSesion(ctx, otra); !errors.Is(err, ErrSesionDuplicada) {
		t.Fatalf("debe rechazar reserva duplicada: %v", err)
	}
}

func TestRepositorioMemoriaDeclararAsistencia(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)

	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}

	ahora := s.Inicio()
	asistencia, err := repo.DeclararAsistencia(ctx, s.ID(), s.TutorID(), ahora)
	if err != nil {
		t.Fatalf("declarar asistencia: %v", err)
	}

	if asistencia.Rol() != RolTutor ||
		asistencia.UsuarioID() != s.TutorID() {
		t.Fatal("datos de asistencia incorrectos")
	}

	guardada, err := repo.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatal(err)
	}

	if len(guardada.Asistencias()) != 1 {
		t.Fatal("la asistencia no fue almacenada")
	}
}

func TestRepositorioMemoriaRechazaParticipanteAjeno(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)

	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}

	_, err := repo.DeclararAsistencia(
		ctx, s.ID(), "usuario-ajeno", s.Inicio(),
	)

	if !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
	}
}

func TestRepositorioMemoriaRechazaAsistenciaDuplicada(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)

	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}

	_, err := repo.DeclararAsistencia(
		ctx, s.ID(), s.TutorID(), s.Inicio(),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.DeclararAsistencia(
		ctx, s.ID(), s.TutorID(), s.Inicio(),
	)

	if !errors.Is(err, ErrAsistenciaDuplicada) {
		t.Fatalf("se esperaba ErrAsistenciaDuplicada, se obtuvo %v", err)
	}
}

func TestRepositorioMemoriaConcurrencia(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)

	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}

	const solicitudes = 10

	resultados := make(chan error, solicitudes)
	var grupo sync.WaitGroup

	for i := 0; i < solicitudes; i++ {
		grupo.Add(1)

		go func() {
			defer grupo.Done()
			_, err := repo.DeclararAsistencia(
				ctx, s.ID(), s.TutorID(), s.Inicio(),
			)
			resultados <- err
		}()
	}

	grupo.Wait()
	close(resultados)

	exitos := 0
	duplicados := 0

	for err := range resultados {
		switch {
		case err == nil:
			exitos++
		case errors.Is(err, ErrAsistenciaDuplicada):
			duplicados++
		default:
			t.Fatalf("error inesperado: %v", err)
		}
	}

	if exitos != 1 || duplicados != solicitudes-1 {
		t.Fatalf(
			"se esperaba 1 exito y %d duplicados; hubo %d exitos y %d duplicados",
			solicitudes-1, exitos, duplicados,
		)
	}
}
