package acceso

import (
	"context"
	"errors"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
)

// repositorioConErrores devuelve errores de base de datos en las consultas que
// F-12 revisa (sesion y usuario), sin dejar de implementar el resto.
type repositorioConErrores struct {
	*memoria.RepositorioMemoria
	errorSesion  error
	errorUsuario error
}

func (r *repositorioConErrores) SesionPorHash(ctx context.Context, hashToken []byte) (*Sesion, error) {
	if r.errorSesion != nil {
		return nil, r.errorSesion
	}
	return r.RepositorioMemoria.SesionPorHash(ctx, hashToken)
}

func (r *repositorioConErrores) UsuarioPorEmail(ctx context.Context, email string) (*Usuario, error) {
	if r.errorUsuario != nil {
		return nil, r.errorUsuario
	}
	return r.RepositorioMemoria.UsuarioPorEmail(ctx, email)
}

// F-12: un fallo de PostgreSQL al recuperar la sesion no es "sesion no
// encontrada"; el error debe propagarse para no diagnosticar mal la caida.
func TestF12UsuarioDeSesionDebePropagarErrorDeRepositorio(t *testing.T) {
	falloBD := errors.New("conexión a PostgreSQL perdida")
	repositorio := &repositorioConErrores{
		RepositorioMemoria: memoria.NuevoRepositorioMemoria(),
		errorSesion:        falloBD,
	}
	servicio, _ := servicioPrueba(t, repositorio, time.Now)

	_, err := servicio.Sesion.UsuarioDeSesion(context.Background(), "token-vigente")
	if !errors.Is(err, falloBD) {
		t.Fatalf("F-12: el error de repositorio se enmascaró como %v; debe propagarse la causa real", err)
	}
}

// F-12: un fallo de base de datos durante la autenticación no debe
// presentarse como "credenciales incorrectas".
func TestF12IngresoDebePropagarErrorDeRepositorio(t *testing.T) {
	falloBD := errors.New("conexión a PostgreSQL perdida")
	repositorio := &repositorioConErrores{
		RepositorioMemoria: memoria.NuevoRepositorioMemoria(),
		errorUsuario:       falloBD,
	}
	servicio, _ := servicioPrueba(t, repositorio, time.Now)

	_, _, err := servicio.Sesion.Ingresar(context.Background(), CredencialesIngreso{
		Email: "f12@example.test", Clave: clavePrueba,
	})
	if !errors.Is(err, falloBD) {
		t.Fatalf("F-12: el fallo de BD se presentó como %v; no debe confundirse con credenciales inválidas", err)
	}
}

// F-12 (consecuencia): los fallos de BD no son intentos fallidos de ingreso y
// no deben consumir el lockout hasta bloquear a un usuario válido.
func TestF12ErrorDeBDNoDebeBloquearLaCuenta(t *testing.T) {
	falloBD := errors.New("conexión a PostgreSQL perdida")
	repositorio := &repositorioConErrores{
		RepositorioMemoria: memoria.NuevoRepositorioMemoria(),
		errorUsuario:       falloBD,
	}
	servicio, _ := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()

	for i := 0; i < 6; i++ {
		if _, _, err := servicio.Sesion.Ingresar(ctx, CredencialesIngreso{
			Email: "f12-lock@example.test", Clave: clavePrueba,
		}); errors.Is(err, ErrDemasiadosIntentos) {
			t.Fatalf("F-12: tras %d fallos de BD la cuenta quedó bloqueada por lockout; una caída no es un intento de fuerza bruta", i)
		}
	}
}
