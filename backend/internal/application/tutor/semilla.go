package tutor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"
)

const (
	PresentacionSintetica    = "Tutor sintético de demostración con experiencia en acompañamiento académico."
	CursosSinteticos         = "Álgebra, geometría y comunicación"
	ExperienciaSintetica     = "Experiencia sintética para demostrar el flujo de revisión de expedientes."
	nombreDocumentoSintetico = "constancia-sintetica.pdf"
)

func SembrarExpedienteSintetico(ctx context.Context, servicio *Servicio, tutorID string) error {
	expediente, err := servicio.MiExpediente(ctx, tutorID)
	if err == nil {
		if !requiereRepararSemilla(ctx, servicio, expediente) {
			return nil
		}
	} else if !errors.Is(err, ErrNoEncontrado) {
		return err
	}

	_, err = servicio.Presentar(ctx, tutorID, datosSinteticos())
	return err
}

func datosSinteticos() DatosExpediente {
	return DatosExpediente{
		Presentacion: PresentacionSintetica,
		Cursos:       CursosSinteticos,
		Experiencia:  ExperienciaSintetica,
		Documentos: []Archivo{{
			Nombre:    nombreDocumentoSintetico,
			Contenido: pdfSintetico(),
		}},
	}
}

func requiereRepararSemilla(ctx context.Context, servicio *Servicio, expediente *Expediente) bool {
	if expediente.Presentacion != PresentacionSintetica || len(expediente.Documentos) != 1 {
		return false
	}
	documento := expediente.Documentos[0]
	if documento.Nombre != nombreDocumentoSintetico {
		return false
	}
	contenido, err := servicio.almacen.Leer(ctx, documento.ID)
	if err != nil {
		return true
	}
	return !pdfSinteticoValido(contenido)
}

func pdfSinteticoValido(contenido []byte) bool {
	if len(contenido) < 100 ||
		!bytes.HasPrefix(contenido, []byte("%PDF-")) ||
		!bytes.Contains(contenido, []byte("%%EOF")) {
		return false
	}
	indice := bytes.LastIndex(contenido, []byte("startxref"))
	if indice < 0 {
		return false
	}
	resto := bytes.TrimSpace(contenido[indice+len("startxref"):])
	linea := bytes.SplitN(resto, []byte("\n"), 2)[0]
	inicioXref, err := strconv.Atoi(string(linea))
	if err != nil {
		return false
	}
	return inicioXref >= 0 && inicioXref < len(contenido) &&
		bytes.HasPrefix(contenido[inicioXref:], []byte("xref"))
}

func pdfSintetico() []byte {
	contenido := "BT\n/F1 18 Tf\n72 720 Td\n(Constancia sintetica de UniApp) Tj\nET"
	objetos := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(contenido), contenido),
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
	}

	var salida bytes.Buffer
	salida.WriteString("%PDF-1.4\n%\xE2\xE3\xCF\xD3\n")
	offsets := make([]int, 0, len(objetos))
	for indice, objeto := range objetos {
		offsets = append(offsets, salida.Len())
		fmt.Fprintf(&salida, "%d 0 obj\n%s\nendobj\n", indice+1, objeto)
	}

	inicioXref := salida.Len()
	fmt.Fprintf(&salida, "xref\n0 %d\n", len(objetos)+1)
	salida.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		fmt.Fprintf(&salida, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&salida, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objetos)+1, inicioXref)
	return salida.Bytes()
}
