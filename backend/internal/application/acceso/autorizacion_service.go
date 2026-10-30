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

func (servicio *AutorizacionService) CambiarAutorizacion(ctx context.Context, solicitud SolicitudCambioAutorizacion) (*Autorizacion, error) {
	vinculacionActiva, err := servicio.obtenerVinculacionVerificada(ctx, solicitud.SolicitudAutorizacion)
	if err != nil {
		return nil, err
	}

	estadoDestino := servicio.determinarEstadoDestino(solicitud.Autorizado)
	cambio := servicio.compilarCambioAutorizacion(solicitud.Apoderado, vinculacionActiva, estadoDestino)

	return servicio.RepositorioAutorizaciones.CambiarAutorizacion(ctx, cambio)
}

func (servicio *AutorizacionService) HistorialAutorizaciones(ctx context.Context, solicitud SolicitudAutorizacion) ([]Autorizacion, error) {
	vinculacionActiva, err := servicio.obtenerVinculacionVerificada(ctx, solicitud)
	if err != nil {
		return nil, err
	}
	return servicio.RepositorioAutorizaciones.AutorizacionesDeVinculacion(ctx, vinculacionActiva.ID)
}

func (servicio *AutorizacionService) AutorizacionVigente(ctx context.Context, escolarID string) (bool, error) {
	identificador := strings.TrimSpace(escolarID)
	if identificador == "" {
		return false, nil
	}

	autorizacionActual, err := servicio.RepositorioAutorizaciones.AutorizacionVigenteDeEscolar(ctx, identificador)
	if errors.Is(err, ErrNoEncontrado) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return autorizacionActual.Vigente(), nil
}

func (servicio *AutorizacionService) obtenerVinculacionVerificada(ctx context.Context, solicitud SolicitudAutorizacion) (*Vinculacion, error) {
	if err := servicio.validarApoderado(solicitud.Apoderado); err != nil {
		return nil, err
	}

	identificadorEscolar, err := servicio.extraerIdentificadorEscolar(solicitud.EscolarID)
	if err != nil {
		return nil, err
	}

	vinculacion, err := servicio.RepositorioVinculaciones.VinculacionAceptadaDeEscolar(ctx, identificadorEscolar)
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

func (servicio *AutorizacionService) validarApoderado(apoderado *Usuario) error {
	if apoderado == nil || apoderado.Rol != RolApoderado || !apoderado.Activo {
		return ErrNoEsApoderado
	}
	return nil
}

func (servicio *AutorizacionService) extraerIdentificadorEscolar(escolarID string) (string, error) {
	identificador := strings.TrimSpace(escolarID)
	if identificador == "" {
		return "", ErrValidacion{Campos: map[string]string{"escolar_id": "el escolar es obligatorio"}}
	}
	return identificador, nil
}

func (servicio *AutorizacionService) determinarEstadoDestino(autorizado bool) EstadoAutorizacion {
	if autorizado {
		return AutorizacionOtorgada
	}
	return AutorizacionRevocada
}

func (servicio *AutorizacionService) compilarCambioAutorizacion(apoderado *Usuario, vinculacion *Vinculacion, estado EstadoAutorizacion) CambioAutorizacion {
	return CambioAutorizacion{
		VinculacionID: vinculacion.ID,
		EscolarID:     vinculacion.EscolarID,
		ApoderadoID:   apoderado.ID,
		Estado:        estado,
		Momento:       servicio.Ahora(),
	}
}
