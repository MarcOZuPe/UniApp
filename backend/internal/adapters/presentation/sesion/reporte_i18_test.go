package sesion

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	modelo "uniapp/internal/application/sesion"
)

var desdeReporte = time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC) // 01/11 00:00 en Lima

// reportePrueba arma un reporte con n sesiones; la primera tiene 45 verificados.
func reportePrueba(t *testing.T, n int, nombreTutor string) *modelo.Reporte {
	t.Helper()
	filtro, err := modelo.NuevoFiltroReporte(desdeReporte, desdeReporte.AddDate(0, 0, 30), "")
	if err != nil {
		t.Fatal(err)
	}
	seudonimos, err := modelo.NuevoSeudonimizador([]byte("clave-de-prueba-i18-suficiente"))
	if err != nil {
		t.Fatal(err)
	}
	var entradas []modelo.EntradaReporte
	for i := 0; i < n; i++ {
		inicio := desdeReporte.Add(time.Duration(i+1) * time.Hour)
		id := fmt.Sprintf("sesion-%03d", i)
		s, err := modelo.NuevaSesion(id, "reserva-"+id, "tutor-a", "escolar-secreto-"+strconv.Itoa(i%3), inicio, inicio.Add(modelo.DuracionSesion))
		if err != nil {
			t.Fatal(err)
		}
		entrada := modelo.EntradaReporte{Sesion: s}
		if i == 0 {
			for _, u := range []string{"tutor-a", s.EscolarID()} {
				if err := s.DeclararAsistencia(u, inicio); err != nil {
					t.Fatal(err)
				}
			}
			r, err := modelo.ProponerMinutos(s, "tutor-a", 45, s.Fin().Add(time.Minute))
			if err != nil {
				t.Fatal(err)
			}
			if err := r.Aprobar("coord", s.Fin().Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
			entrada.Minutos = r
		}
		entradas = append(entradas, entrada)
	}
	reporte, err := modelo.ConstruirReporte("rep-prueba", filtro, "coord", desdeReporte.AddDate(0, 0, 31),
		entradas, map[string]string{"tutor-a": nombreTutor}, seudonimos)
	if err != nil {
		t.Fatal(err)
	}
	return reporte
}

func leerCSV(t *testing.T, contenido []byte) [][]string {
	t.Helper()
	if !bytes.HasPrefix(contenido, []byte(bomUTF8)) {
		t.Fatal("el CSV debe empezar con BOM para Excel")
	}
	lector := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(contenido, []byte(bomUTF8))))
	lector.FieldsPerRecord = -1
	filas, err := lector.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return filas
}

func TestI18CSVContenidoYSeudonimos(t *testing.T) {
	reporte := reportePrueba(t, 3, "Ana Quispe")
	contenido, err := ReporteCSV(reporte)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contenido), "escolar-secreto") {
		t.Fatal("el CSV no puede contener el ID del escolar")
	}
	filas := leerCSV(t, contenido)
	if filas[1][1] != "rep-prueba" || filas[2][1] != "01/11/2026" || filas[3][1] != "30/11/2026" {
		t.Fatalf("metadatos inesperados: %v", filas[:4])
	}
	if filas[7][0] != "aviso" || !strings.Contains(filas[7][1], "Sin validez oficial") {
		t.Fatalf("falta el aviso de RF-18: %v", filas[7])
	}
	// El lector de CSV omite las lineas en blanco: se ubica la cabecera por contenido.
	indice := -1
	for i, fila := range filas {
		if fila[0] == "sesion_id" {
			indice = i
		}
	}
	if indice < 0 {
		t.Fatal("falta la cabecera del detalle")
	}
	cabecera := filas[indice]
	if cabecera[5] != "escolar_seudonimo" || len(cabecera) != 15 {
		t.Fatalf("cabecera inesperada: %v", cabecera)
	}
	primera := filas[indice+1]
	if primera[0] != "sesion-000" || primera[8] != "verificado" || primera[10] != "45" || primera[12] != "aprobar" {
		t.Fatalf("fila inesperada: %v", primera)
	}
	if !strings.HasPrefix(primera[5], modelo.PrefijoSeudonimo) {
		t.Fatalf("el escolar debe ir seudonimizado: %q", primera[5])
	}
	totales := filas[len(filas)-1]
	if totales[0] != "tutor-a" || totales[3] != "45" || totales[8] != "2" {
		t.Fatalf("totales inesperados: %v", totales)
	}
}

