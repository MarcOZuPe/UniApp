package sesion

import (
	"context"
	"path/filepath"
	"testing"
)

// El fixture es el contrato provisional con el flujo D (I16): si cambia su
// formato, esta prueba avisa.
func TestI17FixtureCancelaciones(t *testing.T) {
	ruta := filepath.Join("..", "..", "..", "..", "..", "fixtures", "sesion-cancelaciones.json")
	cancelaciones, err := CargarCancelacionesFixture(ruta)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if cancelada, _ := cancelaciones.Cancelada(ctx, "reserva-sintetica-cancelada-001"); !cancelada {
		t.Fatal("la reserva sintetica del fixture debe figurar cancelada")
	}
	if cancelada, _ := cancelaciones.Cancelada(ctx, "otra-reserva"); cancelada {
		t.Fatal("una reserva ausente del fixture no esta cancelada")
	}
	if _, err := CargarCancelacionesFixture("no-existe.json"); err == nil {
		t.Fatal("un fixture inexistente debe fallar")
	}
}
