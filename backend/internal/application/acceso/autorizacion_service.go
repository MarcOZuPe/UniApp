package acceso

import (
	"context"
	"errors"
	"strings"
)

type AutorizacionService struct{ *Nucleo }

type SolicitudAutorizacion struct {
	Apoderado *Usuario
	EscolarID string
}

type SolicitudCambioAutorizacion struct {
	SolicitudAutorizacion
	Autorizado bool
}

func (s *AutorizacionService) CambiarAutorizacion(ctx context.Context, solicitud SolicitudCambioAutorizacion) (*Autorizacion, error) {
	vinculacion, err := s.vinculacionDelApoderado(ctx, solicitud.SolicitudAutorizacion)
	if err != nil {
		return nil, err
	}

	estado := AutorizacionRevocada
	if solicitud.Autorizado {
		estado = AutorizacionOtorgada
	}
	return s.repositorio.CambiarAutorizacion(ctx, CambioAutorizacion{
		VinculacionID: vinculacion.ID,
		EscolarID:     vinculacion.EscolarID,
		ApoderadoID:   solicitud.Apoderado.ID,
		Estado:        estado,
		Momento:       s.ahora(),
	})
}

func (s *AutorizacionService) HistorialAutorizaciones(ctx context.Context, solicitud SolicitudAutorizacion) ([]Autorizacion, error) {
	vinculacion, err := s.vinculacionDelApoderado(ctx, solicitud)
	if err != nil {
		return nil, err
	}
	return s.repositorio.AutorizacionesDeVinculacion(ctx, vinculacion.ID)
}

func (s *AutorizacionService) vinculacionDelApoderado(ctx context.Context, solicitud SolicitudAutorizacion) (*Vinculacion, error) {
	if solicitud.Apoderado == nil || solicitud.Apoderado.Rol != RolApoderado || !solicitud.Apoderado.Activo {
		return nil, ErrNoEsApoderado
	}
	escolarID := strings.TrimSpace(solicitud.EscolarID)
	if escolarID == "" {
		return nil, ErrValidacion{Campos: map[string]string{"escolar_id": "el escolar es obligatorio"}}
	}

	vinculacion, err := s.repositorio.VinculacionAceptadaDeEscolar(ctx, escolarID)
	if errors.Is(err, ErrNoEncontrado) {
		return nil, ErrSinApoderado
	}
	if err != nil {
		return nil, err
	}
	if vinculacion.ApoderadoID != solicitud.Apoderado.ID {
		return nil, ErrVinculacionAjena
	}
	return vinculacion, nil
}

func (s *AutorizacionService) AutorizacionVigente(ctx context.Context, escolarID string) (bool, error) {
	autorizacion, err := s.repositorio.AutorizacionVigenteDeEscolar(ctx, strings.TrimSpace(escolarID))
	if errors.Is(err, ErrNoEncontrado) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return autorizacion.Vigente(), nil
}
