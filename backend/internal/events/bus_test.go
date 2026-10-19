package events

import "testing"

func TestObservadorFallidoNoRompeALosDemas(t *testing.T) {
	b := NuevoBus()
	n := 0
	b.Registrar("x", NuevoObservadorFunc("malo", func(Event) { panic("boom") }))
	b.Registrar("x", NuevoObservadorFunc("bueno", func(Event) { n++ }))
	b.Notificar(Event{Nombre: "x"})
	if n != 1 {
		t.Fatal("el segundo observador debía ejecutarse")
	}
}

func TestDesregistrar(t *testing.T) {
	b := NuevoBus()
	n := 0
	o := NuevoObservadorFunc("a", func(Event) { n++ })
	b.Registrar("x", o)
	b.Notificar(Event{Nombre: "x"})
	b.Desregistrar("x", o)
	b.Notificar(Event{Nombre: "x"})
	if n != 1 {
		t.Fatalf("tras desregistrar no debe notificarse; n=%d", n)
	}
}
