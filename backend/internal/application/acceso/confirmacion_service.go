package acceso

import (
	"context"
	"errors"
	"time"
)

const duracionConfirmacion = 30 * time.Minute

// LimpiarRegistrosVencidos borra las cuentas sin confirmar que ya no tienen
// ningún enlace vigente, liberando su correo para un nuevo registro.
func (s *Servicio) LimpiarRegistrosVencidos(ctx context.Context) (int64, error) {
	return s.repositorio.EliminarUsuariosSinConfirmar(ctx, s.ahora())
}

func (s *Servicio) SolicitarConfirmacion(ctx context.Context, usuario *Usuario) error {
	if usuario.EmailConfirmado {
		return nil
	}
	token, tokenHash, err := nuevoToken()
	if err != nil {
		return err
	}
	tokenConfirmacion := TokenConfirmacion{TokenHash: tokenHash, UsuarioID: usuario.ID, ExpiraEn: s.ahora().Add(duracionConfirmacion)}
	if err := s.repositorio.GuardarTokenConfirmacion(ctx, tokenConfirmacion); err != nil {
		return err
	}
	s.bus.Notificar(Event{Nombre: ConfirmacionSolicitada, Payload: map[string]any{
		"usuario_id": usuario.ID, "rol": string(usuario.Rol), "email": usuario.Email, "token": token}})
	return nil
}

func (s *Servicio) ReenviarConfirmacion(ctx context.Context, email string) error {
	if _, err := s.LimpiarRegistrosVencidos(ctx); err != nil {
		return err
	}
	usuario, err := s.repositorio.UsuarioPorEmail(ctx, normalizarEmail(email))
	if errors.Is(err, ErrNoEncontrado) || (err == nil && (!usuario.Activo || usuario.EmailConfirmado)) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.SolicitarConfirmacion(ctx, usuario)
}

func (s *Servicio) ConfirmarCorreo(ctx context.Context, token string) error {
	idUsuario, err := s.repositorio.ConsumirTokenConfirmacion(ctx, hashToken(token), s.ahora())
	if err != nil {
		return err
	}
	if err := s.repositorio.MarcarCorreoConfirmado(ctx, idUsuario); err != nil {
		return err
	}
	s.bus.Notificar(Event{Nombre: CorreoConfirmado, Payload: map[string]any{"usuario_id": idUsuario, "rol": "-"}})
	return nil
}
