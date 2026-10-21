package acceso

import (
	"context"
	"time"
)

type Repositorio interface {
	CrearUsuario(ctx context.Context, u *Usuario) (*Usuario, error)
	UsuarioPorEmail(ctx context.Context, email string) (*Usuario, error)
	UsuarioPorID(ctx context.Context, id string) (*Usuario, error)
	MarcarCorreoConfirmado(ctx context.Context, id string) error
	ActualizarClave(ctx context.Context, id, hash string) error
	RegistrarIngreso(ctx context.Context, id string, t time.Time) error

	GuardarSesion(ctx context.Context, s Sesion) error
	SesionPorHash(ctx context.Context, hash []byte) (*Sesion, error)
	EliminarSesion(ctx context.Context, hash []byte) error
	EliminarSesionesDeUsuario(ctx context.Context, id string) error

	GuardarTokenRecuperacion(ctx context.Context, t TokenRecuperacion) error

	UsuarioDeTokenRecuperacion(ctx context.Context, hash []byte, ahora time.Time) (string, error)

	ConsumirTokenRecuperacion(ctx context.Context, hash []byte, ahora time.Time) (string, error)

	GuardarTokenConfirmacion(ctx context.Context, t TokenConfirmacion) error

	ConsumirTokenConfirmacion(ctx context.Context, hash []byte, ahora time.Time) (string, error)

	EliminarUsuariosSinConfirmar(ctx context.Context, ahora time.Time) (int64, error)
}
