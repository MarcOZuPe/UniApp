package acceso

import (
	"context"
	"strings"
)

const (
	maxCambiosClavePorHora   = 3
	maxCambiosClavePorMinuto = 1
)

type PerfilService struct {
	*Nucleo
}

type DatosCambioClave struct {
	ClaveActual string
	ClaveNueva  string
}

func (servicio *PerfilService) CambiarClave(
	ctx context.Context,
	usuarioID string,
	datos DatosCambioClave,
) error {
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, usuarioID)
	if err != nil || !usuario.Activo {
		return ErrNoEncontrado
	}
	if !servicio.Hasher.Verify(datos.ClaveActual, usuario.ClaveHash) {
		return ErrClaveActualIncorrecta
	}
	if datos.ClaveActual == datos.ClaveNueva {
		return ErrValidacion{Campos: map[string]string{
			"clave_nueva": "La nueva contraseña debe ser diferente.",
		}}
	}
	if observaciones := servicio.ValidadorClaves.ValidadorPara(usuario.Rol).Validar(
		datos.ClaveNueva, usuario.Email,
	); len(observaciones) > 0 {
		return ErrValidacion{Campos: map[string]string{
			"clave_nueva": strings.Join(observaciones, "; "),
		}}
	}
	nuevoHash, err := servicio.Hasher.Hash(datos.ClaveNueva)
	if err != nil {
		return err
	}
	if err := servicio.RepositorioPerfil.CambiarClaveConLimite(
		ctx, usuario.ID, nuevoHash, servicio.Ahora(),
		maxCambiosClavePorHora, maxCambiosClavePorMinuto,
	); err != nil {
		return err
	}
	servicio.Bus.Notificar(Event{Nombre: ClaveCambiada, Payload: map[string]any{
		"usuario_id": usuario.ID,
		"rol":        string(usuario.Rol),
		"email":      usuario.Email,
	}})
	return nil
}
