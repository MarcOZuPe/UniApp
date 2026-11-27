package acceso

import (
	"context"
	"errors"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
)

// F-13: si la actualización de la clave falla, el token de recuperación no debe
// quedar consumido.
type repositorioFallaClave struct {
	*memoria.RepositorioMemoria
	fallar bool
}

func (r *repositorioFallaClave) ActualizarClave(ctx context.Context, id, hash string) error {
	if r.fallar {
		return errors.New("fallo simulado al actualizar la clave")
	}
	return r.RepositorioMemoria.ActualizarClave(ctx, id, hash)
}

// F-12: un fallo de BD al recuperar el usuario no es un token inválido; el
// enlace puede seguir siendo válido y la causa real debe propagarse.
type repositorioFallaUsuario struct {
	*memoria.RepositorioMemoria
	fallar bool
}

func (r *repositorioFallaUsuario) UsuarioPorID(ctx context.Context, id string) (*Usuario, error) {
	if r.fallar {
		return nil, errors.New("conexión a PostgreSQL perdida")
	}
	return r.RepositorioMemoria.UsuarioPorID(ctx, id)
}

func TestF12RestablecerClaveDebePropagarErrorDeRepositorio(t *testing.T) {
	repositorio := &repositorioFallaUsuario{RepositorioMemoria: memoria.NuevoRepositorioMemoria()}
	servicio, bus := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()

	var token string
	capturarToken(bus, RecuperacionSolicitada, "captura-token-f12", &token)
	if _, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "f12-rec@example.test", NombreUso: "F12", Clave: clavePrueba, Rol: RolEscolar,
	}); err != nil {
		t.Fatal(err)
	}
	if err := servicio.Recuperacion.SolicitarRecuperacion(ctx, "f12-rec@example.test"); err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("no se emitió token de recuperación")
	}

	repositorio.fallar = true
	err := servicio.Recuperacion.RestablecerClave(ctx, SolicitudRestablecimiento{
		Token: token, NuevaClave: "Nueva-Clave-2026!",
	})
	if err == nil || errors.Is(err, ErrTokenInvalido) {
		t.Fatalf("F-12: un fallo de BD se presentó como token inválido (%v); debe propagarse la causa real", err)
	}
}

func TestF13TokenRecuperacionDebeSobrevivirAFalloDeActualizacion(t *testing.T) {
	repositorio := &repositorioFallaClave{RepositorioMemoria: memoria.NuevoRepositorioMemoria()}
	servicio, bus := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()

	var token string
	capturarToken(bus, RecuperacionSolicitada, "captura-token-f13", &token)
	if _, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "f13@example.test", NombreUso: "F13", Clave: clavePrueba, Rol: RolEscolar,
	}); err != nil {
		t.Fatal(err)
	}
	if err := servicio.Recuperacion.SolicitarRecuperacion(ctx, "f13@example.test"); err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("no se emitió token de recuperación")
	}

	repositorio.fallar = true
	if err := servicio.Recuperacion.RestablecerClave(ctx, SolicitudRestablecimiento{
		Token: token, NuevaClave: "Nueva-Clave-2026!",
	}); err == nil {
		t.Fatal("el fallo simulado de actualización debía propagarse")
	}
	repositorio.fallar = false
	if err := servicio.Recuperacion.RestablecerClave(ctx, SolicitudRestablecimiento{
		Token: token, NuevaClave: "Nueva-Clave-2026!",
	}); err != nil {
		t.Fatalf("F-13: el token debe seguir vigente después del fallo para permitir el reintento: %v", err)
	}
}
