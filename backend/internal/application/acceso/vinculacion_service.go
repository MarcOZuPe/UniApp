package acceso

import (
	"context"
	"errors"
	"time"
)

type VinculacionService struct{ *Nucleo }

const duracionVinculacion = 24 * time.Hour

func (servicio *VinculacionService) SolicitarVinculacion(ctx context.Context, apoderado *Usuario, emailEscolar string) (*Vinculacion, error) {
	if err := servicio.validarApoderado(apoderado); err != nil {
		return nil, err
	}

	escolar, err := servicio.buscarYValidarEscolar(ctx, emailEscolar)
	if err != nil {
		return nil, err
	}

	if err := servicio.validarEstadoVinculaciones(ctx, escolar.ID); err != nil {
		return nil, err
	}

	token, tokenHash, err := servicio.Tokens.Nuevo()
	if err != nil {
		return nil, err
	}

	vinculacion := servicio.compilarVinculacion(apoderado, escolar, tokenHash)
	creada, err := servicio.RepositorioVinculaciones.GuardarSolicitudVinculacion(ctx, vinculacion)
	if err != nil {
		return nil, err
	}

	servicio.notificarSolicitud(apoderado, escolar, token)
	return creada, nil
}

func (servicio *VinculacionService) ConsultarVinculacion(ctx context.Context, escolar *Usuario, token string) (*Vinculacion, error) {
	if escolar == nil || escolar.Rol != RolEscolar || !escolar.Activo {
		return nil, ErrNoEsEscolar
	}
	if token == "" {
		return nil, ErrTokenInvalido
	}
	vinculacion, err := servicio.RepositorioVinculaciones.SolicitudVinculacionPorToken(ctx, ConsultaVinculacionPorToken{
		TokenHash: servicio.Tokens.Hash(token),
		EscolarID: escolar.ID,
		Ahora:     servicio.Ahora(),
	})
	if errors.Is(err, ErrNoEncontrado) {
		return nil, ErrTokenInvalido
	}
	return vinculacion, err
}

func (servicio *VinculacionService) AceptarVinculacion(ctx context.Context, escolar *Usuario, token string) (*Vinculacion, error) {
	if escolar == nil || escolar.Rol != RolEscolar || !escolar.Activo {
		return nil, ErrNoEsEscolar
	}
	if token == "" {
		return nil, ErrTokenInvalido
	}
	vinculacion, err := servicio.RepositorioVinculaciones.ConsumirSolicitudVinculacion(ctx, ConsultaVinculacionPorToken{
		TokenHash: servicio.Tokens.Hash(token),
		EscolarID: escolar.ID,
		Ahora:     servicio.Ahora(),
	})
	if err != nil {
		return nil, err
	}

	servicio.Bus.Notificar(Event{Nombre: VinculacionAceptada, Payload: map[string]any{
		"usuario_id":       escolar.ID,
		"rol":              string(escolar.Rol),
		"apoderado_id":     vinculacion.ApoderadoID,
		"escolar_id":       vinculacion.EscolarID,
		"apoderado_nombre": vinculacion.ApoderadoNombre,
		"escolar_nombre":   vinculacion.EscolarNombre,
		"apoderado_email":  vinculacion.ApoderadoEmail,
	}})
	return vinculacion, nil
}

func (servicio *VinculacionService) VinculacionesDeApoderado(ctx context.Context, apoderado *Usuario) ([]Vinculacion, error) {
	if err := servicio.validarApoderado(apoderado); err != nil {
		return nil, err
	}
	return servicio.RepositorioVinculaciones.VinculacionesDeApoderado(ctx, apoderado.ID, servicio.Ahora())
}

func (servicio *VinculacionService) VinculacionDeEscolar(ctx context.Context, escolar *Usuario) (*Vinculacion, error) {
	if escolar == nil || escolar.Rol != RolEscolar || !escolar.Activo {
		return nil, ErrNoEsEscolar
	}
	vinculacion, err := servicio.RepositorioVinculaciones.VinculacionAceptadaDeEscolar(ctx, escolar.ID)
	if errors.Is(err, ErrNoEncontrado) {
		return nil, nil
	}
	return vinculacion, err
}

func (servicio *VinculacionService) validarApoderado(apoderado *Usuario) error {
	if apoderado == nil || apoderado.Rol != RolApoderado || !apoderado.Activo {
		return ErrNoEsApoderado
	}
	return nil
}

func (servicio *VinculacionService) buscarYValidarEscolar(ctx context.Context, email string) (*Usuario, error) {
	email = normalizarEmail(email)
	if email == "" {
		return nil, ErrValidacion{Campos: map[string]string{"email_escolar": "el correo es obligatorio"}}
	}

	escolar, err := servicio.RepositorioUsuarios.UsuarioPorEmail(ctx, email)
	if errors.Is(err, ErrNoEncontrado) {
		return nil, ErrEscolarNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	if escolar.Rol != RolEscolar || !escolar.Activo || !escolar.EmailConfirmado {
		return nil, ErrEscolarNoValido
	}
	return escolar, nil
}

func (servicio *VinculacionService) validarEstadoVinculaciones(ctx context.Context, escolarID string) error {
	if _, err := servicio.RepositorioVinculaciones.VinculacionAceptadaDeEscolar(ctx, escolarID); err == nil {
		return ErrYaVinculado
	} else if !errors.Is(err, ErrNoEncontrado) {
		return err
	}

	if _, err := servicio.RepositorioVinculaciones.SolicitudPendienteDeEscolar(ctx, escolarID, servicio.Ahora()); err == nil {
		return ErrSolicitudPendiente
	} else if !errors.Is(err, ErrNoEncontrado) {
		return err
	}
	return nil
}

func (servicio *VinculacionService) compilarVinculacion(apoderado, escolar *Usuario, tokenHash []byte) Vinculacion {
	momento := servicio.Ahora()
	return Vinculacion{
		ApoderadoID:     apoderado.ID,
		EscolarID:       escolar.ID,
		Estado:          EstadoVinculacionPendiente,
		TokenHash:       tokenHash,
		ExpiraEn:        momento.Add(duracionVinculacion),
		CreadaEn:        momento,
		ApoderadoNombre: apoderado.NombreUso,
		ApoderadoEmail:  apoderado.Email,
		EscolarNombre:   escolar.NombreUso,
		EscolarEmail:    escolar.Email,
	}
}

func (servicio *VinculacionService) notificarSolicitud(apoderado, escolar *Usuario, token string) {
	servicio.Bus.Notificar(Event{Nombre: VinculacionSolicitada, Payload: map[string]any{
		"usuario_id":       apoderado.ID,
		"rol":              string(apoderado.Rol),
		"apoderado_id":     apoderado.ID,
		"escolar_id":       escolar.ID,
		"apoderado_nombre": apoderado.NombreUso,
		"apoderado_email":  apoderado.Email,
		"escolar_nombre":   escolar.NombreUso,
		"email":            escolar.Email,
		"token":            token,
	}})
}
