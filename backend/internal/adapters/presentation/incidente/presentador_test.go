package incidente

import (
	"testing"
	"time"

	modelo "uniapp/internal/application/incidente"
)

func TestPresentarIncidenteYTutorias(t *testing.T) {
	inicio := time.Date(2026, 11, 26, 18, 0, 0, 0, time.UTC)
	incidente := &modelo.Incidente{
		ID: "i-1", EscolarID: "e-1", EscolarNombre: "Escolar", ReportadoPor: "e-1",
		TutorID: "t-1", TutorNombre: "Tutor", ReservaID: "r-1", Curso: "Álgebra",
		Inicio: inicio, Tipo: modelo.TipoTratoInadecuado, Descripcion: "descripción",
		Estado: modelo.EstadoRevisado, Resultado: modelo.ResultadoProcede,
		ComentarioRevision: "comentario", RevisadoPor: "c-1", RevisadoEn: &inicio,
		CreadoEn:   inicio,
		Evidencias: []modelo.Evidencia{{ID: "ev-1", Nombre: "a.pdf", Tipo: "application/pdf", TamanoBytes: 10, CreadoEn: inicio}},
	}
	presentado := PresentarIncidente(incidente)
	if presentado.ID != "i-1" || presentado.TipoEtiqueta != "Trato inadecuado" || presentado.ReportadoPorRol != "escolar" {
		t.Fatalf("incidente presentado inesperado: %+v", presentado)
	}
	if presentado.InicioLima != "2026-11-26 13:00" {
		t.Fatalf("hora Lima inesperada: %q", presentado.InicioLima)
	}
	if len(presentado.Evidencias) != 1 || presentado.Evidencias[0].DescargarEn == "" {
		t.Fatalf("evidencias inesperadas: %+v", presentado.Evidencias)
	}
	if PresentarIncidente(nil).Evidencias == nil {
		t.Fatal("el incidente nulo debe devolver evidencias vacías")
	}
	if len(PresentarIncidentes([]modelo.Incidente{*incidente})) != 1 {
		t.Fatal("la lista debería presentar un incidente")
	}

	tutorias := PresentarTutorias([]modelo.Tutoria{{
		ReservaID: "r-1", EscolarID: "e-1", EscolarNombre: "Escolar",
		TutorID: "t-1", TutorNombre: "Tutor", Curso: "Álgebra", Inicio: inicio, Fin: inicio.Add(time.Hour),
	}})
	if len(tutorias) != 1 || tutorias[0].InicioLima != "2026-11-26 13:00" || tutorias[0].FinLima != "2026-11-26 14:00" {
		t.Fatalf("tutorías presentadas inesperadas: %+v", tutorias)
	}
}
