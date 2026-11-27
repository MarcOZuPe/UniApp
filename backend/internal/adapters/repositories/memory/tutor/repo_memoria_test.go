package tutor

import (
	"context"
	"testing"
	"time"
)

func expedientePrueba(tutorID, facultad string, estado Estado) *Expediente {
	return &Expediente{
		TutorID: tutorID, FacultadID: facultad, Presentacion: "Presentación",
		Cursos: "Álgebra", Experiencia: "Experiencia", Estado: estado,
		Documentos: []Documento{{ID: "d-" + tutorID, Nombre: "a.pdf", Tamano: 10, Tipo: "application/pdf"}},
	}
}

func TestRepositorioMemoriaTutor(t *testing.T) {
	repo := NuevoRepositorioMemoria()
	ctx := context.Background()
	if _, err := repo.PorTutor(ctx, "t-1"); err == nil {
		t.Fatal("un tutor inexistente debe fallar")
	}
	expediente := expedientePrueba("t-1", "fac-1", EstadoPendiente)
	if err := repo.Guardar(ctx, expediente); err != nil {
		t.Fatal(err)
	}
	guardado, err := repo.PorTutor(ctx, "t-1")
	if err != nil || guardado.TutorID != "t-1" {
		t.Fatalf("expediente guardado inesperado: %+v err=%v", guardado, err)
	}
	pendientes, err := repo.Pendientes(ctx, "fac-1")
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("pendientes inesperados: %+v err=%v", pendientes, err)
	}
	pendientes, err = repo.Pendientes(ctx, "fac-2")
	if err != nil || len(pendientes) != 0 {
		t.Fatalf("no debería haber pendientes de otra facultad: %+v err=%v", pendientes, err)
	}
	if err := repo.Revisar(ctx, &Expediente{TutorID: "t-inexistente", Estado: EstadoAprobado}); err == nil {
		t.Fatal("revisar un inexistente debe fallar")
	}
	aprobado := expedientePrueba("t-1", "fac-1", EstadoAprobado)
	vencimiento := time.Now().Add(24 * time.Hour)
	aprobado.VenceEn = &vencimiento
	if err := repo.Revisar(ctx, aprobado); err != nil {
		t.Fatalf("revisar falló: %v", err)
	}
	aprobados, err := repo.Aprobados(ctx, "fac-1")
	if err != nil || len(aprobados) != 1 {
		t.Fatalf("aprobados inesperados: %+v err=%v", aprobados, err)
	}
	documento, propietario, err := repo.Documento(ctx, "d-t-1")
	if err != nil || propietario != "t-1" || documento.Nombre != "a.pdf" {
		t.Fatalf("documento inesperado: %+v %q err=%v", documento, propietario, err)
	}
	if _, _, err := repo.Documento(ctx, "inexistente"); err == nil {
		t.Fatal("un documento inexistente debe fallar")
	}
	perfiles, err := repo.PerfilesPublicos(ctx, time.Now())
	if err != nil || len(perfiles) != 1 || perfiles[0].TutorID != "t-1" {
		t.Fatalf("perfiles inesperados: %+v err=%v", perfiles, err)
	}
	if _, err := repo.PerfilesPublicos(ctx, vencimiento.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
}
