package main

import (
	"testing"
	"time"
)

func TestConfiguracionRecordatorio(t *testing.T) {
	for _, c := range []struct {
		texto    string
		duracion time.Duration
		falla    bool
	}{{"", 24 * time.Hour, false}, {"2h", 2 * time.Hour, false}, {"0s", 0, true}, {"-1h", 0, true}, {"invalido", 0, true}} {
		t.Setenv("I10_REMINDER_BEFORE", c.texto)
		d, err := anticipacionRecordatorio()
		if (err != nil) != c.falla || (!c.falla && d != c.duracion) {
			t.Fatalf("configuración %q: %v %v", c.texto, d, err)
		}
	}
}
