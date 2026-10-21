package acceso_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	modelo "uniapp/internal/models/acceso"
	"uniapp/migrations"
)

func TestRepositorioPostgresFlujoCompleto(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	ctx := context.Background()
	d, err := migrations.Abrir(url)
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err := migrations.Migrar(ctx, d, migrations.FS); err != nil {
		t.Fatal(err)
	}
	if err := migrations.Migrar(ctx, d, migrations.FS); err != nil {
		t.Fatal(err)
	}
	_, _ = d.Exec(`TRUNCATE usuarios CASCADE`)

	h := modelo.PBKDF2Hasher{Iteraciones: 1000}
	repo := modelo.RepositorioPostgres{DB: d}
	bus := modelo.NuevoBus()
	var tokenConfirmacion string
	bus.Registrar(modelo.ConfirmacionSolicitada, modelo.NuevoObservadorFunc("captura-confirmacion", func(e modelo.Event) {
		tokenConfirmacion, _ = e.Payload["token"].(string)
	}))
	s := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio: repo,
		Fabrica:     modelo.NuevaFabrica(h, "uni.edu.pe"),
		Hasher:      h,
		Bus:         bus,
	})
	registrado, err := s.Registrar(ctx, modelo.DatosRegistro{Email: "pg@example.com", NombreUso: "Pg", Clave: "clave-larga-ok", Rol: modelo.RolEscolar})
	if err != nil {
		t.Fatal(err)
	}
	if registrado.EmailConfirmado || tokenConfirmacion == "" {
		t.Fatal("el registro público debe nacer sin confirmar y emitir el enlace")
	}
	if _, err := s.Registrar(ctx, modelo.DatosRegistro{Email: "PG@example.com", NombreUso: "Pg", Clave: "clave-larga-ok", Rol: modelo.RolEscolar}); !errors.Is(err, modelo.ErrEmailEnUso) {
		t.Fatalf("unicidad: %v", err)
	}
	if _, _, err := s.Ingresar(ctx, "pg@example.com", "clave-larga-ok"); !errors.Is(err, modelo.ErrCredenciales) {
		t.Fatalf("sin confirmar debe ser invisible: %v", err)
	}
	if err := s.ConfirmarCorreo(ctx, tokenConfirmacion); err != nil {
		t.Fatal(err)
	}
	if err := s.ConfirmarCorreo(ctx, tokenConfirmacion); !errors.Is(err, modelo.ErrTokenInvalido) {
		t.Fatal("el enlace de confirmación es de un solo uso")
	}
	confirmado, err := repo.UsuarioPorEmail(ctx, "pg@example.com")
	if err != nil || !confirmado.EmailConfirmado {
		t.Fatalf("correo_confirmado debe persistir: %v", err)
	}
	tok, u, err := s.Ingresar(ctx, "pg@example.com", "clave-larga-ok")
	if err != nil || u.Rol != modelo.RolEscolar {
		t.Fatal(err)
	}
	if got, err := s.UsuarioDeSesion(ctx, tok); err != nil || got.UltimoIngreso == nil {
		t.Fatalf("sesión: %v", err)
	}

	var tokenReset string
	bus.Registrar(modelo.RecuperacionSolicitada, modelo.NuevoObservadorFunc("captura-token-reset", func(e modelo.Event) {
		tokenReset, _ = e.Payload["token"].(string)
	}))
	_ = s.SolicitarRecuperacion(ctx, "pg@example.com")
	if err := s.RestablecerClave(ctx, tokenReset, "nueva-clave-larga"); err != nil {
		t.Fatal(err)
	}
	if err := s.RestablecerClave(ctx, tokenReset, "nueva-clave-larga"); !errors.Is(err, modelo.ErrTokenInvalido) {
		t.Fatal("un solo uso")
	}
	if _, err := s.UsuarioDeSesion(ctx, tok); err == nil {
		t.Fatal("sesiones cerradas tras restablecer")
	}

	if _, err := s.Registrar(ctx, modelo.DatosRegistro{Email: "vencido@example.com", NombreUso: "Vencido", Clave: "clave-larga-ok", Rol: modelo.RolEscolar}); err != nil {
		t.Fatal(err)
	}
	liberados, err := repo.EliminarUsuariosSinConfirmar(ctx, time.Now().Add(31*time.Minute))
	if err != nil || liberados != 1 {
		t.Fatalf("debía liberarse el registro vencido: %d (%v)", liberados, err)
	}
	if _, err := repo.UsuarioPorEmail(ctx, "vencido@example.com"); !errors.Is(err, modelo.ErrNoEncontrado) {
		t.Fatalf("el correo debe quedar libre: %v", err)
	}
	if confirmado, err := repo.UsuarioPorEmail(ctx, "pg@example.com"); err != nil || !confirmado.EmailConfirmado {
		t.Fatalf("no debe borrar cuentas confirmadas: %v", err)
	}
}
