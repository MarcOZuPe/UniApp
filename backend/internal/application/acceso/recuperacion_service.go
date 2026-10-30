package acceso

import (
	"context"
	"errors"
	"strings"
	"time"
)

type RecuperacionService struct{ *Nucleo }

const duracionRestablecimiento = time.Hour

type SolicitudRestablecimiento struct {
	Token      string
	NuevaClave string
}

type solicitudAplicarClave struct {
	tokenHash []byte
	usuario   *Usuario
	clave     string
}

func (servicio *RecuperacionService) SolicitarRecuperacion(ctx context.Context, email string) error {
	usuario, err := servicio.usuarioParaRecuperacion(ctx, email)
	if err != nil || usuario == nil {
		return err
	}
	token, err := servicio.emitirTokenRecuperacion(ctx, usuario)
	if err != nil {
		return err
	}
	servicio.notificarRecuperacion(usuario, token)
	return nil
}

func (servicio *RecuperacionService) RestablecerClave(ctx context.Context, solicitud SolicitudRestablecimiento) error {
	tokenHash := servicio.Tokens.Hash(solicitud.Token)
	idUsuario, err := servicio.RepositorioTokens.UsuarioDeTokenRecuperacion(ctx, tokenHash, servicio.Ahora())
	if err != nil {
		return err
	}
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, idUsuario)
	if err != nil {
		return ErrTokenInvalido
	}
	if err := servicio.validarNuevaClave(usuario, solicitud.NuevaClave); err != nil {
		return err
	}
	return servicio.consumirYAplicar(ctx, solicitudAplicarClave{tokenHash: tokenHash, usuario: usuario, clave: solicitud.NuevaClave})
}

func (servicio *RecuperacionService) usuarioParaRecuperacion(ctx context.Context, email string) (*Usuario, error) {
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorEmail(ctx, normalizarEmail(email))
	if errors.Is(err, ErrNoEncontrado) || (err == nil && (!usuario.Activo || !usuario.EmailConfirmado)) {
		return nil, nil
	}
	return usuario, err
}

func (servicio *RecuperacionService) emitirTokenRecuperacion(ctx context.Context, usuario *Usuario) (string, error) {
	token, tokenHash, err := servicio.Tokens.Nuevo()
	if err != nil {
		return "", err
	}
	solicitud := TokenRecuperacion{
		TokenHash: tokenHash,
		UsuarioID: usuario.ID,
		ExpiraEn:  servicio.Ahora().Add(duracionRestablecimiento),
	}
	if err := servicio.RepositorioTokens.GuardarTokenRecuperacion(ctx, solicitud); err != nil {
		return "", err
	}
	return token, nil
}

func (servicio *RecuperacionService) notificarRecuperacion(usuario *Usuario, token string) {
	servicio.Bus.Notificar(Event{Nombre: RecuperacionSolicitada, Payload: map[string]any{
		"usuario_id": usuario.ID,
		"rol":        string(usuario.Rol),
		"email":      usuario.Email,
		"token":      token,
	}})
}

func (servicio *RecuperacionService) validarNuevaClave(usuario *Usuario, clave string) error {
	if observaciones := servicio.ValidadorClaves.ValidadorPara(usuario.Rol).Validar(clave, usuario.Email); len(observaciones) > 0 {
		return ErrValidacion{Campos: map[string]string{"clave": strings.Join(observaciones, "; ")}}
	}
	return nil
}

func (servicio *RecuperacionService) consumirYAplicar(ctx context.Context, solicitud solicitudAplicarClave) error {
	if _, err := servicio.RepositorioTokens.ConsumirTokenRecuperacion(ctx, solicitud.tokenHash, servicio.Ahora()); err != nil {
		return err
	}
	nuevoHash, err := servicio.Hasher.Hash(solicitud.clave)
	if err != nil {
		return err
	}
	if err := servicio.RepositorioUsuarios.ActualizarClave(ctx, solicitud.usuario.ID, nuevoHash); err != nil {
		return err
	}
	_ = servicio.RepositorioSesiones.EliminarSesionesDeUsuario(ctx, solicitud.usuario.ID)
	servicio.Bus.Notificar(eventoDeUsuario(ClaveRestablecida, solicitud.usuario))
	return nil
}
