package infra

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"

	modelo "uniapp/internal/models/acceso"
)

type LogMailer struct{ MostrarCuerpo bool }

func (l LogMailer) Enviar(_ context.Context, mensaje modelo.Mensaje) error {
	if l.MostrarCuerpo {
		log.Printf("[correo-dev] para=%s asunto=%q\n%s", mensaje.Para, mensaje.Asunto, mensaje.Cuerpo)
	} else {
		log.Printf("[correo-dev] para=%s asunto=%q", mensaje.Para, mensaje.Asunto)
	}
	return nil
}

type EnvioSMTP struct {
	Servidor      string
	Autenticacion smtp.Auth
	Remitente     string
	Destinatarios []string
	Contenido     []byte
}

type transporteSMTP interface {
	EnviarCorreo(envio EnvioSMTP) error
}

type smtpEstandar struct{}

func (smtpEstandar) EnviarCorreo(envio EnvioSMTP) error {
	return smtp.SendMail(envio.Servidor, envio.Autenticacion, envio.Remitente, envio.Destinatarios, envio.Contenido)
}

type SMTPMailer struct {
	Host, Puerto, Usuario, Clave, Remitente string
	transporte                              transporteSMTP
}

func (s SMTPMailer) Enviar(ctx context.Context, mensaje modelo.Mensaje) error {
	if s.transporte == nil {
		return enviarSMTPConContexto(ctx, s.construirEnvio(mensaje))
	}
	return s.transporteORespaldo().EnviarCorreo(s.construirEnvio(mensaje))
}

func (s SMTPMailer) transporteORespaldo() transporteSMTP {
	if s.transporte != nil {
		return s.transporte
	}
	return smtpEstandar{}
}

func (s SMTPMailer) construirEnvio(mensaje modelo.Mensaje) EnvioSMTP {
	return EnvioSMTP{
		Servidor:      s.Host + ":" + s.Puerto,
		Autenticacion: s.autenticacion(),
		Remitente:     s.Remitente,
		Destinatarios: []string{mensaje.Para},
		Contenido:     []byte(construirMensajeRFC(s.Remitente, mensaje)),
	}
}

func (s SMTPMailer) autenticacion() smtp.Auth {
	if s.Usuario == "" {
		return nil
	}
	return smtp.PlainAuth("", s.Usuario, s.Clave, s.Host)
}

func construirMensajeRFC(remitente string, mensaje modelo.Mensaje) string {
	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		remitente, mensaje.Para, mensaje.Asunto, mensaje.Cuerpo)
}

type ReintentoMailer struct {
	Siguiente modelo.Mailer
	Intentos  int
	Espera    time.Duration
}

func NuevoReintentoMailer(mailer modelo.Mailer, intentos int, espera time.Duration) modelo.Mailer {
	return ReintentoMailer{Siguiente: mailer, Intentos: intentos, Espera: espera}
}

func (r ReintentoMailer) Enviar(ctx context.Context, mensaje modelo.Mensaje) error {
	var ultimoError error
	for intento := 0; intento < r.Intentos; intento++ {
		if ultimoError = r.Siguiente.Enviar(ctx, mensaje); ultimoError == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.Espera * time.Duration(intento+1)):
		}
	}
	return ultimoError
}

type TiempoLimiteMailer struct {
	Siguiente modelo.Mailer
	Limite    time.Duration
}

func NuevoTiempoLimiteMailer(mailer modelo.Mailer, limite time.Duration) modelo.Mailer {
	return TiempoLimiteMailer{Siguiente: mailer, Limite: limite}
}

func (t TiempoLimiteMailer) Enviar(ctx context.Context, mensaje modelo.Mensaje) error {
	contextoConLimite, cancelar := context.WithTimeout(ctx, t.Limite)
	defer cancelar()
	return t.Siguiente.Enviar(contextoConLimite, mensaje)
}
