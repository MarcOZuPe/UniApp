package acceso

import (
	"context"
	"errors"
	"strings"
)

// CambiarAutorizacion permite al apoderado revocar u otorgar el permiso para
// crear nuevas reservas. No modifica reservas ya confirmadas: la comprobación
// ocurre únicamente al intentar reservar.
func (s *Servicio) CambiarAutorizacion(ctx context.Context, apoderado *Usuario, escolarID string, autorizado bool) (*Autorizacion, error) {
	if apoderado == nil || apoderado.Rol != RolApoderado || !apoderado.Activo {
		return nil, ErrNoEsApoderado
	}
	escolarID = strings.TrimSpace(escolarID)
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
	if vinculacion.ApoderadoID != apoderado.ID {
		return nil, ErrVinculacionAjena
	}

	estado := AutorizacionRevocada
	if autorizado {
		estado = AutorizacionOtorgada
	}
	return s.repositorio.CambiarAutorizacion(ctx, vinculacion.ID, vinculacion.EscolarID, apoderado.ID, estado, s.ahora())
}

// AutorizacionVigente responde si el escolar puede crear nuevas reservas. La
// ausencia de un vínculo o de una autorización equivale a no autorizado.
func (s *Servicio) AutorizacionVigente(ctx context.Context, escolarID string) (bool, error) {
	autorizacion, err := s.repositorio.AutorizacionVigenteDeEscolar(ctx, strings.TrimSpace(escolarID))
	if errors.Is(err, ErrNoEncontrado) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return autorizacion.Vigente(), nil
}
