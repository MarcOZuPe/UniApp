package sesion

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRepositorioMemoriaCrearYConsultar(testingT *testing.T) {
	contexto := context.Background()
	repositorio := NuevoRepositorioMemoria()
	sesion := nuevaSesionDePrueba(testingT)

	if err := repositorio.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatalf("crear sesion: %v", err)
	}

	obtenida, err := repositorio.SesionPorID(contexto, sesion.ID())
	if err != nil {
		testingT.Fatalf("consultar sesion: %v", err)
	}

	if obtenida.ID() != sesion.ID() ||
		obtenida.ReservaID() != sesion.ReservaID() {
		testingT.Fatal("la sesion recuperada no coincide")
	}
}

func TestRepositorioMemoriaSesionNoEncontrada(testingT *testing.T) {
	repositorio := NuevoRepositorioMemoria()

	_, err := repositorio.SesionPorID(context.Background(), "inexistente")
	if !errors.Is(err, ErrSesionNoEncontrada) {
		testingT.Fatalf("se esperaba ErrSesionNoEncontrada, se obtuvo %v", err)
	}
}

func TestRepositorioMemoriaRechazaDuplicados(testingT *testing.T) {
	contexto := context.Background()
	repositorio := NuevoRepositorioMemoria()
	sesion := nuevaSesionDePrueba(testingT)

	if err := repositorio.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatal(err)
	}

	if err := repositorio.CrearSesion(contexto, sesion); !errors.Is(err, ErrSesionDuplicada) {
		testingT.Fatalf("debe rechazar ID duplicado: %v", err)
	}

	otra, err := NuevaSesion(
		"sesion-002",
		sesion.ReservaID(),
		"tutor-001",
		"escolar-001",
		sesion.Inicio(),
		sesion.Fin(),
	)
	if err != nil {
		testingT.Fatal(err)
	}

	if err := repositorio.CrearSesion(contexto, otra); !errors.Is(err, ErrSesionDuplicada) {
		testingT.Fatalf("debe rechazar reserva duplicada: %v", err)
	}
}

func TestRepositorioMemoriaDeclararAsistencia(testingT *testing.T) {
	contexto := context.Background()
	repositorio := NuevoRepositorioMemoria()
	sesion := nuevaSesionDePrueba(testingT)

	if err := repositorio.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatal(err)
	}

	momento := sesion.Inicio()
	asistencia, err := repositorio.DeclararAsistencia(contexto, sesion.ID(), sesion.TutorID(), momento)
	if err != nil {
		testingT.Fatalf("declarar asistencia: %v", err)
	}

	if asistencia.Rol() != RolTutor ||
		asistencia.UsuarioID() != sesion.TutorID() {
		testingT.Fatal("datos de asistencia incorrectos")
	}

	guardada, err := repositorio.SesionPorID(contexto, sesion.ID())
	if err != nil {
		testingT.Fatal(err)
	}

	if len(guardada.Asistencias()) != 1 {
		testingT.Fatal("la asistencia no fue almacenada")
	}
}

func TestRepositorioMemoriaRechazaParticipanteAjeno(testingT *testing.T) {
	contexto := context.Background()
	repositorio := NuevoRepositorioMemoria()
	sesion := nuevaSesionDePrueba(testingT)

	if err := repositorio.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatal(err)
	}

	_, err := repositorio.DeclararAsistencia(
		contexto, sesion.ID(), "usuario-ajeno", sesion.Inicio(),
	)

	if !errors.Is(err, ErrNoAutorizado) {
		testingT.Fatalf("se esperaba ErrNoAutorizado, se obtuvo %v", err)
	}
}

func TestRepositorioMemoriaRechazaAsistenciaDuplicada(testingT *testing.T) {
	contexto := context.Background()
	repositorio := NuevoRepositorioMemoria()
	sesion := nuevaSesionDePrueba(testingT)

	if err := repositorio.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatal(err)
	}

	_, err := repositorio.DeclararAsistencia(
		contexto, sesion.ID(), sesion.TutorID(), sesion.Inicio(),
	)
	if err != nil {
		testingT.Fatal(err)
	}

	_, err = repositorio.DeclararAsistencia(
		contexto, sesion.ID(), sesion.TutorID(), sesion.Inicio(),
	)

	if !errors.Is(err, ErrAsistenciaDuplicada) {
		testingT.Fatalf("se esperaba ErrAsistenciaDuplicada, se obtuvo %v", err)
	}
}

