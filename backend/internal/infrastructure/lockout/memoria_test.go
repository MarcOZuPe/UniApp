package lockout

import (
	"context"
	"testing"
	"time"
)

var inicioLockout = time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)

// F-08: los fallos antiguos de una cuenta que nunca se bloqueó deben dejar de contar.
func TestF08FallosAntiguosNoSeAcumulanIndefinidamente(t *testing.T) {
	actual := inicioLockout
	memoria := NuevaMemoriaConReloj(func() time.Time { return actual })
	ctx := context.Background()
	email := "ana@uni.edu.pe"
	for intento := 0; intento < maxFallos-1; intento++ {
		_ = memoria.RegistrarFallo(ctx, email)
	}
	actual = inicioLockout.Add(24 * time.Hour)
	_ = memoria.RegistrarFallo(ctx, email)
	if bloqueado, _ := memoria.EstaBloqueado(ctx, email); bloqueado {
		t.Fatal("F-08: un único fallo reciente no debe bloquear contando fallos caducados")
	}
}
