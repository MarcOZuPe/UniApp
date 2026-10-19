package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type SMTP struct{ Host, Puerto, Usuario, Clave, Remitente string }

type Config struct {
	Puerto           string
	DatabaseURL      string
	OrigenFrontend   string
	Entorno          string
	CookieSegura     bool
	DominiosTutor    string // uno o varios dominios institucionales separados por coma
	SembrarSintetico bool
	ClaveSemilla     string
	SMTP             SMTP
}

// EnProduccion indica que la configuración es de despliegue real (APP_ENV=prod).
func (c Config) EnProduccion() bool {
	switch strings.ToLower(strings.TrimSpace(c.Entorno)) {
	case "prod", "production", "produccion", "producción":
		return true
	}
	return false
}

// validarOrigen exige que FRONTEND_ORIGIN sea una URL absoluta http(s) y, en
// producción, que no apunte a la máquina local: los enlaces de recuperación
// viajan por correo y deben resolverse desde el navegador del usuario.
func validarOrigen(origen string, produccion bool) error {
	u, err := url.Parse(origen)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("FRONTEND_ORIGIN debe ser una URL absoluta http(s): %q", origen)
	}
	if produccion && hostLocal(u.Hostname()) {
		return fmt.Errorf("FRONTEND_ORIGIN apunta a %s; con APP_ENV=prod debe ser la URL pública del frontend", u.Hostname())
	}
	return nil
}

func hostLocal(host string) bool {
	h := strings.ToLower(host)
	return h == "localhost" || strings.HasSuffix(h, ".localhost") || h == "127.0.0.1" || h == "::1"
}

func variableEntorno(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

// variableSecreta lee un secreto de la variable directa o, para Docker secrets
// y gestores equivalentes, del archivo indicado por <CLAVE>_FILE. El archivo
// tiene prioridad y se le quita el salto de línea final.
func variableSecreta(clave string) (string, error) {
	ruta := os.Getenv(clave + "_FILE")
	if ruta == "" {
		return os.Getenv(clave), nil
	}
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return "", fmt.Errorf("no se pudo leer %s_FILE (%s): %w", clave, ruta, err)
	}
	return strings.TrimRight(string(datos), "\r\n"), nil
}

func Cargar() (Config, error) {
	configuracion := Config{
		Puerto:         variableEntorno("PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		OrigenFrontend: variableEntorno("FRONTEND_ORIGIN", "http://localhost:5173"),
		Entorno:        variableEntorno("APP_ENV", "dev"),
		DominiosTutor:  variableEntorno("TUTOR_EMAIL_DOMAIN", "uni.edu.pe,uni.pe"),
		ClaveSemilla:   os.Getenv("SEED_PASSWORD"),
	}
	configuracion.CookieSegura, _ = strconv.ParseBool(variableEntorno("COOKIE_SECURE", "false"))
	configuracion.SembrarSintetico, _ = strconv.ParseBool(variableEntorno("SEED_SYNTHETIC", "false"))
	claveSMTP, err := variableSecreta("SMTP_PASSWORD")
	if err != nil {
		return configuracion, err
	}
	configuracion.SMTP = SMTP{
		Host: os.Getenv("SMTP_HOST"), Puerto: variableEntorno("SMTP_PORT", "587"),
		Usuario: os.Getenv("SMTP_USER"), Clave: claveSMTP,
		Remitente: variableEntorno("SMTP_FROM", "no-responder@uniapp.test"),
	}
	if configuracion.DatabaseURL == "" {
		return configuracion, fmt.Errorf("DATABASE_URL es obligatoria")
	}
	if configuracion.SembrarSintetico && configuracion.ClaveSemilla == "" {
		return configuracion, fmt.Errorf("SEED_PASSWORD es obligatoria con SEED_SYNTHETIC=true")
	}
	if err := validarOrigen(configuracion.OrigenFrontend, configuracion.EnProduccion()); err != nil {
		return configuracion, err
	}
	return configuracion, nil
}
