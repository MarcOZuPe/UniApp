package main

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"uniapp/internal/infrastructure/config"
)

func TestPrimerDominio(t *testing.T) {
	casos := map[string]string{
		"":                      "",
		" uni.edu.pe ":          "uni.edu.pe",
		", uni.pe, example.com": "uni.pe",
		" , ":                   "",
	}
	for entrada, esperado := range casos {
		if obtenido := primerDominio(entrada); obtenido != esperado {
			t.Fatalf("primerDominio(%q)=%q, se esperaba %q", entrada, obtenido, esperado)
		}
	}
}

func TestNuevoServidorYOpenAPI(t *testing.T) {
	servidor := nuevoServidor("9090", http.NewServeMux())
	if servidor.Addr != ":9090" || servidor.ReadHeaderTimeout != 5*time.Second ||
		servidor.ReadTimeout != 15*time.Second || servidor.WriteTimeout != 30*time.Second {
		t.Fatalf("servidor inesperado: %+v", servidor)
	}
	t.Setenv("OPENAPI_PATH", " /ruta/openapi.yaml ")
	if ruta := rutaContratoOpenAPI(); ruta != "/ruta/openapi.yaml" {
		t.Fatalf("ruta OpenAPI inesperada: %q", ruta)
	}
	t.Setenv("OPENAPI_PATH", "")
	if ruta := rutaContratoOpenAPI(); ruta == "" {
		t.Fatal("debe devolver una ruta por defecto")
	}
}

func TestAnticipacionRecordatorio(t *testing.T) {
	t.Setenv("I10_REMINDER_BEFORE", "")
	if duracion, err := anticipacionRecordatorio(); err != nil || duracion != 24*time.Hour {
		t.Fatalf("anticipación por defecto inesperada: %v %v", duracion, err)
	}
	t.Setenv("I10_REMINDER_BEFORE", "2h")
	if duracion, err := anticipacionRecordatorio(); err != nil || duracion != 2*time.Hour {
		t.Fatalf("anticipación inesperada: %v %v", duracion, err)
	}
	for _, invalido := range []string{"cero", "0s", "-1h"} {
		t.Setenv("I10_REMINDER_BEFORE", invalido)
		if _, err := anticipacionRecordatorio(); err == nil {
			t.Fatalf("la anticipación %q debe rechazarse", invalido)
		}
	}
}

func TestEstadoReservasI17(t *testing.T) {
	t.Setenv("I17_CANCELACIONES_FIXTURE", "")
	estado, err := estadoReservasI17()
	if err != nil || estado == nil {
		t.Fatalf("estado por defecto inesperado: %v %v", estado, err)
	}
	ruta := filepath.Join(t.TempDir(), "cancelaciones.json")
	if err := os.WriteFile(ruta, []byte(`{"reservas_canceladas":["r-1"]}`), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("I17_CANCELACIONES_FIXTURE", ruta)
	if _, err := estadoReservasI17(); err != nil {
		t.Fatalf("fixture válido falló: %v", err)
	}
	t.Setenv("I17_CANCELACIONES_FIXTURE", "/ruta/inexistente")
	if _, err := estadoReservasI17(); err == nil {
		t.Fatal("un fixture inexistente debe fallar")
	}
}

func TestConstruirMailer(t *testing.T) {
	configuracion := config.Config{SMTP: config.SMTP{Host: "smtp.example.test", Puerto: "587", Usuario: "u", Clave: "c", Remitente: "r"}}
	mailer := construirMailer(configuracion)
	if mailer == nil {
		t.Fatal("debe construir un mailer SMTP")
	}
	vacio := construirMailer(config.Config{})
	if vacio == nil {
		t.Fatal("debe construir un mailer de log")
	}
}
