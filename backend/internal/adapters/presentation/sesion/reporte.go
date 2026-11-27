package sesion

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	modelo "uniapp/internal/application/sesion"
)

// I18: presentacion del reporte de minutos en CSV y PDF.

// ZonaLima: Peru no usa horario de verano; UTC-5 fijo evita depender de tzdata.
var ZonaLima = time.FixedZone("PET", -5*60*60)

const tituloReporte = "Reporte de minutos verificados"

// bomUTF8 marca el CSV como UTF-8 para que Excel muestre bien las tildes.
var bomUTF8 = string([]byte{0xEF, 0xBB, 0xBF})

var etiquetasAsistencia = map[modelo.AsistenciaReporte]string{
	modelo.AsistenciaCompleta:    "Completa",
	modelo.AsistenciaSoloTutor:   "Solo tutor",
	modelo.AsistenciaSoloEscolar: "Solo escolar",
	modelo.AsistenciaSinDeclarar: "Sin declarar",
	modelo.AsistenciaConAusencia: "Ausencia reportada",
}

var etiquetasEstado = map[modelo.EstadoMinutos]string{
	modelo.MinutosPropuestos:  "Propuesto",
	modelo.MinutosEnDisputa:   "En disputa",
	modelo.MinutosVerificados: "Verificado",
	modelo.MinutosAnulados:    "Anulado",
	modelo.EstadoSinRegistro:  "Sin registro",
}

func etiqueta[T comparable](tabla map[T]string, valor T) string {
	if texto, ok := tabla[valor]; ok {
		return texto
	}
	return fmt.Sprint(valor)
}

func fechaLima(t time.Time) string     { return t.In(ZonaLima).Format("02/01/2006") }
func fechaHoraLima(t time.Time) string { return t.In(ZonaLima).Format("02/01/2006 15:04") }

// ultimoDiaLima muestra "hasta" como dia inclusivo (el filtro es semiabierto).
func ultimoDiaLima(hasta time.Time) string { return fechaLima(hasta.Add(-time.Nanosecond)) }

func filtroTutor(r *modelo.Reporte) string {
	if r.Filtro.TutorID() == "" {
		return "Todos"
	}
	return r.Filtro.TutorID()
}

// generadoPor muestra el nombre y conserva el ID para la trazabilidad.
func generadoPor(r *modelo.Reporte) string {
	if r.GeneradoPorNombre == "" {
		return r.GeneradoPor
	}
	return fmt.Sprintf("%s (%s)", r.GeneradoPorNombre, r.GeneradoPor)
}

func siNo(v bool) string {
	if v {
		return "Sí"
	}
	return "No"
}

// NombreArchivoReporte: reporte-minutos-AAAAMMDD-AAAAMMDD.<formato>
func NombreArchivoReporte(r *modelo.Reporte, formato modelo.FormatoReporte) string {
	desde := r.Filtro.Desde().In(ZonaLima).Format("20060102")
	hasta := r.Filtro.Hasta().Add(-time.Nanosecond).In(ZonaLima).Format("20060102")
	return fmt.Sprintf("reporte-minutos-%s-%s.%s", desde, hasta, formato)
}

// TipoContenidoReporte devuelve el Content-Type del formato.
func TipoContenidoReporte(formato modelo.FormatoReporte) string {
	if exportador, err := ExportadorPara(formato); err == nil {
		return exportador.TipoContenido()
	}
	return ExportadorCSV{}.TipoContenido()
}

// RenderizarReporte produce el archivo con la estrategia del formato pedido.
func RenderizarReporte(r *modelo.Reporte, formato modelo.FormatoReporte) ([]byte, error) {
	exportador, err := ExportadorPara(formato)
	if err != nil {
		return nil, err
	}
	return exportador.Exportar(r)
}

// celdaCSV evita la inyeccion de formulas al abrir el archivo en una hoja de
// calculo: una celda que empieza con = + - @ (o tab/CR) se antepone con '.
func celdaCSV(valor string) string {
	if valor != "" && strings.ContainsRune("=+-@\t\r", rune(valor[0])) {
		return "'" + valor
	}
	return valor
}

