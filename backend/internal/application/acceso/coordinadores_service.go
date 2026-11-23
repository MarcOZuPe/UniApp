package acceso

import "context"

type CoordinadoresService struct {
	*Nucleo
}

func (servicio *CoordinadoresService) Listar(ctx context.Context, actor *Usuario) ([]*Usuario, error) {
	if !esCoordinadorGeneral(actor) {
		return nil, ErrNoAutorizado
	}
	return servicio.RepositorioUsuarios.UsuariosActivosPorRol(ctx, RolCoordinador)
}

func (servicio *CoordinadoresService) Crear(
	ctx context.Context,
	actor *Usuario,
	datos SolicitudRegistro,
) (*Usuario, error) {
	if !esCoordinadorGeneral(actor) {
		return nil, ErrNoAutorizado
	}
	datos.Rol = RolCoordinador
	usuario, err := servicio.CreadorUsuarios.CrearInterno(datos)
	if err != nil {
		return nil, err
	}
	return servicio.RepositorioUsuarios.CrearUsuario(ctx, usuario)
}

func (servicio *CoordinadoresService) Eliminar(
	ctx context.Context,
	actor *Usuario,
	coordinadorID string,
) error {
	if !esCoordinadorGeneral(actor) {
		return ErrNoAutorizado
	}
	coordinador, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, coordinadorID)
	if err != nil || coordinador.Rol != RolCoordinador || !coordinador.Activo {
		return ErrCoordinadorNoEncontrado
	}
	if err := servicio.RepositorioUsuarios.DesactivarUsuario(ctx, coordinador.ID); err != nil {
		return err
	}
	return servicio.RepositorioSesiones.EliminarSesionesDeUsuario(ctx, coordinador.ID)
}

func esCoordinadorGeneral(usuario *Usuario) bool {
	return usuario != nil && usuario.Rol == RolCoordinadorGeneral && usuario.Activo
}
