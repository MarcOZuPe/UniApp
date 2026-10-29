package acceso

import (
	"context"
	"errors"
	"strings"
	"time"
)

const duracionRestablecimiento = time.Hour

func (s *Servicio) SolicitarRecuperacion(ctx context.Context, email string) error {
	email = normalizarEmail(email)
	usuario, err := s.repositorio.UsuarioPorEmail(ctx, email)
	if errors.Is(err, ErrNoEncontrado) || (err == nil && (!usuario.Activo || !usuario.EmailConfirmado)) {
		return nil
	}
	if err != nil {
		return err
	}
	token, tokenHash, err := nuevoToken()
	if err != nil {
		return err
	}
	tokenRecuperacion := TokenRecuperacion{TokenHash: tokenHash, UsuarioID: usuario.ID, ExpiraEn: s.ahora().Add(duracionRestablecimiento)}
	if err := s.repositorio.GuardarTokenRecuperacion(ctx, tokenRecuperacion); err != nil {
		return err
	}
	s.bus.Notificar(Event{Nombre: RecuperacionSolicitada, Payload: map[string]any{
		"usuario_id": usuario.ID, "rol": string(usuario.Rol), "email": usuario.Email, "token": token}})
	return nil
}

func (s *Servicio) RestablecerClave(ctx context.Context, token, nuevaClave string) error {
	tokenHash := hashToken(token)
	idUsuario, err := s.repositorio.UsuarioDeTokenRecuperacion(ctx, tokenHash, s.ahora())
	if err != nil {
		return err
	}
	usuario, err := s.repositorio.UsuarioPorID(ctx, idUsuario)
	if err != nil {
		return ErrTokenInvalido
	}
	if err := s.validarNuevaClave(usuario, nuevaClave); err != nil {
		return err
	}
	return s.consumirYAplicar(ctx, tokenHash, usuario, nuevaClave)
}

func (s *Servicio) validarNuevaClave(usuario *Usuario, clave string) error {
	if observaciones := s.fabrica.ValidadorPara(usuario.Rol).Validar(clave, usuario.Email); len(observaciones) > 0 {
		return ErrValidacion{Campos: map[string]string{"clave": strings.Join(observaciones, "; ")}}
	}
	return nil
}

func (s *Servicio) consumirYAplicar(ctx context.Context, tokenHash []byte, usuario *Usuario, clave string) error {
	if _, err := s.repositorio.ConsumirTokenRecuperacion(ctx, tokenHash, s.ahora()); err != nil {
		return err
	}
	nuevoHash, err := s.hasher.Hash(clave)
	if err != nil {
		return err
	}
	if err := s.repositorio.ActualizarClave(ctx, usuario.ID, nuevoHash); err != nil {
		return err
	}
	_ = s.repositorio.EliminarSesionesDeUsuario(ctx, usuario.ID)
	s.bus.Notificar(eventoDeUsuario(ClaveRestablecida, usuario))
	return nil
}
