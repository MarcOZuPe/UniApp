package acceso

import (
	"context"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
)

// F-01: la confirmación y la limpieza no deben competir por la misma cuenta.
type repositorioCarreraConfirmacion struct {
	*memoria.RepositorioMemoria
	ahora time.Time
}

func (r *repositorioCarreraConfirmacion) ConsumirTokenConfirmacion(
	ctx context.Context,
	hash []byte,
	momento time.Time,
) (string, error) {
	id, err := r.RepositorioMemoria.ConsumirTokenConfirmacion(ctx, hash, momento)
	if err == nil {
		_, _ = r.RepositorioMemoria.EliminarUsuariosSinConfirmar(ctx, r.ahora.Add(2*time.Hour))
	}
	return id, err
}

func (r *repositorioCarreraConfirmacion) MarcarCorreoConfirmado(ctx context.Context, id string) error {
	_ = r.RepositorioMemoria.MarcarCorreoConfirmado(ctx, id)
	return nil
}

// B-14 / F-01: el UPDATE de confirmación puede afectar 0 filas si la cuenta
// desapareció. La API no debe responder como si la confirmación hubiera
// funcionado.
type repositorioConfirmacionSinFilas struct {
	*memoria.RepositorioMemoria
}

func (r *repositorioConfirmacionSinFilas) MarcarCorreoConfirmado(context.Context, string) error {
	// Igual que PostgreSQL cuando el UPDATE no encuentra la fila: sin error,
	// pero sin efecto.
	return nil
}

func TestB14ConfirmarCorreoDebeFallarSiElUsuarioYaNoExiste(t *testing.T) {
	base := memoria.NuevoRepositorioMemoria()
	repositorio := &repositorioConfirmacionSinFilas{RepositorioMemoria: base}
	servicio, _ := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()

	token := "token-de-cuenta-inexistente"
	if err := base.GuardarTokenConfirmacion(ctx, TokenConfirmacion{
		TokenHash: servicio.Tokens.Hash(token),
		UsuarioID: "u-inexistente",
		ExpiraEn:  servicio.Ahora().Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}

	if err := servicio.Confirmacion.ConfirmarCorreo(ctx, token); err == nil {
		t.Fatal("B-14: se confirmó una cuenta que ya no existe porque el UPDATE afectó 0 filas y no se comprobó RowsAffected")
	}
}

func TestF01ConfirmacionNoDebePerderCuentaEnCarreraConLimpieza(t *testing.T) {
	base := memoria.NuevoRepositorioMemoria()
	repositorio := &repositorioCarreraConfirmacion{RepositorioMemoria: base, ahora: time.Now().UTC()}
	servicio, bus := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()

	var token string
	capturarToken(bus, ConfirmacionSolicitada, "captura-token-f01", &token)
	if _, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{
		Email: "f01@example.test", NombreUso: "F01", Clave: clavePrueba, Rol: RolEscolar,
	}); err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("no se emitió token de confirmación")
	}

	if err := servicio.Confirmacion.ConfirmarCorreo(ctx, token); err != nil {
		t.Fatalf("F-01: la confirmación debe completar la cuenta: %v", err)
	}
	if _, err := base.UsuarioPorEmail(ctx, "f01@example.test"); err != nil {
		t.Fatalf("F-01: la cuenta no debe desaparecer por la carrera con la limpieza: %v", err)
	}
}