func TestRepositorioMemoriaConcurrencia(testingT *testing.T) {
	contexto := context.Background()
	repositorio := NuevoRepositorioMemoria()
	sesion := nuevaSesionDePrueba(testingT)

	if err := repositorio.CrearSesion(contexto, sesion); err != nil {
		testingT.Fatal(err)
	}

	const solicitudes = 10

	resultados := make(chan error, solicitudes)
	var grupo sync.WaitGroup

	for indiceSolicitud := 0; indiceSolicitud < solicitudes; indiceSolicitud++ {
		grupo.Add(1)

		go func() {
			defer grupo.Done()
			_, err := repositorio.DeclararAsistencia(
				contexto, sesion.ID(), sesion.TutorID(), sesion.Inicio(),
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
			testingT.Fatalf("error inesperado: %v", err)
		}
	}

	if exitos != 1 || duplicados != solicitudes-1 {
		testingT.Fatalf(
			"se esperaba 1 exito y %d duplicados; hubo %d exitos y %d duplicados",
			solicitudes-1, exitos, duplicados,
		)
	}
}

func TestI11MemoriaAsistenciaYAusenciaConcurrentes(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)
	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReportarAusencia(ctx, "inexistente", s.TutorID(), s.Fin()); !errors.Is(err, ErrSesionNoEncontrada) {
		t.Fatalf("sesion inexistente: %v", err)
	}
	if _, err := repo.ReportarAusencia(ctx, s.ID(), "ajeno", s.Fin()); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("autorizacion: %v", err)
	}
	resultados := make(chan error, 2)
	inicio := make(chan struct{})
	go func() {
		<-inicio
		_, err := repo.ReportarAusencia(ctx, s.ID(), s.TutorID(), s.Fin())
		resultados <- err
	}()
	go func() {
		<-inicio
		_, err := repo.DeclararAsistencia(ctx, s.ID(), s.TutorID(), s.Fin())
		resultados <- err
	}()
	close(inicio)
	exitos, conflictos := 0, 0
	for i := 0; i < 2; i++ {
		err := <-resultados
		if err == nil {
			exitos++
		} else if errors.Is(err, ErrDeclaracionContradictoria) {
			conflictos++
		} else {
			t.Fatal(err)
		}
	}
	if exitos != 1 || conflictos != 1 {
		t.Fatalf("exitos=%d conflictos=%d", exitos, conflictos)
	}
	if _, err := repo.ReportarAusencia(ctx, s.ID(), s.EscolarID(), s.Fin()); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ReportarAusencia(ctx, s.ID(), s.EscolarID(), s.Fin()); !errors.Is(err, ErrAusenciaDuplicada) {
		t.Fatalf("duplicado: %v", err)
	}
	guardada, err := repo.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatal(err)
	}
	ausencias := guardada.Ausencias()
	ultima := ausencias[len(ausencias)-1]
	if ultima.UsuarioID() != s.EscolarID() || !ultima.ReportadaEn().Equal(s.Fin()) {
		t.Fatal("ausencia no guardada")
	}
	if len(ausencias)+len(guardada.Asistencias()) != 2 {
		t.Fatal("declaraciones contradictorias persistidas")
	}
}

func TestI11MemoriaResultadoConcurrente(t *testing.T) {
	ctx := context.Background()
	repo := NuevoRepositorioMemoria()
	s := nuevaSesionDePrueba(t)
	if err := repo.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.RegistrarResultado(ctx, "inexistente", s.TutorID(), "Resumen", s.Fin()); !errors.Is(err, ErrSesionNoEncontrada) {
		t.Fatalf("inexistente: %v", err)
	}
	if _, err := repo.RegistrarResultado(ctx, s.ID(), s.EscolarID(), "Resumen", s.Fin()); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("autorizacion: %v", err)
	}
	resultados := make(chan error, 2)
	inicio := make(chan struct{})
	for i := 0; i < 2; i++ {
		go func() {
			<-inicio
			_, err := repo.RegistrarResultado(ctx, s.ID(), s.TutorID(), "Resumen", s.Fin())
			resultados <- err
		}()
	}
	close(inicio)
	exitos, duplicados := 0, 0
	for i := 0; i < 2; i++ {
		err := <-resultados
		if err == nil {
			exitos++
		} else if errors.Is(err, ErrResultadoDuplicado) {
			duplicados++
		} else {
			t.Fatal(err)
		}
	}
	if exitos != 1 || duplicados != 1 {
		t.Fatalf("exitos=%d duplicados=%d", exitos, duplicados)
	}
	guardada, err := repo.SesionPorID(ctx, s.ID())
	if err != nil {
		t.Fatal(err)
	}
	if guardada.Resultado() == nil || guardada.Resultado().Resumen() != "Resumen" || guardada.Resultado().UsuarioID() != s.TutorID() {
		t.Fatal("resultado no guardado")
	}
}

func nuevaSesionDePrueba(testingT *testing.T) *Sesion {
	testingT.Helper()

	inicio := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)

	sesion, err := NuevaSesion(
		"sesion-001",
		"reserva-001",
		"tutor-001",
		"escolar-001",
		inicio,
		inicio.Add(DuracionSesion),
	)
	if err != nil {
		testingT.Fatalf("no se pudo crear la sesion: %v", err)
	}

	return sesion
}
