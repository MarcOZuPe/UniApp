package tutor

import (
	"strings"
	"testing"
	"time"
)

func datosExpedienteValidos() DatosExpediente {
	return DatosExpediente{
		Presentacion: "Docente de matemáticas con experiencia.",
		Cursos:       "Álgebra y geometría",
		Experiencia:  "Cinco años enseñando en secundaria.",
		Documentos:   []Archivo{{Nombre: "titulo.pdf", Contenido: []byte("%PDF-1.4 contenido")}},
	}
}

func TestNuevoExpediente(t *testing.T) {
	ahora := time.Now().UTC()
	expediente, err := NuevoExpediente("tutor-1", datosExpedienteValidos(), ahora)
	if err != nil {
		t.Fatalf("expediente válido rechazado: %v", err)
	}
	if expediente.Estado != EstadoPendiente || expediente.PresentadoEn.IsZero() || len(expediente.Documentos) != 1 {
		t.Fatalf("expediente inesperado: %+v", expediente)
	}
	if expediente.Documentos[0].Tipo != "application/pdf" || expediente.Documentos[0].Tamano <= 0 {
		t.Fatalf("documento inesperado: %+v", expediente.Documentos[0])
	}
}

func TestNuevoExpedienteRechazaDatosInvalidos(t *testing.T) {
	casos := []struct {
		tutorID string
		datos   DatosExpediente
		ahora   time.Time
	}{
		{"", datosExpedienteValidos(), time.Now()},
		{"t-1", DatosExpediente{}, time.Now()},
		{"t-1", DatosExpediente{
			Presentacion: "ok", Cursos: "ok", Experiencia: "ok",
			Documentos: []Archivo{{Nombre: "a.txt", Contenido: []byte("texto")}},
		}, time.Now()},
		{"t-1", datosExpedienteValidos(), time.Time{}},
	}
	for _, caso := range casos {
		if _, err := NuevoExpediente(caso.tutorID, caso.datos, caso.ahora); err == nil {
			t.Fatalf("se esperaba error para %+v", caso)
		}
	}
}

func TestRevisarExpediente(t *testing.T) {
	ahora := time.Now().UTC()
	expediente, _ := NuevoExpediente("tutor-1", datosExpedienteValidos(), ahora)
	vencimiento := ahora.Add(30 * 24 * time.Hour)
	if err := expediente.Revisar(EstadoAprobado, "coord-1", "", vencimiento, ahora); err != nil {
		t.Fatalf("aprobar falló: %v", err)
	}
	if expediente.Estado != EstadoAprobado || !expediente.Publicable(ahora) || expediente.RevisadoPor != "coord-1" {
		t.Fatalf("expediente aprobado inesperado: %+v", expediente)
	}
	if expediente.Publicable(vencimiento.Add(time.Hour)) {
		t.Fatal("un expediente vencido no debe ser publicable")
	}
	if err := expediente.Revisar(EstadoAprobado, "coord-1", "", vencimiento, ahora); err == nil {
		t.Fatal("un expediente ya revisado no debe revisarse")
	}

	rechazado, _ := NuevoExpediente("tutor-2", datosExpedienteValidos(), ahora)
	if err := rechazado.Revisar(EstadoRechazado, "coord-1", "", time.Time{}, ahora); err == nil {
		t.Fatal("rechazar exige motivo")
	}
	if err := rechazado.Revisar(EstadoRechazado, "coord-1", "documento ilegible", time.Time{}, ahora); err != nil {
		t.Fatalf("rechazo válido falló: %v", err)
	}
	if rechazado.Publicable(ahora) {
		t.Fatal("un expediente rechazado no es publicable")
	}
	if err := rechazado.Revisar("otro", "coord-1", "", time.Time{}, ahora); err == nil {
		t.Fatal("un estado inválido debe rechazarse")
	}
}

func TestCopiarExpedienteEsProfundo(t *testing.T) {
	if Copiar(nil) != nil {
		t.Fatal("copiar nil debe devolver nil")
	}
	vencimiento := time.Now()
	revision := time.Now()
	original := &Expediente{
		TutorID: "t-1", Documentos: []Documento{{ID: "d-1"}},
		VenceEn: &vencimiento, RevisadoEn: &revision,
	}
	copia := Copiar(original)
	copia.Documentos[0].ID = "d-2"
	*copia.VenceEn = vencimiento.Add(time.Hour)
	*copia.RevisadoEn = revision.Add(time.Hour)
	if original.Documentos[0].ID != "d-1" || !original.VenceEn.Equal(vencimiento) || !original.RevisadoEn.Equal(revision) {
		t.Fatal("la copia no debe compartir datos")
	}
	if nombreSeguro("../") != "documento.pdf" || nombreSeguro("..") != ".." {
		t.Fatal("el nombre seguro debe conservar el comportamiento esperado")
	}
	if !strings.HasSuffix(nombreSeguro(`C:\ruta\a.pdf`), "a.pdf") {
		t.Fatal("el nombre seguro debe extraer el archivo")
	}
}
