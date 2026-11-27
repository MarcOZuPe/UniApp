package sesion

import modelo "uniapp/internal/application/sesion"

// Estrategias concretas del patron Strategy (puerto modelo.ExportadorReporte).
// Agregar un formato nuevo (por ejemplo XLSX) es agregar un tipo aqui y
// registrarlo; el controlador y el caso de uso no cambian (abierto/cerrado).

// ExportadorCSV genera el CSV (UTF-8 con BOM, celdas protegidas contra formulas).
type ExportadorCSV struct{}

func (ExportadorCSV) Formato() modelo.FormatoReporte { return modelo.FormatoCSV }
func (ExportadorCSV) TipoContenido() string          { return "text/csv; charset=utf-8" }
func (ExportadorCSV) Exportar(r *modelo.Reporte) ([]byte, error) {
	return ReporteCSV(r)
}

// ExportadorPDF genera el PDF 1.4 propio, sin dependencias externas.
type ExportadorPDF struct{}

func (ExportadorPDF) Formato() modelo.FormatoReporte { return modelo.FormatoPDF }
func (ExportadorPDF) TipoContenido() string          { return "application/pdf" }
func (ExportadorPDF) Exportar(r *modelo.Reporte) ([]byte, error) {
	return ReportePDF(r), nil
}

// exportadores es el registro de estrategias disponibles, por formato.
var exportadores = registrarExportadores(ExportadorCSV{}, ExportadorPDF{})

func registrarExportadores(lista ...modelo.ExportadorReporte) map[modelo.FormatoReporte]modelo.ExportadorReporte {
	registro := make(map[modelo.FormatoReporte]modelo.ExportadorReporte, len(lista))
	for _, exportador := range lista {
		registro[exportador.Formato()] = exportador
	}
	return registro
}

// ExportadorPara elige la estrategia del formato pedido.
func ExportadorPara(formato modelo.FormatoReporte) (modelo.ExportadorReporte, error) {
	exportador, existe := exportadores[formato]
	if !existe {
		return nil, modelo.ErrFormatoReporteInvalido
	}
	return exportador, nil
}