// ReporteCSV: UTF-8 con BOM (para Excel); metadatos, detalle y totales.
func ReporteCSV(r *modelo.Reporte) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(bomUTF8)
	w := csv.NewWriter(&b)
	escribir := func(celdas ...string) {
		for i := range celdas {
			celdas[i] = celdaCSV(celdas[i])
		}
		_ = w.Write(celdas)
	}

	escribir(tituloReporte + " - UniApp")
	escribir("reporte_id", r.ID)
	escribir("periodo_desde", fechaLima(r.Filtro.Desde()))
	escribir("periodo_hasta", ultimoDiaLima(r.Filtro.Hasta()))
	escribir("tutor", filtroTutor(r))
	escribir("generado_por", generadoPor(r))
	escribir("generado_en", fechaHoraLima(r.GeneradoEn)+" (hora de Lima)")
	escribir("aviso", r.Aviso)
	escribir()

	escribir("sesion_id", "reserva_id", "inicio_lima", "tutor_id", "tutor_nombre", "escolar_seudonimo",
		"asistencia", "resultado_registrado", "estado_minutos", "minutos_propuestos", "minutos_computables",
		"ajustes", "ultima_accion", "ultima_accion_actor", "ultima_accion_en")
	for _, f := range r.Filas {
		accion, actor, en := "", "", ""
		if f.UltimaAccion != nil {
			accion, actor, en = string(f.UltimaAccion.Accion), f.UltimaAccion.Actor, fechaHoraLima(f.UltimaAccion.En)
		}
		escribir(f.SesionID, f.ReservaID, fechaHoraLima(f.Inicio), f.TutorID, f.TutorNombre, f.EscolarSeudonimo,
			string(f.Asistencia), siNo(f.ResultadoRegistrado), string(f.EstadoMinutos),
			strconv.Itoa(f.MinutosPropuestos), strconv.Itoa(f.MinutosComputables), strconv.Itoa(f.Ajustes),
			accion, actor, en)
	}
	escribir()

	escribir("tutor_id", "tutor_nombre", "sesiones", "minutos_verificados",
		"propuestos", "en_disputa", "verificados", "anulados", "sin_registro")
	for _, t := range r.Totales {
		escribir(t.TutorID, t.TutorNombre, strconv.Itoa(t.Sesiones), strconv.Itoa(t.MinutosVerificados),
			strconv.Itoa(t.PorEstado[modelo.MinutosPropuestos]), strconv.Itoa(t.PorEstado[modelo.MinutosEnDisputa]),
			strconv.Itoa(t.PorEstado[modelo.MinutosVerificados]), strconv.Itoa(t.PorEstado[modelo.MinutosAnulados]),
			strconv.Itoa(t.PorEstado[modelo.EstadoSinRegistro]))
	}
	w.Flush()
	return b.Bytes(), w.Error()
}

type columnaPDF struct {
	titulo string
	ancho  float64
	valor  func(modelo.FilaReporte) string
}

var columnasPDF = []columnaPDF{
	{"Inicio (Lima)", 66, func(f modelo.FilaReporte) string { return fechaHoraLima(f.Inicio) }},
	{"Sesión", 74, func(f modelo.FilaReporte) string { return f.SesionID }},
	{"Reserva", 74, func(f modelo.FilaReporte) string { return f.ReservaID }},
	{"Tutor", 96, func(f modelo.FilaReporte) string { return f.TutorNombre }},
	{"Escolar", 62, func(f modelo.FilaReporte) string { return f.EscolarSeudonimo }},
	{"Asistencia", 72, func(f modelo.FilaReporte) string { return etiqueta(etiquetasAsistencia, f.Asistencia) }},
	{"Resultado", 44, func(f modelo.FilaReporte) string { return siNo(f.ResultadoRegistrado) }},
	{"Estado", 56, func(f modelo.FilaReporte) string { return etiqueta(etiquetasEstado, f.EstadoMinutos) }},
	{"Prop.", 30, func(f modelo.FilaReporte) string { return strconv.Itoa(f.MinutosPropuestos) }},
	{"Cuentan", 38, func(f modelo.FilaReporte) string { return strconv.Itoa(f.MinutosComputables) }},
	{"Ajustes", 36, func(f modelo.FilaReporte) string { return strconv.Itoa(f.Ajustes) }},
	{"Última acción", 122, func(f modelo.FilaReporte) string {
		if f.UltimaAccion == nil {
			return "-"
		}
		return fmt.Sprintf("%s · %s · %s", f.UltimaAccion.Accion, f.UltimaAccion.Actor, fechaHoraLima(f.UltimaAccion.En))
	}},
}

const (
	margenPDF         = 36.0
	altoFilaPDF       = 13.0
	tamanoTablaPDF    = 7.0
	limiteInferiorPDF = 48.0
)

