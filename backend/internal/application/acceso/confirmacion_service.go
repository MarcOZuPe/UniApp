package acceso

import (
	"context"
	"errors"
	"time"
)

type ConfirmacionService struct{ *Nucleo }

const duracionConfirmacion = 30 * time.Minute

func (servicio *ConfirmacionService) LimpiarRegistrosVencidos(ctx context.Context) (int64, error) {
	return servicio.RepositorioUsuarios.EliminarUsuariosSinConfirmar(ctx, servicio.Ahora())
}

func (servicio *ConfirmacionService) SolicitarConfirmacion(ctx context.Context, usuario *Usuario) error {
	if usuario.EmailConfirmado {
		return nil
	}
	token, tokenHash, err := servicio.Tokens.Nuevo()
	if err != nil {
		return err
	}
	tokenConfirmacion := TokenConfirmacion{TokenHash: tokenHash, UsuarioID: usuario.ID, ExpiraEn: servicio.Ahora().Add(duracionConfirmacion)}
	if err := servicio.RepositorioTokens.GuardarTokenConfirmacion(ctx, tokenConfirmacion); err != nil {
		return err
	}
	servicio.Bus.Notificar(Event{Nombre: ConfirmacionSolicitada, Payload: map[string]any{
		"usuario_id": usuario.ID, "rol": string(usuario.Rol), "email": usuario.Email, "token": token}})
	return nil
}

func (servicio *ConfirmacionService) ReenviarConfirmacion(ctx context.Context, email string) error {
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorEmail(ctx, normalizarEmail(email))
	if errors.Is(err, ErrNoEncontrado) || (err == nil && (!usuario.Activo || usuario.EmailConfirmado)) {
		return nil
	}
	if err != nil {
		return err
	}
	return servicio.SolicitarConfirmacion(ctx, usuario)
}

func (servicio *ConfirmacionService) ConfirmarCorreo(ctx context.Context, token string) error {
	idUsuario, err := servicio.RepositorioTokens.ConsumirTokenConfirmacion(ctx, servicio.Tokens.Hash(token), servicio.Ahora())
	if err != nil {
		return err
	}
	if err := servicio.RepositorioUsuarios.MarcarCorreoConfirmado(ctx, idUsuario); err != nil {
		return err
	}
	servicio.Bus.Notificar(Event{Nombre: CorreoConfirmado, Payload: map[string]any{"usuario_id": idUsuario, "rol": "-"}})
	return nil
}
