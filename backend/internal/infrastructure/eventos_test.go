package infra

import (
	"testing"

	modelo "uniapp/internal/domain/acceso"
)

func TestObservadorFallidoNoRompeALosDemas(t *testing.T) {
	b := NuevoBus()
	n := 0
	b.Registrar("x", NuevoObservadorFunc("malo", func(modelo.Event) { panic("boom") }))
	b.Registrar("x", NuevoObservadorFunc("bueno", func(modelo.Event) { n++ }))
	b.Notificar(modelo.Event{Nombre: "x"})
	if n != 1 {
		t.Fatal("el segundo observador debía ejecutarse")
	}
}

func TestDesregistrar(t *testing.T) {
	b := NuevoBus()
	n := 0
	o := NuevoObservadorFunc("a", func(modelo.Event) { n++ })
	b.Registrar("x", o)
	b.Notificar(modelo.Event{Nombre: "x"})
	b.Desregistrar("x", o)
	b.Notificar(modelo.Event{Nombre: "x"})
	if n != 1 {
		t.Fatalf("tras desregistrar no debe notificarse; n=%d", n)
	}
}
