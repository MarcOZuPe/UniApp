package catalogo

import (
	"context"
	"testing"
)

func TestRepositorioMemoriaCatalogo(t *testing.T) {
	repo := NuevoRepositorioMemoria()
	ctx := context.Background()
	if _, err := repo.ImportacionActual(ctx); err == nil {
		t.Fatal("un repositorio vacío no debe tener importación")
	}
	importacion := &Importacion{
		Fuente:   "origen",
		Cursos:   []Curso{{Nombre: "Álgebra", Grado: "5to", Estado: EstadoPendiente}},
		Colegios: []Colegio{{Nombre: "Colegio", Distrito: "Lima"}},
	}
	if err := repo.Importar(ctx, importacion); err != nil {
		t.Fatal(err)
	}
	if importacion.ID == 0 || importacion.Cursos[0].ID == 0 {
		t.Fatal("la importación y el curso deben tener ID")
	}
	actual, err := repo.ImportacionActual(ctx)
	if err != nil || len(actual.Cursos) != 1 || len(actual.Colegios) != 1 {
		t.Fatalf("importación actual inesperada: %+v err=%v", actual, err)
	}
	curso, err := repo.Curso(ctx, importacion.Cursos[0].ID)
	if err != nil || curso.Nombre != "Álgebra" {
		t.Fatalf("curso inesperado: %+v err=%v", curso, err)
	}
	if _, err := repo.Curso(ctx, 9999); err == nil {
		t.Fatal("un curso inexistente debe fallar")
	}
	curso.Estado = EstadoAprobado
	if err := repo.RevisarCurso(ctx, curso, "coord-1"); err != nil {
		t.Fatalf("revisar curso falló: %v", err)
	}
	if err := repo.RevisarCurso(ctx, curso, "coord-1"); err == nil {
		t.Fatal("un curso ya revisado debe rechazarse")
	}
	solicitud := SolicitudColegio{Nombre: "Otro", Distrito: "Surco"}
	if err := repo.SolicitarColegio(ctx, solicitud, "u-1"); err != nil {
		t.Fatalf("solicitar colegio falló: %v", err)
	}
	if err := repo.SolicitarColegio(ctx, solicitud, "u-1"); err == nil {
		t.Fatal("una solicitud pendiente duplicada debe rechazarse")
	}
	pendientes, err := repo.SolicitudesPendientes(ctx)
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("solicitudes pendientes inesperadas: %+v err=%v", pendientes, err)
	}
	if err := repo.Importar(ctx, &Importacion{
		Fuente:   "segundo",
		Cursos:   []Curso{{Nombre: "Física", Grado: "5to", Estado: EstadoPendiente}},
		Colegios: []Colegio{{Nombre: "Otro", Distrito: "Surco"}},
	}); err != nil {
		t.Fatal(err)
	}
	pendientes, _ = repo.SolicitudesPendientes(ctx)
	if len(pendientes) != 0 {
		t.Fatal("una solicitud incorporada ya no debe estar pendiente")
	}
}