func TestI18CSVNeutralizaFormulas(t *testing.T) {
	reporte := reportePrueba(t, 1, `=HYPERLINK("http://malo")`)
	contenido, err := ReporteCSV(reporte)
	if err != nil {
		t.Fatal(err)
	}
	for _, fila := range leerCSV(t, contenido) {
		for _, celda := range fila {
			if celda != "" && strings.ContainsRune("=+-@", rune(celda[0])) {
				t.Fatalf("celda con formula sin neutralizar: %q", celda)
			}
		}
	}
	if !strings.Contains(string(contenido), `'=HYPERLINK`) {
		t.Fatal("la formula debe quedar como texto con ' adelante")
	}
}

func TestI18PDFEstructuraValida(t *testing.T) {
	contenido := ReportePDF(reportePrueba(t, 3, "Ana Quispe"))
	texto := string(contenido)
	if !strings.HasPrefix(texto, "%PDF-1.4\n") || !strings.HasSuffix(texto, "%%EOF\n") {
		t.Fatal("cabecera o fin de PDF invalidos")
	}
	coincidencia := regexp.MustCompile(`startxref\n(\d+)\n`).FindStringSubmatch(texto)
	if coincidencia == nil {
		t.Fatal("falta startxref")
	}
	inicio, _ := strconv.Atoi(coincidencia[1])
	if !strings.HasPrefix(texto[inicio:], "xref\n") {
		t.Fatal("startxref no apunta a la tabla xref")
	}
	// Cada entrada de la xref debe apuntar al inicio de su objeto.
	entradas := regexp.MustCompile(`(\d{10}) 00000 n `).FindAllStringSubmatch(texto[inicio:], -1)
	for i, e := range entradas {
		desplazamiento, _ := strconv.Atoi(e[1])
		if !strings.HasPrefix(texto[desplazamiento:], fmt.Sprintf("%d 0 obj", i+1)) {
			t.Fatalf("la xref del objeto %d no apunta a su inicio", i+1)
		}
	}
	for _, esperado := range []string{"/Count 1", "Sin validez oficial", "gina 1 de 1", "/WinAnsiEncoding"} {
		if !strings.Contains(texto, esperado) {
			t.Errorf("el PDF no contiene %q", esperado)
		}
	}
	if strings.Contains(texto, "escolar-secreto") {
		t.Fatal("el PDF no puede contener el ID del escolar")
	}
}

func TestI18PDFPaginaMuchasFilas(t *testing.T) {
	texto := string(ReportePDF(reportePrueba(t, 120, "Ana Quispe")))
	paginas := regexp.MustCompile(`/Count (\d+)`).FindStringSubmatch(texto)
	total, _ := strconv.Atoi(paginas[1])
	if total < 3 {
		t.Fatalf("120 filas deberian ocupar varias paginas, hay %d", total)
	}
	if !strings.Contains(texto, fmt.Sprintf("gina %d de %d", total, total)) {
		t.Fatal("cada pagina debe numerarse sobre el total")
	}
}

func TestI18PDFReporteVacio(t *testing.T) {
	texto := string(ReportePDF(reportePrueba(t, 0, "")))
	if !strings.Contains(texto, "No hay sesiones en el per") {
		t.Fatal("un reporte vacio debe decirlo")
	}
}

func TestI18CadenaPDFEscapaYCodifica(t *testing.T) {
	if got := cadenaPDF(`a(b)c\d`); got != `a\(b\)c\\d` {
		t.Fatalf("escape inesperado: %q", got)
	}
	if got := cadenaPDF("Peña–día"); got != "Pe\xf1a\x96d\xeda" {
		t.Fatalf("WinAnsi inesperado: %q", got)
	}
	if got := cadenaPDF("日本\n"); got != "?? " {
		t.Fatalf("fuera de WinAnsi debe usar ? y los controles espacio: %q", got)
	}
	// 30 pt a 7 pt de cuerpo: entran 8 caracteres (5 + "...").
	if got := recortarPDF("abcdefghijklmnopqrstuvwxyz", 30, 7); got != "abcde..." {
		t.Fatalf("recorte inesperado: %q", got)
	}
}

func TestI18NombreYTipoDeArchivo(t *testing.T) {
	reporte := reportePrueba(t, 0, "")
	if got := NombreArchivoReporte(reporte, modelo.FormatoPDF); got != "reporte-minutos-20261101-20261130.pdf" {
		t.Fatalf("nombre inesperado: %s", got)
	}
	if TipoContenidoReporte(modelo.FormatoPDF) != "application/pdf" || !strings.HasPrefix(TipoContenidoReporte(modelo.FormatoCSV), "text/csv") {
		t.Fatal("Content-Type inesperado")
	}
	if _, err := RenderizarReporte(reporte, "xlsx"); err == nil {
		t.Fatal("un formato desconocido debe fallar")
	}
}
