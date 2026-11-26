package config

import "testing"

func limpiarEntorno(t *testing.T) {
	t.Helper()
	for _, clave := range []string{
		"DATABASE_URL", "FRONTEND_ORIGIN", "APP_ENV", "COOKIE_SECURE",
		"SEED_SYNTHETIC", "SEED_PASSWORD", "SMTP_PASSWORD", "SMTP_PASSWORD_FILE",
		"PORT", "SWAGGER_PORT", "TUTOR_EMAIL_DOMAIN", "TUTOR_DOCUMENTS_DIR",
		"RESERVA_ATTACHMENTS_DIR", "INCIDENT_EVIDENCE_DIR", "RESERVA_ATTACHMENTS_SCAN",
		"SMTP_HOST", "SMTP_PORT", "SMTP_USER", "SMTP_FROM",
	} {
		t.Setenv(clave, "")
	}
}

func entornoMinimo(t *testing.T) {
	t.Helper()
	limpiarEntorno(t)
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/uniapp?sslmode=disable")
	t.Setenv("FRONTEND_ORIGIN", "https://app.uni.edu.pe")
}

// F-05: con APP_ENV=prod no deberían aceptarse semillas sintéticas ni cookies sin Secure.
func TestF05ProduccionRechazaSemillasYCookiesInseguras(t *testing.T) {
	entornoMinimo(t)
	t.Setenv("APP_ENV", "prod")
	t.Setenv("FRONTEND_ORIGIN", "https://app.uni.edu.pe")
	t.Setenv("SEED_SYNTHETIC", "true")
	t.Setenv("SEED_PASSWORD", "clave-conocida")
	t.Setenv("COOKIE_SECURE", "false")
	if _, err := Cargar(); err == nil {
		t.Fatal("F-05: producción no debe aceptar SEED_SYNTHETIC=true ni COOKIE_SECURE=false")
	}
}

// F-23: un booleano de entorno inválido no debe convertirse silenciosamente en false.
func TestF23BooleanosInvalidosNoSeIgnoran(t *testing.T) {
	entornoMinimo(t)
	t.Setenv("COOKIE_SECURE", "sí")
	if _, err := Cargar(); err == nil {
		t.Fatal("F-23: COOKIE_SECURE inválido debe reportarse, no convertirse a false")
	}
}
