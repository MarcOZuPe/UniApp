package incidente

import (
	"context"
	"testing"
	"time"
)

func TestRepositorioMemoriaIncidente(t *testing.T) {
	repo := NuevoRepositorioMemoria()
	ctx := context.Background()
	base := time.Now().UTC().Add(-2 * time.Hour)
	repo.SembrarTutoria(Tutoria{
		ReservaID: "r-1", EscolarID: "e-1", EscolarNombre: "Escolar",
		TutorID: "t-1", TutorNombre: "Tutor", TutorFacultadID: "fac-1",
		Curso: "Álgebra", Inicio: base, Fin: base.Add(time.Hour),
	})
	repo.SembrarVinculacion("a-1", EscolarVinculado{ID: "e-1", Nombre: "Escolar"})

	tutorias, err := repo.TutoriasDeEscolar(ctx, "e-1", time.Now())
	if err != nil || len(tutorias) != 1 {
		t.Fatalf("tutorías inesperadas: %+v err=%v", tutorias, err)
	}
	if _, err := repo.TutoriaDeEscolar(ctx, "e-1", "r-1", time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.TutoriaDeEscolar(ctx, "e-1", "otra", time.Now()); err != ErrTutoriaNoEncontrada {
		t.Fatalf("se esperaba ErrTutoriaNoEncontrada: %v", err)
	}
	escolares, err := repo.EscolaresDeApoderado(ctx, "a-1")
	if err != nil || len(escolares) != 1 {
		t.Fatalf("escolares inesperados: %+v err=%v", escolares, err)
	}
	if vinculado, err := repo.EsApoderadoDe(ctx, "a-1", "e-1"); err != nil || !vinculado {
		t.Fatalf("el apoderado debería estar vinculado: %v %v", vinculado, err)
	}
	if vinculado, _ := repo.EsApoderadoDe(ctx, "a-1", "otro"); vinculado {
		t.Fatal("un escolar ajeno no debe estar vinculado")
	}

	creado, err := repo.Crear(ctx, &Incidente{
		EscolarID: "e-1", ReportadoPor: "e-1", TutorID: "t-1", FacultadID: "fac-1",
		ReservaID: "r-1", Tipo: TipoTratoInadecuado, Estado: EstadoPendiente,
		Evidencias: []Evidencia{{Nombre: "a.pdf", Tipo: "application/pdf"}},
	})
	if err != nil || creado.ID == "" || len(creado.Evidencias) != 1 || creado.Evidencias[0].ID == "" {
		t.Fatalf("incidente creado inesperado: %+v err=%v", creado, err)
	}
	if _, err := repo.PorID(ctx, "inexistente"); err != ErrIncidenteNoEncontrado {
		t.Fatalf("se esperaba ErrIncidenteNoEncontrado: %v", err)
	}
	porFacultad, err := repo.Listar(ctx, Filtro{FacultadID: "fac-1"})
	if err != nil || len(porFacultad) != 1 {
		t.Fatalf("listado por facultad inesperado: %+v err=%v", porFacultad, err)
	}
	porEscolar, err := repo.Listar(ctx, Filtro{EscolarID: "e-1"})
	if err != nil || len(porEscolar) != 1 {
		t.Fatalf("listado por escolar inesperado: %+v err=%v", porEscolar, err)
	}
	porApoderado, err := repo.Listar(ctx, Filtro{EscolaresID: []string{"e-1"}})
	if err != nil || len(porApoderado) != 1 {
		t.Fatalf("listado por apoderado inesperado: %+v err=%v", porApoderado, err)
	}
	if _, err := repo.Revisar(ctx, "inexistente", "coord-1", ResultadoProcede, "comentario", time.Now()); err != ErrIncidenteNoEncontrado {
		t.Fatalf("se esperaba ErrIncidenteNoEncontrado: %v", err)
	}
	revisado, err := repo.Revisar(ctx, creado.ID, "coord-1", ResultadoProcede, "comentario", time.Now())
	if err != nil || revisado.Estado != EstadoRevisado || revisado.RevisadoPor != "coord-1" {
		t.Fatalf("revisión inesperada: %+v err=%v", revisado, err)
	}
	if _, err := repo.Revisar(ctx, creado.ID, "coord-1", ResultadoProcede, "comentario", time.Now()); err != ErrYaRevisado {
		t.Fatalf("se esperaba ErrYaRevisado: %v", err)
	}
	evidencia, err := repo.EvidenciaDe(ctx, creado.ID, creado.Evidencias[0].ID)
	if err != nil || evidencia.Nombre != "a.pdf" {
		t.Fatalf("evidencia inesperada: %+v err=%v", evidencia, err)
	}
	if _, err := repo.EvidenciaDe(ctx, creado.ID, "inexistente"); err != ErrEvidenciaNoEncontrada {
		t.Fatalf("se esperaba ErrEvidenciaNoEncontrada: %v", err)
	}
}
