package sesion

import (
	"errors"
	"testing"

	modelo "uniapp/internal/application/sesion"
)

// Strategy: cada formato tiene su estrategia y todas cumplen el mismo contrato.
func TestExportadoresSonEstrategiasIntercambiables(t *testing.T) {
	reporte := reportePrueba(t, 3, "Ana Quispe")
	casos := map[modelo.FormatoReporte]struct {
		tipo    string
		prefijo string
	}{
		modelo.FormatoCSV: {"text/csv; charset=utf-8", "\xef\xbb\xbf"},
		modelo.FormatoPDF: {"application/pdf", "%PDF-1.4"},
	}
	for formato, esperado := range casos {
		var exportador modelo.ExportadorReporte
		exportador, err := ExportadorPara(formato)
		if err != nil {
			t.Fatalf("%s: %v", formato, err)
		}
		if exportador.Formato() != formato || exportador.TipoContenido() != esperado.tipo {
			t.Fatalf("%s: estrategia incorrecta %T", formato, exportador)
		}
		contenido, err := exportador.Exportar(reporte)
		if err != nil || len(contenido) < len(esperado.prefijo) || string(contenido[:len(esperado.prefijo)]) != esperado.prefijo {
			t.Fatalf("%s: contenido inesperado (%v)", formato, err)
		}
	}
}

func TestExportadorParaFormatoDesconocido(t *testing.T) {
	if _, err := ExportadorPara("xlsx"); !errors.Is(err, modelo.ErrFormatoReporteInvalido) {
		t.Fatalf("un formato sin estrategia debe rechazarse: %v", err)
	}
}
