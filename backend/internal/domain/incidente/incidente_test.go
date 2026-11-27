package incidente

import (
	"strings"
	"testing"
	"time"
)

var pdfValido = []byte("%PDF-1.4\n%evidencia de prueba\n")

func tutoriaFinalizada(base time.Time) *Tutoria {
	return &Tutoria{
		ReservaID: "reserva-1", EscolarID: "escolar-1", EscolarNombre: "Escolar Uno",
		TutorID: "tutor-1", TutorNombre: "Tutor Uno", TutorFacultadID: "fac-1",
		Curso: "Álgebra", Inicio: base, Fin: base.Add(time.Hour),
	}
}

func datosValidos() DatosIncidente {
	return DatosIncidente{
		EscolarID: "escolar-1", ReportadoPor: "escolar-1", ReservaID: "reserva-1",
		Tipo: TipoTratoInadecuado, Descripcion: strings.Repeat("descripción válida ", 3),
		Evidencias: []Archivo{{Nombre: "evidencia.pdf", Contenido: pdfValido}},
	}
}

func TestPrepararEvidenciaAceptaFormatosPermitidos(t *testing.T) {
	ahora := time.Now().UTC()
	casos := []struct {
		nombre    string
		contenido []byte
		tipo      string
	}{
		{"evidencia.pdf", pdfValido, "application/pdf"},
		{"evidencia.png", append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, []byte("datos")...), "image/png"},
		{"evidencia.jpg", append([]byte{0xff, 0xd8, 0xff, 0xe0}, []byte("datos")...), "image/jpeg"},
	}
	for _, caso := range casos {
		evidencia, err := PrepararEvidencia(Archivo{Nombre: caso.nombre, Contenido: caso.contenido}, ahora)
		if err != nil {
			t.Fatalf("no se aceptó %s: %v", caso.nombre, err)
		}
		if evidencia.Tipo != caso.tipo || evidencia.Nombre != caso.nombre || evidencia.TamanoBytes <= 0 {
			t.Fatalf("evidencia inesperada: %+v", evidencia)
		}
	}
}

func TestPrepararEvidenciaRechazaFormatosYMetadatosInvalidos(t *testing.T) {
	casos := []Archivo{
		{Nombre: "", Contenido: pdfValido},
		{Nombre: "..", Contenido: pdfValido},
		{Nombre: "evidencia.pdf", Contenido: nil},
		{Nombre: "evidencia.txt", Contenido: []byte("texto")},
		{Nombre: "evidencia.exe", Contenido: []byte("MZ")},
		{Nombre: "evidencia.pdf", Contenido: []byte("no es pdf")},
	}
	for _, archivo := range casos {
		if _, err := PrepararEvidencia(archivo, time.Now().UTC()); err == nil {
			t.Fatalf("se esperaba error para nombre=%q contenido=%q", archivo.Nombre, archivo.Contenido)
		}
	}
}

func TestNuevoIncidenteYRevisión(t *testing.T) {
	base := time.Now().UTC()
	ahora := base.Add(2 * time.Hour)
	incidente, err := NuevoIncidente(datosValidos(), tutoriaFinalizada(base), ahora)
	if err != nil {
		t.Fatalf("no se creó el incidente: %v", err)
	}
	if incidente.Estado != EstadoPendiente || incidente.Resultado != "" ||
		incidente.EscolarNombre != "Escolar Uno" || incidente.FacultadID != "fac-1" {
		t.Fatalf("incidente inesperado: %+v", incidente)
	}
	if len(incidente.Evidencias) != 1 {
		t.Fatalf("se esperaba una evidencia, se obtuvo %d", len(incidente.Evidencias))
	}

	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: ResultadoProcede, Comentario: "Revisión válida"}, ahora); err != nil {
		t.Fatalf("no se pudo revisar: %v", err)
	}
	if incidente.Estado != EstadoRevisado || incidente.Resultado != ResultadoProcede || incidente.RevisadoPor != "coord-1" {
		t.Fatalf("revisión inesperada: %+v", incidente)
	}
	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: ResultadoProcede, Comentario: "Otra"}, ahora); err != ErrYaRevisado {
		t.Fatalf("se esperaba ErrYaRevisado, se obtuvo %v", err)
	}
}

func TestNuevoIncidenteValidaCamposYVentana(t *testing.T) {
	base := time.Now().UTC()
	datos := datosValidos()
	datos.EscolarID = ""
	datos.ReservaID = ""
	datos.Tipo = "desconocido"
	datos.Descripcion = "corta"
	datos.Evidencias = nil
	if _, err := NuevoIncidente(datos, tutoriaFinalizada(base), base.Add(2*time.Hour)); err == nil {
		t.Fatal("se esperaban errores de validación")
	}
	if _, err := NuevoIncidente(datosValidos(), &Tutoria{ReservaID: "otra"}, base.Add(2*time.Hour)); err == nil {
		t.Fatal("una tutoría ajena no debe aceptarse")
	}
	if _, err := NuevoIncidente(datosValidos(), tutoriaFinalizada(base), time.Time{}); err == nil {
		t.Fatal("una fecha cero no debe aceptarse")
	}
}

func TestRevisarValidaDatos(t *testing.T) {
	incidente := &Incidente{Estado: EstadoPendiente}
	if err := incidente.Revisar("", DatosRevision{}, time.Now()); err == nil {
		t.Fatal("un revisor vacío no debe aceptarse")
	}
	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: "otro", Comentario: "abc"}, time.Now()); err == nil {
		t.Fatal("un resultado inválido no debe aceptarse")
	}
	if err := incidente.Revisar("coord-1", DatosRevision{Resultado: ResultadoProcede, Comentario: "abc"}, time.Now()); err == nil {
		t.Fatal("un comentario corto no debe aceptarse")
	}
}
