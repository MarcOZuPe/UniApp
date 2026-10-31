package infra

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
	modelo "uniapp/internal/domain/acceso"
)

func TestSMTPRespetaContextoDuranteSaludo(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	done := make(chan struct{})
	defer close(done)
	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		<-done
	}()
	host, port, _ := net.SplitHostPort(l.Addr().String())
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	antes := time.Now()
	err = (SMTPMailer{Host: host, Puerto: port, Remitente: "a@example.test"}).Enviar(ctx, modelo.Mensaje{Para: "b@example.test"})
	if err == nil {
		t.Fatal("se esperaba timeout")
	}
	if time.Since(antes) > time.Second {
		t.Fatal("el transporte ignoró el deadline")
	}
}

type mockTransportePrueba struct {
	servidorReal string
}

func (m mockTransportePrueba) EnviarCorreo(envio EnvioSMTP) error {
	return (smtpEstandar{}).EnviarCorreo(envio)
}

func TestSMTPEntregaConTransporteRealLocal(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	recibido := make(chan string, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_ = c.SetDeadline(time.Now().Add(3 * time.Second))
		fmt.Fprint(c, "220 prueba ESMTP\r\n")
		reader := bufio.NewReader(c)
		data := false
		var cuerpo strings.Builder
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			if data {
				if line == ".\r\n" {
					data = false
					recibido <- cuerpo.String()
					fmt.Fprint(c, "250 aceptado\r\n")
				} else {
					cuerpo.WriteString(line)
				}
				continue
			}
			switch {
			case strings.HasPrefix(line, "DATA"):
				data = true
				fmt.Fprint(c, "354 continuar\r\n")
			case strings.HasPrefix(line, "QUIT"):
				fmt.Fprint(c, "221 cierre\r\n")
				return
			default:
				fmt.Fprint(c, "250 OK\r\n")
			}
		}
	}()
	host, port, _ := net.SplitHostPort(l.Addr().String())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	mailer := SMTPMailer{
		Host:       host,
		Puerto:     port,
		Remitente:  "a@example.test",
		transporte: mockTransportePrueba{},
	}

	err = mailer.Enviar(ctx, modelo.Mensaje{Para: "b@example.test", Asunto: "Confirmación", Cuerpo: "Reserva confirmada"})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case contenido := <-recibido:
		if !strings.Contains(contenido, "Reserva confirmada") {
			t.Fatal("contenido perdido")
		}
	case <-ctx.Done():
		t.Fatal("no recibió mensaje")
	}
}
