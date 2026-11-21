package catalogo

import (
	"context"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/catalogo"
)

func TestSembrarAprobacionSinteticaSoloUnaVez(t *testing.T) {
	ctx := context.Background()
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria())
	servicio.ahora = func() time.Time { return time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC) }

	if _, err := servicio.Importar(ctx, DatosImportacion{
		Fuente: FuenteSinteticaInicial,
		Cursos: []CursoEntrada{
			{Nombre: "Matemática básica", Grado: "1.° de secundaria"},
			{Nombre: "Comunicación", Grado: "1.° de secundaria"},
		},
		Colegios: []ColegioEntrada{{Nombre: "Colegio Sintético", Distrito: "Distrito Sintético"}},
	}); err != nil {
		t.Fatal(err)
	}

	if err := SembrarAprobacionSintetica(ctx, servicio, "coordinador-sintetico"); err != nil {
		t.Fatal(err)
	}
	publica, err := servicio.CatalogoPublico(ctx)
	if err != nil || len(publica.Cursos) != 1 || publica.Cursos[0].Nombre != "Matemática básica" {
		t.Fatalf("la semilla debe dejar un curso público: %+v, error: %v", publica, err)
	}
	pendientes, err := servicio.CursosPendientes(ctx)
	if err != nil || len(pendientes) != 1 || pendientes[0].Nombre != "Comunicación" {
		t.Fatalf("la semilla debe conservar cursos pendientes para revisión: %+v, error: %v", pendientes, err)
	}

	if err := SembrarAprobacionSintetica(ctx, servicio, "coordinador-sintetico"); err != nil {
		t.Fatal(err)
	}
	publica, err = servicio.CatalogoPublico(ctx)
	if err != nil || len(publica.Cursos) != 1 {
		t.Fatalf("la semilla debe ser idempotente: %+v, error: %v", publica, err)
	}
}

func TestSembrarAprobacionSinteticaRespetaRevisionYFuente(t *testing.T) {
	ctx := context.Background()
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria())
	servicio.ahora = func() time.Time { return time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC) }

	importada, err := servicio.Importar(ctx, DatosImportacion{
		Fuente: FuenteSinteticaInicial,
		Cursos: []CursoEntrada{
			{Nombre: "Matemática básica", Grado: "1.° de secundaria"},
		},
		Colegios: []ColegioEntrada{{Nombre: "Colegio Sintético", Distrito: "Distrito Sintético"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.RevisarCurso(ctx, importada.Cursos[0].ID, "coordinador-real", EstadoRechazado, "No corresponde"); err != nil {
		t.Fatal(err)
	}
	if err := SembrarAprobacionSintetica(ctx, servicio, "coordinador-sintetico"); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.repositorio.Curso(ctx, importada.Cursos[0].ID); err != nil {
		t.Fatal(err)
	}
	publica, err := servicio.CatalogoPublico(ctx)
	if err != nil || len(publica.Cursos) != 0 {
		t.Fatalf("la semilla no debe revertir un rechazo: %+v, error: %v", publica, err)
	}

	otro := NuevoServicio(memoria.NuevoRepositorioMemoria())
	if _, err := otro.Importar(ctx, DatosImportacion{
		Fuente:   "Padrón institucional",
		Cursos:   []CursoEntrada{{Nombre: "Matemática básica", Grado: "1.° de secundaria"}},
		Colegios: []ColegioEntrada{{Nombre: "Colegio Real", Distrito: "Lima"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := SembrarAprobacionSintetica(ctx, otro, "coordinador-sintetico"); err != nil {
		t.Fatal(err)
	}
	publica, err = otro.CatalogoPublico(ctx)
	if err != nil || len(publica.Cursos) != 0 {
		t.Fatalf("la semilla no debe tocar importaciones reales: %+v, error: %v", publica, err)
	}
	curso, err := otro.repositorio.Curso(ctx, 1)
	if err != nil {
		t.Fatalf("el curso real debe seguir existiendo: %v", err)
	}
	if curso.Estado != EstadoPendiente {
		t.Fatalf("el curso real debe seguir pendiente: %+v", curso)
	}
}
