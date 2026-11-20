package incidente

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPrepararEvidenciaSoloAceptaContenidoYExtensionPermitidos(t *testing.T) {
	ahora := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	casos := []struct {
		nombre    string
		contenido []byte
		ok        bool
	}{
		{"captura.pdf", []byte("%PDF-1.4\n% prueba\n"), true},
		{"foto.png", append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, []byte("datos")...), true},
		{"foto.jpg", append([]byte{0xff, 0xd8, 0xff, 0xe0, 0, 0x10, 'J', 'F', 'I', 'F', 0, 1}, []byte("datos")...), true},
		{"foto.jpeg", append([]byte{0xff, 0xd8, 0xff, 0xe0, 0, 0x10, 'J', 'F', 'I', 'F', 0, 1}, []byte("datos")...), true},
		{"notas.txt", []byte("texto plano"), false},
		{"captura.pdf.txt", []byte("%PDF-1.4\n% prueba\n"), false},
		{"captura.pdf", []byte("no es un pdf"), false},
		{"../captura.pdf", []byte("%PDF-1.4\n% prueba\n"), true},
		{"captura\n.pdf", []byte("%PDF-1.4\n% prueba\n"), false},
	}
	for _, caso := range casos {
		evidencia, err := PrepararEvidencia(Archivo{Nombre: caso.nombre, Contenido: caso.contenido}, ahora)
		if caso.ok && (err != nil || evidencia.Nombre == "" || evidencia.Tipo == "") {
			t.Fatalf("%s: se esperaba evidencia válida, error: %v", caso.nombre, err)
		}
		if !caso.ok && !errors.Is(err, ErrEvidenciaInvalida) {
			t.Fatalf("%s: se esperaba ErrEvidenciaInvalida, se obtuvo %v", caso.nombre, err)
		}
	}
}

func TestNuevoIncidenteExigeEvidenciaYTutoriaPrevia(t *testing.T) {
	ahora := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	tutoria := Tutoria{
		ReservaID: "reserva-1", EscolarID: "escolar-1", EscolarNombre: "Escolar",
		TutorID: "tutor-1", TutorNombre: "Tutor", Curso: "Álgebra",
		Inicio: ahora.Add(-2 * time.Hour), Fin: ahora.Add(-time.Hour),
	}
	datos := DatosIncidente{
		EscolarID: "escolar-1", ReportadoPor: "escolar-1", ReservaID: "reserva-1",
		Tipo: TipoTratoInadecuado, Descripcion: strings.Repeat("a", DescripcionMinima),
	}
	if _, err := NuevoIncidente(datos, &tutoria, ahora); err == nil {
		t.Fatal("no debe aceptar un incidente sin evidencia")
	}
	datos.Evidencias = []Archivo{{Nombre: "captura.pdf", Contenido: []byte("%PDF-1.4\n% prueba\n")}}
	incidente, err := NuevoIncidente(datos, &tutoria, ahora)
	if err != nil || incidente.Estado != EstadoPendiente || len(incidente.Evidencias) != 1 {
		t.Fatalf("incidente válido: %+v, error: %v", incidente, err)
	}
	tutoria.Fin = ahora.Add(time.Hour)
	if _, err := NuevoIncidente(datos, &tutoria, ahora); err == nil {
		t.Fatal("no debe aceptar una tutoría futura")
	}
}

func TestRevisarIncidenteExigeDecisionYComentario(t *testing.T) {
	incidente := &Incidente{Estado: EstadoPendiente}
	ahora := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: "otro", Comentario: "ok"}, ahora); err == nil {
		t.Fatal("resultado inválido")
	}
	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: ResultadoProcede, Comentario: "x"}, ahora); err == nil {
		t.Fatal("comentario demasiado corto")
	}
	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: ResultadoProcede, Comentario: "Se revisó la evidencia."}, ahora); err != nil {
		t.Fatal(err)
	}
	if incidente.Estado != EstadoRevisado || incidente.Resultado != ResultadoProcede || incidente.RevisadoEn == nil {
		t.Fatalf("revisión incompleta: %+v", incidente)
	}
	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: ResultadoNoProcede, Comentario: "Otra revisión"}, ahora); !errors.Is(err, ErrYaRevisado) {
		t.Fatalf("segunda revisión: %v", err)
	}
}
