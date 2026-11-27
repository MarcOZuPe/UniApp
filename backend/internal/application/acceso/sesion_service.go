package acceso

import (
	"context"
)

type SesionService struct {
	*Nucleo
	lockout LockoutManager
}

type CredencialesIngreso struct {
	Email string
	Clave string
}

func (servicio *SesionService) Ingresar(ctx context.Context, credenciales CredencialesIngreso) (string, *Usuario, error) {
	credenciales.Email = normalizarEmail(credenciales.Email)

	bloqueado, err := servicio.lockout.EstaBloqueado(ctx, credenciales.Email)
	if err != nil {
		return "", nil, err
	}
	if bloqueado {
		return "", nil, ErrDemasiadosIntentos
	}

	usuario, err := servicio.autenticar(ctx, credenciales)
	if err != nil {
		_ = servicio.lockout.RegistrarFallo(ctx, credenciales.Email)
		return "", nil, err
	}

	token, err := servicio.abrirSesion(ctx, usuario)
	if err != nil {
		return "", nil, err
	}

	_ = servicio.lockout.LimpiarFallos(ctx, credenciales.Email)
	return token, usuario, nil
}

func (servicio *SesionService) Salir(context.Context, string) error {
	// La sesión es autocontenida: el cliente la elimina borrando su cookie.
	// No existe estado de sesión en el servidor que revocar.
	return nil
}

func (servicio *SesionService) UsuarioDeSesion(ctx context.Context, token string) (*Usuario, error) {
	sesion, err := servicio.TokensSesion.Verificar(token, servicio.Ahora())
	if err != nil {
		return nil, ErrNoEncontrado
	}
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, sesion.UsuarioID)
	if err != nil {
		return nil, err
	}
	if !usuario.Activo || huellaClave(usuario.ClaveHash) != sesion.HuellaClave {
		return nil, ErrNoEncontrado
	}
	return usuario, nil
}

func (servicio *SesionService) autenticar(ctx context.Context, credenciales CredencialesIngreso) (*Usuario, error) {
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorEmail(ctx, credenciales.Email)
	hash := servicio.HashSenuelo
	if err == nil {
		hash = usuario.ClaveHash
	}
	credencialesValidas := servicio.Hasher.Verify(credenciales.Clave, hash)
	if err != nil || !credencialesValidas || !usuario.Activo || !usuario.EmailConfirmado {
		servicio.Bus.Notificar(Event{Nombre: IngresoFallido, Payload: map[string]any{"usuario_id": "-", "rol": "-"}})
		return nil, ErrCredenciales
	}
	return usuario, nil
}

func (servicio *SesionService) abrirSesion(ctx context.Context, usuario *Usuario) (string, error) {
	momento := servicio.Ahora()
	token, err := servicio.TokensSesion.Firmar(usuario, momento.Add(DuracionSesion))
	if err != nil {
		return "", err
	}
	_ = servicio.RepositorioUsuarios.RegistrarIngreso(ctx, usuario.ID, momento)
	servicio.Bus.Notificar(eventoDeUsuario(IngresoExitoso, usuario))
	return token, nil
}
