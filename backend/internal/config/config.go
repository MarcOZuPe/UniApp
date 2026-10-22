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
	Puerto               string
	DatabaseURL          string
	OrigenFrontend       string
	OrigenesFrontend     []string
	Entorno              string
	CookieSegura         bool
	DominiosTutor        string
	SembrarSintetico     bool
	ClaveSemilla         string
	DirectorioDocumentos string
	SMTP                 SMTP
}

func (c Config) EnProduccion() bool {
	switch strings.ToLower(strings.TrimSpace(c.Entorno)) {
	case "prod", "production", "produccion", "producción":
		return true
	}
	return false
}

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

func origenesDe(texto string) []string {
	origenes := make([]string, 0, 2)
	for _, parte := range strings.Split(texto, ",") {
		origen := strings.TrimSpace(parte)
		if origen == "" {
			continue
		}
		origenes = append(origenes, strings.TrimRight(origen, "/"))
	}
	return origenes
}

func validarOrigenes(origenes []string, produccion bool) error {
	if len(origenes) == 0 {
		return fmt.Errorf("FRONTEND_ORIGIN debe indicar al menos un origen")
	}
	for _, origen := range origenes {
		if err := validarOrigen(origen, produccion); err != nil {
			return err
		}
	}
	return nil
}

func variableEntorno(clave, porDefecto string) string {
	if valor := os.Getenv(clave); valor != "" {
		return valor
	}
	return porDefecto
}

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
		Puerto:               variableEntorno("PORT", "8080"),
		DatabaseURL:          os.Getenv("DATABASE_URL"),
		Entorno:              variableEntorno("APP_ENV", "dev"),
		DominiosTutor:        variableEntorno("TUTOR_EMAIL_DOMAIN", "uni.edu.pe,uni.pe"),
		ClaveSemilla:         os.Getenv("SEED_PASSWORD"),
		DirectorioDocumentos: variableEntorno("TUTOR_DOCUMENTS_DIR", "./data/tutor-documents"),
	}
	origenes := origenesDe(variableEntorno("FRONTEND_ORIGIN", "http://localhost:5173"))
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
	if err := validarOrigenes(origenes, configuracion.EnProduccion()); err != nil {
		return configuracion, err
	}
	configuracion.OrigenesFrontend = origenes
	configuracion.OrigenFrontend = origenes[0]
	return configuracion, nil
}
