package sesion

import (
	"os"
	"regexp"
	"testing"
)

// F-04: buscar un usuario por `id::text` inutiliza el índice UUID de usuarios.
func TestF04OutboxNoCastUUIDATextoEnComparaciones(t *testing.T) {
	contenido, err := os.ReadFile("repo_outbox_postgres.go")
	if err != nil {
		t.Fatalf("no se pudo leer repo_outbox_postgres.go: %v", err)
	}
	comparacionConCast := regexp.MustCompile(
		`(?i)(::text\s*(=|<>|!=)|(=|<>|!=)\s*[a-z_][a-z0-9_.]*::text)`,
	)
	if encontrado := comparacionConCast.FindString(string(contenido)); encontrado != "" {
		t.Fatalf(
			"F-04: el outbox compara UUID mediante cast a TEXT (%q); debe comparar columnas UUID directamente",
			encontrado,
		)
	}
}
