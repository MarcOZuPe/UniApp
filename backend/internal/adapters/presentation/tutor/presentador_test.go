package tutor

import (
	"testing"
	"time"

	modelo "uniapp/internal/application/tutor"
)

func TestPresentarExpedienteYPerfiles(t *testing.T) {
	ahora := time.Now()
	expediente := &modelo.Expediente{
		TutorID: "t-1", Presentacion: "Presentación", Cursos: "Álgebra",
		Experiencia: "Experiencia", Estado: modelo.EstadoAprobado, PresentadoEn: ahora,
		VenceEn: &ahora, Motivo: "",
		Documentos: []modelo.Documento{{ID: "d-1", Nombre: "a.pdf", Tamano: 10, Tipo: "application/pdf", CreadoEn: ahora}},
	}
	presentado := PresentarExpediente(expediente)
	if presentado.TutorID != "t-1" || presentado.Estado != "aprobado" || len(presentado.Documentos) != 1 {
		t.Fatalf("expediente presentado inesperado: %+v", presentado)
	}
	perfil := PresentarPerfilPublico(expediente)
	if perfil.TutorID != "t-1" || perfil.Presentacion != "Presentación" {
		t.Fatalf("perfil presentado inesperado: %+v", perfil)
	}
	perfiles := PresentarPerfilesPublicos([]modelo.PerfilListado{{TutorID: "t-2", Presentacion: "Otra", Cursos: "Física", Experiencia: "2 años"}})
	if len(perfiles) != 1 || perfiles[0].TutorID != "t-2" {
		t.Fatalf("perfiles presentados inesperados: %+v", perfiles)
	}
}
