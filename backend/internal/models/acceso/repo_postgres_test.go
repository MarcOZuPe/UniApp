package acceso_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"uniapp/internal/adapters"
	"uniapp/internal/db"
	"uniapp/internal/events"
	modelo "uniapp/internal/models/acceso"
	"uniapp/migrations"
)

func TestRepositorioPostgresFlujoCompleto(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	ctx := context.Background()
	d, err := db.Abrir(url)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err := db.Migrar(ctx, d, migrations.FS); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrar(ctx, d, migrations.FS); err != nil {
		t.Fatal(err)
	}
	_, _ = d.Exec(`TRUNCATE usuarios CASCADE`)

	h := adapters.PBKDF2Hasher{Iteraciones: 1000}
	s := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: modelo.RepositorioPostgres{DB: d},
		Fabrica:     modelo.NuevaFabrica(h, "uni.edu.pe"),
		Hasher:      h,
		Bus:         events.NuevoBus(),
	})
	if _, err := s.Registrar(ctx, modelo.DatosRegistro{Email: "pg@example.com", NombreUso: "Pg", Clave: "clave-larga-ok", Rol: modelo.RolEscolar}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Registrar(ctx, modelo.DatosRegistro{Email: "PG@example.com", NombreUso: "Pg", Clave: "clave-larga-ok", Rol: modelo.RolEscolar}); !errors.Is(err, modelo.ErrEmailEnUso) {
		t.Fatalf("unicidad: %v", err)
	}
	tok, u, err := s.Ingresar(ctx, "pg@example.com", "clave-larga-ok")
	if err != nil || u.Rol != modelo.RolEscolar {
		t.Fatal(err)
	}
	if got, err := s.UsuarioDeSesion(ctx, tok); err != nil || got.UltimoIngreso == nil {
		t.Fatalf("sesión: %v", err)
	}
	var tokenReset string
	bus := events.NuevoBus()
	bus.Registrar(events.RecuperacionSolicitada, events.NuevoObservadorFunc("captura-token", func(e events.Event) { tokenReset = e.Payload["token"].(string) }))
	s2 := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: modelo.RepositorioPostgres{DB: d},
		Fabrica:     modelo.NuevaFabrica(h, "uni.edu.pe"),
		Hasher:      h,
		Bus:         bus,
	})
	_ = s2.SolicitarRecuperacion(ctx, "pg@example.com")
	if err := s2.RestablecerClave(ctx, tokenReset, "nueva-clave-larga"); err != nil {
		t.Fatal(err)
	}
	if err := s2.RestablecerClave(ctx, tokenReset, "nueva-clave-larga"); !errors.Is(err, modelo.ErrTokenInvalido) {
		t.Fatal("un solo uso")
	}
	if _, err := s.UsuarioDeSesion(ctx, tok); err == nil {
		t.Fatal("sesiones cerradas tras restablecer")
	}
	_ = time.Now
}
