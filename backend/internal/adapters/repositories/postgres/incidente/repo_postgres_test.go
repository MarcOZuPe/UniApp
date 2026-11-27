package incidente

import (
	"context"
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/lib/pq"
)

// F-11: un identificador que no es UUID debe tratarse como recurso no encontrado.
func TestF11IncidenteConIDInvalidoDebeSerNoEncontrado(t *testing.T) {
	repositorio := RepositorioPostgres{DB: abrirDBConError(t, &pq.Error{Code: "22P02"})}
	_, err := repositorio.PorID(context.Background(), "no-es-un-uuid")
	if !errors.Is(err, ErrIncidenteNoEncontrado) {
		t.Fatalf("F-11: un ID malformado debe producir ErrIncidenteNoEncontrado, se obtuvo %v", err)
	}
}

// F-04 (ampliación): las consultas de tutorías también comparan UUID mediante
// cast a TEXT, lo que impide usar los índices de reservas y usuarios.
func TestF04TutoriasNoCastUUIDATextoEnComparaciones(t *testing.T) {
	contenido, err := os.ReadFile("repo_postgres.go")
	if err != nil {
		t.Fatalf("no se pudo leer repo_postgres.go: %v", err)
	}
	comparacionConCast := regexp.MustCompile(
		`(?i)(::text\s*(=|<>|!=)|(=|<>|!=)\s*[a-z_][a-z0-9_.]*::text)`,
	)
	if encontrado := comparacionConCast.FindString(string(contenido)); encontrado != "" {
		t.Fatalf(
			"F-04: las consultas de incidentes comparan UUID mediante cast a TEXT (%q)",
			encontrado,
		)
	}
}
