package migrations

import (
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
)

func archivosDeMigracion(t *testing.T) []string {
	t.Helper()
	entradas, err := fs.ReadDir(FS, ".")
	if err != nil {
		t.Fatalf("no se pudo leer el FS embebido: %v", err)
	}
	nombres := make([]string, 0, len(entradas))
	for _, entrada := range entradas {
		nombres = append(nombres, entrada.Name())
	}
	return nombres
}

// F-22: el prefijo numérico debe ser único para que el orden sea determinista.
func TestF22PrefijosDeMigracionUnicos(t *testing.T) {
	vistos := map[string]string{}
	for _, nombre := range archivosDeMigracion(t) {
		if !strings.HasSuffix(nombre, ".up.sql") {
			continue
		}
		prefijo := strings.SplitN(nombre, "_", 2)[0]
		if anterior, existe := vistos[prefijo]; existe {
			t.Fatalf("F-22: prefijo %q duplicado en %q y %q", prefijo, anterior, nombre)
		}
		vistos[prefijo] = nombre
	}
}

// B-17: sesion_sesiones.reserva_id debe ser UUID con FK a reservas, igual que
// sesion_outbox, para no necesitar casts TEXT ni aceptar referencias rotas.
func TestB17SesionReservaIDDebeSerUUIDConFK(t *testing.T) {
	contenido, err := fs.ReadFile(FS, "005_sesion.up.sql")
	if err != nil {
		t.Fatalf("no se pudo leer la migración de sesiones: %v", err)
	}
	definicionValida := regexp.MustCompile(
		`(?is)reserva_id\s+UUID\s+NOT NULL[^,]*REFERENCES\s+reservas\s*\(\s*id\s*\)`,
	)
	if !definicionValida.Match(contenido) {
		t.Fatalf(
			"B-17: sesion_sesiones.reserva_id debe declararse UUID NOT NULL REFERENCES reservas(id); se encontró %q",
			lineaDe(string(contenido), "reserva_id"),
		)
	}
}

func lineaDe(contenido, fragmento string) string {
	for _, linea := range strings.Split(contenido, "\n") {
		if strings.Contains(linea, fragmento) {
			return strings.TrimSpace(linea)
		}
	}
	return ""
}

// B-19: el pool de conexiones solo fija el máximo abierto; sin máximo de
// inactivas ni caducidad las conexiones quedan retenidas indefinidamente.
func TestB19PoolDebeConfigurarInactivasYCaducidad(t *testing.T) {
	contenido, err := os.ReadFile("db.go")
	if err != nil {
		t.Fatalf("no se pudo leer db.go: %v", err)
	}
	codigo := string(contenido)
	for _, ajuste := range []string{
		"SetMaxIdleConns",
		"SetConnMaxIdleTime",
		"SetConnMaxLifetime",
	} {
		if !strings.Contains(codigo, ajuste) {
			t.Fatalf("B-19: la configuración del pool no llama a %s", ajuste)
		}
	}
}
