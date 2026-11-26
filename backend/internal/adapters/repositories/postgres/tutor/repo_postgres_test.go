package tutor

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
)

// F-11: un identificador que no es UUID debe tratarse como expediente no encontrado.
func TestF11ExpedienteConIDInvalidoDebeSerNoEncontrado(t *testing.T) {
	repositorio := RepositorioPostgres{DB: abrirDBConError(t, &pq.Error{Code: "22P02"})}
	_, err := repositorio.PorTutor(context.Background(), "no-es-un-uuid")
	if !errors.Is(err, ErrNoEncontrado) {
		t.Fatalf("F-11: un ID malformado debe producir ErrNoEncontrado, se obtuvo %v", err)
	}
}
