package acceso

import (
	"context"
	"strings"
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
	if err := servicio.RepositorioUsuarios.ActualizarClave(ctx, usuario.ID, nuevoHash); err != nil {
		return err
	}
	servicio.Bus.Notificar(eventoDeUsuario(ClaveRestablecida, usuario))
	return nil
}
