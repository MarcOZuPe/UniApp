package sesion

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	dominioAcceso "uniapp/internal/domain/acceso"
)

// outboxConFalloSimulaUnLeaseVencido: el primer Resolver falla (como si el
// proceso muriera despues de enviar el correo) y el reclamo vuelve a entregar
// la misma fila. Bajo entrega "al menos una vez" el correo se reenvia.
type outboxConFallo struct {
	mu              sync.Mutex
	entrega         Entrega
	resoluciones    int
	fallosDeResolve int
}

func (o *outboxConFallo) Reclamar(_ context.Context, _ time.Time) (*Entrega, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.entrega.Intentos++
	copia := o.entrega
	return &copia, nil
}

func (o *outboxConFallo) Resolver(_ context.Context, _ Entrega, _ time.Time, _ string) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.resoluciones++
	if o.resoluciones <= o.fallosDeResolve {
		return errors.New("lease vencido: no se pudo confirmar la entrega")
	}
	return nil
}

func (o *outboxConFallo) CorreoDe(_ context.Context, _ string) (string, error) {
	return "destino@example.test", nil
}

type mailerContador struct {
	mu       sync.Mutex
	enviados int
}

func (m *mailerContador) Enviar(context.Context, dominioAcceso.Mensaje) error {
	m.mu.Lock()
	m.enviados++
	m.mu.Unlock()
	return nil
}

// F-14: si falla la resolucion del outbox despues de enviar, el reintento no
// debe volver a enviar el mismo correo al destinatario.
func TestF14OutboxNoDebeReenviarCorreoTrasFalloDeResolucion(t *testing.T) {
	repositorio := &outboxConFallo{
		entrega: Entrega{
			ID: 1, DestinatarioID: "u-1", Canal: "correo",
			Asunto: "Recordatorio", Cuerpo: "Tu sesion empieza pronto",
		},
		fallosDeResolve: 1,
	}
	mailer := &mailerContador{}
	procesador := ProcesadorOutbox{Repositorio: repositorio, Mailer: mailer}

	ctx := context.Background()
	if _, err := procesador.ProcesarUno(ctx); err == nil {
		t.Fatal("el fallo simulado de Resolver debia propagarse")
	}
	// El lease vencio y otro worker (o el mismo) reclama la misma entrega.
	if _, err := procesador.ProcesarUno(ctx); err != nil {
		t.Fatalf("el segundo intento no debia fallar: %v", err)
	}

	if mailer.enviados != 1 {
		t.Fatalf("F-14: el correo se envio %d veces para una sola entrega; un fallo al confirmar no debe duplicar el envio", mailer.enviados)
	}
}
