package sesion

import (
	"bytes"
	"fmt"
	"strings"
)

// Generador PDF 1.4 minimo, solo con la libreria estandar (ADR 0001).
// Usa las fuentes estandar Helvetica y Helvetica-Bold (no se incrustan) con
// codificacion WinAnsi, suficiente para tildes, enie y signos del castellano.

const (
	anchoPaginaPDF = 842.0 // A4 horizontal, en puntos
	altoPaginaPDF  = 595.0
	factorAnchoPDF = 0.52 // ancho medio de un caracter de Helvetica, en em
)

type documentoPDF struct {
	titulo  string
	paginas []*bytes.Buffer
	activa  int // indice de la pagina donde se dibuja
}

func nuevoDocumentoPDF(titulo string) *documentoPDF {
	return &documentoPDF{titulo: titulo}
}

func (d *documentoPDF) nuevaPagina() {
	d.paginas = append(d.paginas, &bytes.Buffer{})
	d.activa = len(d.paginas) - 1
}

func (d *documentoPDF) actual() *bytes.Buffer {
	if len(d.paginas) == 0 {
		d.nuevaPagina()
	}
	return d.paginas[d.activa]
}

// texto escribe una linea con la esquina inferior izquierda en (x, y).
func (d *documentoPDF) texto(x, y, tamano float64, negrita bool, contenido string) {
	fuente := "F1"
	if negrita {
		fuente = "F2"
	}
	fmt.Fprintf(d.actual(), "BT /%s %.1f Tf %.2f %.2f Td (%s) Tj ET\n", fuente, tamano, x, y, cadenaPDF(contenido))
}

// textoAjustado recorta el texto para que quepa en ancho puntos.
func (d *documentoPDF) textoAjustado(x, y, ancho, tamano float64, negrita bool, contenido string) {
	d.texto(x, y, tamano, negrita, recortarPDF(contenido, ancho, tamano))
}

func (d *documentoPDF) linea(x1, y1, x2, y2, grosor float64) {
	fmt.Fprintf(d.actual(), "%.2f w %.2f %.2f m %.2f %.2f l S\n", grosor, x1, y1, x2, y2)
}

// rectanguloGris rellena un rectangulo (x, y es la esquina inferior izquierda).
func (d *documentoPDF) rectanguloGris(x, y, ancho, alto, gris float64) {
	fmt.Fprintf(d.actual(), "q %.2f g %.2f %.2f %.2f %.2f re f Q\n", gris, x, y, ancho, alto)
}

// bytes serializa el documento. pie recibe (pagina, total) y dibuja el pie.
func (d *documentoPDF) bytes(pie func(d *documentoPDF, pagina, total int)) []byte {
	if len(d.paginas) == 0 {
		d.nuevaPagina()
	}
	total := len(d.paginas)
	if pie != nil {
		for i := range d.paginas {
			d.activa = i
			pie(d, i+1, total)
		}
	}

	var salida bytes.Buffer
	var desplazamientos []int
	objeto := func(cuerpo string) {
		desplazamientos = append(desplazamientos, salida.Len())
		fmt.Fprintf(&salida, "%d 0 obj\n%s\nendobj\n", len(desplazamientos), cuerpo)
	}

	salida.WriteString("%PDF-1.4\n%\xe2\xe3\xcf\xd3\n")
	// 1 catalogo, 2 arbol de paginas, 3-4 fuentes, 5 info; luego pagina + contenido.
	objeto("<< /Type /Catalog /Pages 2 0 R >>")
	hijos := make([]string, total)
	for i := range hijos {
		hijos[i] = fmt.Sprintf("%d 0 R", 6+2*i)
	}
	objeto(fmt.Sprintf("<< /Type /Pages /Kids [%s] /Count %d >>", strings.Join(hijos, " "), total))
	objeto("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica /Encoding /WinAnsiEncoding >>")
	objeto("<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica-Bold /Encoding /WinAnsiEncoding >>")
	objeto(fmt.Sprintf("<< /Title (%s) /Producer (UniApp) >>", cadenaPDF(d.titulo)))
	for i, pagina := range d.paginas {
		objeto(fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %.0f %.0f] "+
			"/Resources << /Font << /F1 3 0 R /F2 4 0 R >> >> /Contents %d 0 R >>",
			anchoPaginaPDF, altoPaginaPDF, 7+2*i))
		objeto(fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", pagina.Len(), pagina.String()))
	}

	inicioXref := salida.Len()
	fmt.Fprintf(&salida, "xref\n0 %d\n0000000000 65535 f \n", len(desplazamientos)+1)
	for _, desplazamiento := range desplazamientos {
		fmt.Fprintf(&salida, "%010d 00000 n \n", desplazamiento)
	}
	fmt.Fprintf(&salida, "trailer\n<< /Size %d /Root 1 0 R /Info 5 0 R >>\nstartxref\n%d\n%%%%EOF\n",
		len(desplazamientos)+1, inicioXref)
	return salida.Bytes()
}

// cadenaPDF codifica en WinAnsi y escapa los caracteres especiales de PDF.
func cadenaPDF(texto string) string {
	var b strings.Builder
	for _, r := range texto {
		var c byte
		switch {
		case r == '(' || r == ')' || r == '\\':
			b.WriteByte('\\')
			c = byte(r)
		case r < 0x20:
			c = ' '
		case r < 0x7f:
			c = byte(r)
		case r >= 0xa0 && r <= 0xff:
			c = byte(r)
		default:
			c = winAnsiExtra[r]
			if c == 0 {
				c = '?'
			}
		}
		b.WriteByte(c)
	}
	return b.String()
}

var winAnsiExtra = map[rune]byte{
	'€': 0x80, '…': 0x85, '‘': 0x91, '’': 0x92, '“': 0x93, '”': 0x94,
	'•': 0x95, '–': 0x96, '—': 0x97, '™': 0x99,
}

// recortarPDF deja el texto dentro del ancho, con "..." si no entra.
func recortarPDF(texto string, ancho, tamano float64) string {
	maximo := int(ancho / (tamano * factorAnchoPDF))
	runas := []rune(texto)
	if maximo <= 0 {
		return ""
	}
	if len(runas) <= maximo {
		return texto
	}
	if maximo <= 3 {
		return string(runas[:maximo])
	}
	return string(runas[:maximo-3]) + "..."
}
