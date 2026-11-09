package acceso

import (
	"context"
	"time"
)

type RepositorioUsuarios interface {
	CrearUsuario(ctx context.Context, usuario *Usuario) (*Usuario, error)
	UsuarioPorEmail(ctx context.Context, email string) (*Usuario, error)
	UsuarioPorID(ctx context.Context, usuarioID string) (*Usuario, error)
	MarcarCorreoConfirmado(ctx context.Context, usuarioID string) error
	ActualizarClave(ctx context.Context, usuarioID, hashClave string) error
	RegistrarIngreso(ctx context.Context, usuarioID string, momento time.Time) error
	EliminarUsuariosSinConfirmar(ctx context.Context, ahora time.Time) (int64, error)
}

type RepositorioSesiones interface {
	GuardarSesion(ctx context.Context, sesion Sesion) error
	SesionPorHash(ctx context.Context, hashToken []byte) (*Sesion, error)
	EliminarSesion(ctx context.Context, hashToken []byte) error
	EliminarSesionesDeUsuario(ctx context.Context, usuarioID string) error
}

type RepositorioTokens interface {
	GuardarTokenRecuperacion(ctx context.Context, token TokenRecuperacion) error
	UsuarioDeTokenRecuperacion(ctx context.Context, hashToken []byte, ahora time.Time) (string, error)
	ConsumirTokenRecuperacion(ctx context.Context, hashToken []byte, ahora time.Time) (string, error)
	GuardarTokenConfirmacion(ctx context.Context, token TokenConfirmacion) error
	ConsumirTokenConfirmacion(ctx context.Context, hashToken []byte, ahora time.Time) (string, error)
}

type RepositorioVinculaciones interface {
	GuardarSolicitudVinculacion(ctx context.Context, vinculacion Vinculacion) (*Vinculacion, error)
	SolicitudPendienteDeEscolar(ctx context.Context, escolarID string, ahora time.Time) (*Vinculacion, error)
	SolicitudVinculacionPorToken(ctx context.Context, consulta ConsultaVinculacionPorToken) (*Vinculacion, error)
	VinculacionAceptadaDeEscolar(ctx context.Context, escolarID string) (*Vinculacion, error)
	ConsumirSolicitudVinculacion(ctx context.Context, consulta ConsultaVinculacionPorToken) (*Vinculacion, error)
	VinculacionesDeApoderado(ctx context.Context, apoderadoID string, ahora time.Time) ([]Vinculacion, error)
}

type RepositorioAutorizaciones interface {
	AutorizacionVigenteDeEscolar(ctx context.Context, escolarID string) (*Autorizacion, error)
	AutorizacionesDeVinculacion(ctx context.Context, vinculacionID string) ([]Autorizacion, error)
	CambiarAutorizacion(ctx context.Context, cambio CambioAutorizacion) (*Autorizacion, error)
}

type Repositorio interface {
	RepositorioUsuarios
	RepositorioSesiones
	RepositorioTokens
	RepositorioVinculaciones
	RepositorioAutorizaciones
}
