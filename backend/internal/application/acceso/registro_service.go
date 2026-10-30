package acceso

import (
	"context"
)

type RegistroService struct {
	*Nucleo
}

func (servicio *RegistroService) Registrar(ctx context.Context, datos SolicitudRegistro) (*Usuario, error) {
	usuario, err := servicio.CreadorUsuarios.CrearPublico(datos)
	if err != nil {
		return nil, err
	}

	registrado, err := servicio.RepositorioUsuarios.CrearUsuario(ctx, usuario)
	if err != nil {
		return nil, err
	}

	servicio.Bus.Notificar(eventoDeUsuario(UsuarioRegistrado, registrado))

	return registrado, nil
}

func (servicio *RegistroService) CrearAdministrado(ctx context.Context, datos SolicitudRegistro) (*Usuario, error) {
	usuario, err := servicio.CreadorUsuarios.CrearInterno(datos)
	if err != nil {
		return nil, err
	}
	return servicio.RepositorioUsuarios.CrearUsuario(ctx, usuario)
}