// ReportePDF dibuja encabezado, aviso, detalle paginado y totales por tutor.
func ReportePDF(r *modelo.Reporte) []byte {
	d := nuevoDocumentoPDF(tituloReporte + " " + r.ID)
	y := 0.0

	encabezado := func() {
		d.nuevaPagina()
		y = altoPaginaPDF - margenPDF
		d.texto(margenPDF, y-14, 14, true, "UniApp · "+tituloReporte)
		y -= 30
		d.texto(margenPDF, y, 9, false, fmt.Sprintf("Periodo: %s al %s (hora de Lima)   ·   Tutor: %s",
			fechaLima(r.Filtro.Desde()), ultimoDiaLima(r.Filtro.Hasta()), filtroTutor(r)))
		y -= 12
		d.texto(margenPDF, y, 9, false, fmt.Sprintf("Generado por: %s   ·   %s (hora de Lima)   ·   Reporte: %s",
			generadoPor(r), fechaHoraLima(r.GeneradoEn), r.ID))
		y -= 18
		d.rectanguloGris(margenPDF, y-4, anchoPaginaPDF-2*margenPDF, 15, 0.9)
		d.texto(margenPDF+6, y, 8.5, true, "AVISO: "+r.Aviso)
		y -= 22
	}
	cabeceraTabla := func(columnas []columnaPDF) {
		d.rectanguloGris(margenPDF, y-4, anchoPaginaPDF-2*margenPDF, altoFilaPDF, 0.82)
		x := margenPDF
		for _, c := range columnas {
			d.textoAjustado(x+2, y, c.ancho-4, tamanoTablaPDF, true, c.titulo)
			x += c.ancho
		}
		y -= altoFilaPDF
	}
	asegurarEspacio := func(columnas []columnaPDF) {
		if y < limiteInferiorPDF {
			encabezado()
			cabeceraTabla(columnas)
		}
	}

	encabezado()
	d.texto(margenPDF, y, 10, true, fmt.Sprintf("Detalle por sesión (%d)", len(r.Filas)))
	y -= 16
	cabeceraTabla(columnasPDF)
	if len(r.Filas) == 0 {
		d.texto(margenPDF+2, y, 8, false, "No hay sesiones en el período.")
		y -= altoFilaPDF
	}
	for _, f := range r.Filas {
		asegurarEspacio(columnasPDF)
		x := margenPDF
		for _, c := range columnasPDF {
			d.textoAjustado(x+2, y, c.ancho-4, tamanoTablaPDF, false, c.valor(f))
			x += c.ancho
		}
		d.linea(margenPDF, y-4, anchoPaginaPDF-margenPDF, y-4, 0.2)
		y -= altoFilaPDF
	}

	columnasTotales := []columnaPDF{
		{titulo: "Tutor", ancho: 200}, {titulo: "Sesiones", ancho: 70}, {titulo: "Minutos verificados", ancho: 110},
		{titulo: "Propuestos", ancho: 70}, {titulo: "En disputa", ancho: 70}, {titulo: "Verificados", ancho: 70},
		{titulo: "Anulados", ancho: 70}, {titulo: "Sin registro", ancho: 70},
	}
	y -= 14
	if y < limiteInferiorPDF+3*altoFilaPDF {
		encabezado()
	}
	d.texto(margenPDF, y, 10, true, "Totales por tutor (solo suman los minutos verificados)")
	y -= 16
	cabeceraTabla(columnasTotales)
	for _, t := range r.Totales {
		asegurarEspacio(columnasTotales)
		valores := []string{t.TutorNombre, strconv.Itoa(t.Sesiones), strconv.Itoa(t.MinutosVerificados),
			strconv.Itoa(t.PorEstado[modelo.MinutosPropuestos]), strconv.Itoa(t.PorEstado[modelo.MinutosEnDisputa]),
			strconv.Itoa(t.PorEstado[modelo.MinutosVerificados]), strconv.Itoa(t.PorEstado[modelo.MinutosAnulados]),
			strconv.Itoa(t.PorEstado[modelo.EstadoSinRegistro])}
		x := margenPDF
		for i, c := range columnasTotales {
			d.textoAjustado(x+2, y, c.ancho-4, tamanoTablaPDF+1, i == 2, valores[i])
			x += c.ancho
		}
		d.linea(margenPDF, y-4, anchoPaginaPDF-margenPDF, y-4, 0.2)
		y -= altoFilaPDF
	}

	return d.bytes(func(d *documentoPDF, pagina, total int) {
		d.linea(margenPDF, 30, anchoPaginaPDF-margenPDF, 30, 0.3)
		d.texto(margenPDF, 18, 7, false, "Reporte "+r.ID+" · Escolares seudonimizados. La huella SHA-256 de este archivo queda en la constancia de descargas de coordinación.")
		d.texto(anchoPaginaPDF-margenPDF-60, 18, 7, false, fmt.Sprintf("Página %d de %d", pagina, total))
	})
}

// RegistroReporteJSON es la vista de una constancia de descarga.
type RegistroReporteJSON struct {
	ID          string    `json:"id"`
	GeneradoPor string    `json:"generado_por"`
	GeneradoEn  time.Time `json:"generado_en"`
	Desde       time.Time `json:"desde"`
	Hasta       time.Time `json:"hasta"`
	TutorID     string    `json:"tutor_id"`
	Formato     string    `json:"formato"`
	Filas       int       `json:"filas"`
	SHA256      string    `json:"sha256"`
}

func PresentarRegistrosReporte(registros []modelo.RegistroReporte) []RegistroReporteJSON {
	vista := make([]RegistroReporteJSON, 0, len(registros))
	for _, r := range registros {
		vista = append(vista, RegistroReporteJSON{
			ID: r.ID, GeneradoPor: r.GeneradoPor, GeneradoEn: r.GeneradoEn, Desde: r.Desde, Hasta: r.Hasta,
			TutorID: r.TutorID, Formato: string(r.Formato), Filas: r.Filas, SHA256: r.SHA256,
		})
	}
	return vista
}
