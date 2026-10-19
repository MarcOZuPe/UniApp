package config

import (
	"fmt"
	"os"
	"strconv"
)

type SMTP struct{ Host, Puerto, Usuario, Clave, Remitente string }

type Config struct {
	Puerto           string
	DatabaseURL      string
	OrigenFrontend   string
	CookieSegura     bool
	DominioTutor     string
	SembrarSintetico bool
	ClaveSemilla     string
	SMTP             SMTP
}

func variableEntorno(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

func Cargar() (Config, error) {
	configuracion := Config{
		Puerto:         variableEntorno("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		OrigenFrontend: variableEntorno("FRONTEND_ORIGIN", "http://localhost:5173"),
		DominioTutor:   variableEntorno("TUTOR_EMAIL_DOMAIN", "uni.edu.pe"),
		ClaveSemilla:   os.Getenv("SEED_PASSWORD"),
	}
	configuracion.CookieSegura, _ = strconv.ParseBool(variableEntorno("COOKIE_SECURE", "false"))
	configuracion.SembrarSintetico, _ = strconv.ParseBool(variableEntorno("SEED_SYNTHETIC", "false"))
	configuracion.SMTP = SMTP{
		Host: os.Getenv("SMTP_HOST"), Puerto: variableEntorno("SMTP_PORT", "587"),
		Usuario: os.Getenv("SMTP_USER"), Clave: os.Getenv("SMTP_PASSWORD"),
		Remitente: variableEntorno("SMTP_FROM", "no-responder@uniapp.test"),
	}
	if configuracion.DatabaseURL == "" {
		return configuracion, fmt.Errorf("DATABASE_URL es obligatoria")
	}
	if configuracion.SembrarSintetico && configuracion.ClaveSemilla == "" {
		return configuracion, fmt.Errorf("SEED_PASSWORD es obligatoria con SEED_SYNTHETIC=true")
	}
	return configuracion, nil
}
