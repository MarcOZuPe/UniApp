package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidarOrigen(t *testing.T) {
	casos := []struct {
		nombre     string
		origen     string
		produccion bool
		quiereErr  bool
	}{
		{"desarrollo con localhost", "http://localhost:5173", false, false},
		{"desarrollo con loopback", "http://127.0.0.1:5173", false, false},
		{"producción con dominio público", "https://uniapp.uni.pe", true, false},
		{"producción con localhost", "http://localhost:5173", true, true},
		{"producción con loopback", "http://127.0.0.1:8080", true, true},
		{"producción con subdominio localhost", "http://app.localhost", true, true},
		{"sin esquema", "uniapp.uni.pe", false, true},
		{"esquema no http", "ftp://uniapp.uni.pe", false, true},
		{"vacío", "", false, true},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			err := validarOrigen(caso.origen, caso.produccion)
			if caso.quiereErr && err == nil {
				t.Fatalf("validarOrigen(%q, %v) debía fallar", caso.origen, caso.produccion)
			}
			if !caso.quiereErr && err != nil {
				t.Fatalf("validarOrigen(%q, %v) no debía fallar: %v", caso.origen, caso.produccion, err)
			}
		})
	}
}

func TestCargarRechazaOrigenLocalEnProduccion(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("APP_ENV", "prod")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:5173")
	if _, err := Cargar(); err == nil {
		t.Fatal("Cargar debía rechazar FRONTEND_ORIGIN local con APP_ENV=prod")
	}
}

func TestCargarAceptaOrigenPublicoEnProduccion(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("APP_ENV", "prod")
	t.Setenv("FRONTEND_ORIGIN", "https://uniapp.uni.pe")
	configuracion, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar no debía fallar: %v", err)
	}
	if !configuracion.EnProduccion() {
		t.Fatal("APP_ENV=prod debía activar EnProduccion")
	}
}

func TestCargarEnDesarrolloPermiteLocalhost(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("APP_ENV", "dev")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:5173")
	configuracion, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar no debía fallar en desarrollo: %v", err)
	}
	if configuracion.EnProduccion() {
		t.Fatal("APP_ENV=dev no debe ser producción")
	}
}

func TestCargarAceptaVariosOrigenes(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("APP_ENV", "dev")
	t.Setenv("FRONTEND_ORIGIN", " http://localhost:5173 , http://203.0.113.10:5173/ ")
	configuracion, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar no debía fallar: %v", err)
	}
	if configuracion.OrigenFrontend != "http://localhost:5173" {
		t.Fatalf("el primer origen es el de los enlaces: %q", configuracion.OrigenFrontend)
	}
	if len(configuracion.OrigenesFrontend) != 2 || configuracion.OrigenesFrontend[1] != "http://203.0.113.10:5173" {
		t.Fatalf("orígenes permitidos: %v", configuracion.OrigenesFrontend)
	}
}

func TestCargarRechazaOrigenInvalidoEnLista(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("APP_ENV", "dev")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:5173,uniapp.uni.pe")
	if _, err := Cargar(); err == nil {
		t.Fatal("debía rechazar un origen sin esquema dentro de la lista")
	}
}

func TestCargarRechazaLocalhostEnListaDeProduccion(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("APP_ENV", "prod")
	t.Setenv("FRONTEND_ORIGIN", "https://uniapp.uni.pe,http://localhost:5173")
	if _, err := Cargar(); err == nil {
		t.Fatal("con APP_ENV=prod ningún origen puede ser local")
	}
}

func TestCargarRechazaListaVacia(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("APP_ENV", "dev")
	t.Setenv("FRONTEND_ORIGIN", " , ")
	if _, err := Cargar(); err == nil {
		t.Fatal("debía exigir al menos un origen")
	}
}

func TestVariableSecretaDesdeArchivo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "smtp_password")
	if err := os.WriteFile(ruta, []byte("clave-del-archivo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SMTP_PASSWORD", "clave-de-entorno")
	t.Setenv("SMTP_PASSWORD_FILE", ruta)
	clave, err := variableSecreta("SMTP_PASSWORD")
	if err != nil {
		t.Fatal(err)
	}
	if clave != "clave-del-archivo" {
		t.Fatalf("el archivo debe tener prioridad y sin salto final: %q", clave)
	}
}

func TestVariableSecretaUsaEntornoSiNoHayArchivo(t *testing.T) {
	t.Setenv("SMTP_PASSWORD", "clave-de-entorno")
	t.Setenv("SMTP_PASSWORD_FILE", "")
	clave, err := variableSecreta("SMTP_PASSWORD")
	if err != nil {
		t.Fatal(err)
	}
	if clave != "clave-de-entorno" {
		t.Fatalf("debía usar la variable de entorno: %q", clave)
	}
}

func TestVariableSecretaFallaSiElArchivoNoExiste(t *testing.T) {
	t.Setenv("SMTP_PASSWORD_FILE", filepath.Join(t.TempDir(), "no-existe"))
	if _, err := variableSecreta("SMTP_PASSWORD"); err == nil {
		t.Fatal("debía fallar si SMTP_PASSWORD_FILE no se puede leer")
	}
}

func TestCargarLeeClaveSMTPDeArchivo(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "smtp_password")
	if err := os.WriteFile(ruta, []byte("secreto\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DATABASE_URL", "postgres://u:p@localhost:5432/uniapp?sslmode=disable")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_PASSWORD_FILE", ruta)
	configuracion, err := Cargar()
	if err != nil {
		t.Fatalf("Cargar no debía fallar: %v", err)
	}
	if configuracion.SMTP.Clave != "secreto" {
		t.Fatalf("la clave SMTP debía venir del archivo: %q", configuracion.SMTP.Clave)
	}
}
