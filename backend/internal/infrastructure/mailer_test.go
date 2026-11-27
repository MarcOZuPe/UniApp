package infra

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	modelo "uniapp/internal/domain/acceso"
)

type mailerPrueba struct {
	fallos int
	envios int
}

func (m *mailerPrueba) Enviar(context.Context, modelo.Mensaje) error {
	m.envios++
	if m.fallos > 0 {
		m.fallos--
		return errors.New("fallo simulado")
	}
	return nil
}

func TestLogMailerYSMTP(t *testing.T) {
	logMailer := LogMailer{MostrarCuerpo: true}
	if err := logMailer.Enviar(context.Background(), modelo.Mensaje{Para: "a@example.test", Asunto: "Asunto", Cuerpo: "Cuerpo"}); err != nil {
		t.Fatalf("LogMailer falló: %v", err)
	}
	if err := (LogMailer{}).Enviar(context.Background(), modelo.Mensaje{Para: "a@example.test"}); err != nil {
		t.Fatalf("LogMailer sin cuerpo falló: %v", err)
	}
	smtpMailer := SMTPMailer{Host: "smtp.example.test", Puerto: "587", Usuario: "usuario", Clave: "clave", Remitente: "no-responder@example.test"}
	envio := smtpMailer.construirEnvio(modelo.Mensaje{Para: "a@example.test", Asunto: "Asunto", Cuerpo: "Cuerpo"})
	if envio.Servidor != "smtp.example.test:587" || envio.Destinatarios[0] != "a@example.test" || !strings.Contains(string(envio.Contenido), "Subject: Asunto") {
		t.Fatalf("envío SMTP inesperado: %+v", envio)
	}
	if smtpMailer.autenticacion() == nil {
		t.Fatal("debería construir autenticación con usuario")
	}
	if (SMTPMailer{}).autenticacion() != nil {
		t.Fatal("sin usuario no debe haber autenticación")
	}
	if !strings.Contains(construirMensajeRFC("de@example.test", modelo.Mensaje{Para: "a@example.test", Asunto: "S", Cuerpo: "C"}), "From: de@example.test") {
		t.Fatal("el mensaje RFC no contiene remitente")
	}
}

func TestReintentoYTiempoLimiteMailer(t *testing.T) {
	base := &mailerPrueba{fallos: 1}
	reintento := NuevoReintentoMailer(base, 3, time.Millisecond)
	if err := reintento.Enviar(context.Background(), modelo.Mensaje{}); err != nil {
		t.Fatalf("el reintento debería recuperarse: %v", err)
	}
	if base.envios != 2 {
		t.Fatalf("se esperaban 2 envíos, se hicieron %d", base.envios)
	}

	cancelado, cancelar := context.WithCancel(context.Background())
	cancelar()
	if err := (ReintentoMailer{Siguiente: &mailerPrueba{fallos: 1}, Intentos: 2, Espera: time.Hour}).Enviar(cancelado, modelo.Mensaje{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("se esperaba contexto cancelado, se obtuvo %v", err)
	}

	limite := TiempoLimiteMailer{Siguiente: &mailerPrueba{}, Limite: time.Second}
	if err := limite.Enviar(context.Background(), modelo.Mensaje{}); err != nil {
		t.Fatalf("el límite de tiempo no debería fallar: %v", err)
	}
}
