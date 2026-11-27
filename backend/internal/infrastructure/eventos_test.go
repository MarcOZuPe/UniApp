package infra

import (
	"bytes"
	"log"
	"testing"

	modelo "uniapp/internal/domain/acceso"
)

func TestBusNotificaYDesregistra(t *testing.T) {
	bus := NuevoBus()
	recibidos := 0
	observador := NuevoObservadorFunc("contador", func(modelo.Event) { recibidos++ })
	bus.Registrar("evento", observador)
	bus.Notificar(modelo.Event{Nombre: "evento"})
	if recibidos != 1 {
		t.Fatalf("se esperaba 1 notificación, se recibieron %d", recibidos)
	}
	bus.Desregistrar("evento", observador)
	bus.Notificar(modelo.Event{Nombre: "evento"})
	if recibidos != 1 {
		t.Fatalf("no debería notificar tras desregistrar: %d", recibidos)
	}
	bus.Notificar(modelo.Event{Nombre: "otro"})
	if recibidos != 1 {
		t.Fatalf("no debería notificar eventos distintos: %d", recibidos)
	}
}

func TestBusAislaPanicos(t *testing.T) {
	bus := NuevoBus()
	bus.Registrar("evento", NuevoObservadorFunc("panico", func(modelo.Event) { panic("fallo") }))
	recibidos := 0
	bus.Registrar("evento", NuevoObservadorFunc("sano", func(modelo.Event) { recibidos++ }))
	bus.Notificar(modelo.Event{Nombre: "evento"})
	if recibidos != 1 {
		t.Fatal("un observador no debe impedir a los demás")
	}
}

func TestAuditoriaObservador(t *testing.T) {
	var salida bytes.Buffer
	auditoria := NuevaAuditoria(log.New(&salida, "", 0))
	auditoria.Actualizar(modelo.Event{Nombre: "ingreso", Payload: map[string]any{"usuario_id": "u-1", "rol": "escolar"}})
	if auditoria.ID() != "auditoria" || salida.Len() == 0 {
		t.Fatalf("auditoría inesperada: %q", salida.String())
	}
}
