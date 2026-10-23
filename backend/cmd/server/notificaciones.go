package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	"uniapp/internal/config"
	"uniapp/internal/infra"
	acceso "uniapp/internal/models/acceso"
	sesion "uniapp/internal/models/sesion"
)

// Configuración I10: anticipación positiva; 24 h es una decisión configurable, no RF-12.
func anticipacionRecordatorio() (time.Duration, error) {
	texto := os.Getenv("I10_REMINDER_BEFORE")
	if texto == "" {
		return 24 * time.Hour, nil
	}
	d, err := time.ParseDuration(texto)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("I10_REMINDER_BEFORE debe ser una duración positiva, por ejemplo 24h")
	}
	return d, nil
}

func iniciarOutbox(ctx context.Context, db *sql.DB, c config.Config, registro *log.Logger) {
	var mailer acceso.Mailer = infra.LogMailer{MostrarCuerpo: false}
	if c.SMTP.Host != "" {
		mailer = infra.SMTPMailer{Host: c.SMTP.Host, Puerto: c.SMTP.Puerto, Usuario: c.SMTP.Usuario, Clave: c.SMTP.Clave, Remitente: c.SMTP.Remitente}
	}
	procesador := sesion.ProcesadorOutbox{Repositorio: sesion.RepositorioOutboxPostgres{DB: db}, Mailer: mailer}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			for i := 0; i < 50 && ctx.Err() == nil; i++ {
				hay, err := procesador.ProcesarUno(ctx)
				if err != nil {
					registro.Print("I10: fallo de procesamiento; se recuperará el trabajo pendiente")
					break
				}
				if !hay {
					break
				}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}
