package adapters

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestPBKDF2VectoresConocidos(t *testing.T) {
	casos := []struct {
		iter int
		want string
	}{
		{1, "120fb6cffcf8b32c43e7225256c4f837a86548c92ccc35480805987cb70be17b"},
		{2, "ae4d0c95af6b46d32d0adff928f06dd02a303f8ef3c251dfd6e2d85a95474c43"},
	}
	for _, c := range casos {
		if got := hex.EncodeToString(pbkdf2([]byte("password"), []byte("salt"), c.iter, 32)); got != c.want {
			t.Fatalf("iter=%d got %s", c.iter, got)
		}
	}
}

func TestHasherVerifica(t *testing.T) {
	h := PBKDF2Hasher{Iteraciones: 1000}
	a, _ := h.Hash("Clave-Segura-1!")
	b, _ := h.Hash("Clave-Segura-1!")
	if a == b {
		t.Fatal("la sal debe variar")
	}
	if !h.Verify("Clave-Segura-1!", a) || h.Verify("otra", a) || h.Verify("x", "basura") {
		t.Fatal("verificación incorrecta")
	}
}

type mailerQueFalla struct{ fallos, llamadas int }

func (m *mailerQueFalla) Enviar(context.Context, Mensaje) error {
	m.llamadas++
	if m.llamadas <= m.fallos {
		return errors.New("smtp caído")
	}
	return nil
}

func TestReintentoMailerDecorator(t *testing.T) {
	falla := &mailerQueFalla{fallos: 2}
	m := ReintentoMailer{Siguiente: falla, Intentos: 3, Espera: time.Millisecond}
	if err := m.Enviar(context.Background(), Mensaje{}); err != nil || falla.llamadas != 3 {
		t.Fatalf("err=%v llamadas=%d", err, falla.llamadas)
	}
	sigueFallando := &mailerQueFalla{fallos: 9}
	if err := (ReintentoMailer{Siguiente: sigueFallando, Intentos: 2, Espera: time.Millisecond}).Enviar(context.Background(), Mensaje{}); err == nil {
		t.Fatal("debía fallar tras agotar intentos")
	}
}

type transporteFalso struct{ envio EnvioSMTP }

func (f *transporteFalso) EnviarCorreo(envio EnvioSMTP) error {
	f.envio = envio
	return nil
}

func TestSMTPMailerAdaptaAlTransporte(t *testing.T) {
	transporte := &transporteFalso{}
	mailer := SMTPMailer{Host: "smtp.test", Puerto: "587", Remitente: "no@uniapp.test", transporte: transporte}
	if err := mailer.Enviar(context.Background(), Mensaje{Para: "a@b.test", Asunto: "Hola", Cuerpo: "cuerpo"}); err != nil {
		t.Fatal(err)
	}
	envio := transporte.envio
	if envio.Servidor != "smtp.test:587" || envio.Remitente != "no@uniapp.test" ||
		len(envio.Destinatarios) != 1 || envio.Destinatarios[0] != "a@b.test" {
		t.Fatalf("traducción incorrecta: %+v", envio)
	}
	contenido := string(envio.Contenido)
	if !strings.Contains(contenido, "Subject: Hola") || !strings.Contains(contenido, "cuerpo") {
		t.Fatalf("mensaje mal armado: %q", contenido)
	}
}

func TestDecoratorsApilados(t *testing.T) {
	falla := &mailerQueFalla{fallos: 1}
	m := NuevoReintentoMailer(NuevoTiempoLimiteMailer(falla, time.Second), 3, time.Millisecond)
	if err := m.Enviar(context.Background(), Mensaje{}); err != nil || falla.llamadas != 2 {
		t.Fatalf("err=%v llamadas=%d", err, falla.llamadas)
	}
}
