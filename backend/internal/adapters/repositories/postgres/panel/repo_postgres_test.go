package panel

import (
	"os"
	"regexp"
	"testing"
)

// F-04: comparar columnas UUID con su versión TEXT impide usar los índices de
// UUID y degrada las agregaciones del panel al crecer los datos.
func TestF04MetricasDelPanelNoCastUUIDATextoEnComparaciones(t *testing.T) {
	contenido, err := os.ReadFile("repo_postgres.go")
	if err != nil {
		t.Fatalf("no se pudo leer repo_postgres.go: %v", err)
	}
	comparacionConCast := regexp.MustCompile(
		`(?i)(::text\s*(=|<>|!=)|(=|<>|!=)\s*[a-z_][a-z0-9_.]*::text)`,
	)
	if encontrado := comparacionConCast.FindString(string(contenido)); encontrado != "" {
		t.Fatalf(
			"F-04: el panel compara UUID mediante cast a TEXT (%q); debe comparar columnas UUID directamente",
			encontrado,
		)
	}
}
