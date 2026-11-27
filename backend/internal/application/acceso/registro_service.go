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
	if err := servicio.validarFacultad(ctx, usuario); err != nil {
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
	if err := servicio.validarFacultad(ctx, usuario); err != nil {
		return nil, err
	}
	return servicio.RepositorioUsuarios.CrearUsuario(ctx, usuario)
}

func (servicio *RegistroService) Facultades(ctx context.Context) ([]Facultad, error) {
	return servicio.RepositorioFacultades.Facultades(ctx)
}

func (servicio *RegistroService) validarFacultad(ctx context.Context, usuario *Usuario) error {
	if !usuario.Rol.RequiereFacultad() {
		usuario.FacultadID = ""
		return nil
	}
	if usuario.FacultadID == "" {
		return ErrValidacion{Campos: map[string]string{"facultad_id": "Elige una facultad."}}
	}
	facultad, err := servicio.RepositorioFacultades.FacultadPorID(ctx, usuario.FacultadID)
	if err != nil || facultad == nil {
		return ErrValidacion{Campos: map[string]string{"facultad_id": "La facultad no existe."}}
	}
	usuario.FacultadID = facultad.ID
	usuario.FacultadNombre = facultad.Nombre
	return nil
}
