package infra

import (
	"context"
	"crypto/tls"
	"net"
	"net/smtp"
	"time"
)

// Reutiliza net/smtp, pero limita la conexión real: el contexto solo no detiene SendMail.
func enviarSMTPConContexto(ctx context.Context, e EnvioSMTP) error {
	ctx, cancelar := context.WithTimeout(ctx, 30*time.Second)
	defer cancelar()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", e.Servidor)
	if err != nil {
		return err
	}
	defer conn.Close()
	limite, _ := ctx.Deadline()
	if err := conn.SetDeadline(limite); err != nil {
		return err
	}
	detener := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer detener()
	host, _, err := net.SplitHostPort(e.Servidor)
	if err != nil {
		return err
	}
	cliente, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer cliente.Close()
	if ok, _ := cliente.Extension("STARTTLS"); ok {
		if err := cliente.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if e.Autenticacion != nil {
		if err := cliente.Auth(e.Autenticacion); err != nil {
			return err
		}
	}
	if err := cliente.Mail(e.Remitente); err != nil {
		return err
	}
	for _, destino := range e.Destinatarios {
		if err := cliente.Rcpt(destino); err != nil {
			return err
		}
	}
	writer, err := cliente.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(e.Contenido); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return cliente.Quit()
}
