package reserva

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"
)

// F-11: un ID que no es UUID debe tratarse como reserva inexistente.
func TestF11ReservaConIDInvalidoDebeSerNoEncontrada(t *testing.T) {
	repositorio := RepositorioPostgres{DB: abrirDBConError(t, &pq.Error{Code: "22P02"})}
	_, err := repositorio.PorID(context.Background(), "no-es-un-uuid")
	if !errors.Is(err, ErrReservaNoEncontrada) {
		t.Fatalf("F-11: un ID malformado debe producir ErrReservaNoEncontrada, se obtuvo %v", err)
	}
}
