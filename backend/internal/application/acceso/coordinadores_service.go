package acceso

import "context"

type CoordinadoresService struct {
	*Nucleo
}

func (servicio *CoordinadoresService) Listar(ctx context.Context, actor *Usuario) ([]*Usuario, error) {
	if !esCoordinadorGeneral(actor) || actor.FacultadID == "" {
		return nil, ErrNoAutorizado
	}
	return servicio.RepositorioUsuarios.UsuariosActivosPorRol(ctx, RolCoordinador, actor.FacultadID)
}

func (servicio *CoordinadoresService) Crear(
	ctx context.Context,
	actor *Usuario,
	datos SolicitudRegistro,
) (*Usuario, error) {
	if !esCoordinadorGeneral(actor) || actor.FacultadID == "" {
		return nil, ErrNoAutorizado
	}
	datos.Rol = RolCoordinador
	datos.FacultadID = actor.FacultadID
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
	if !esCoordinadorGeneral(actor) || actor.FacultadID == "" {
		return ErrNoAutorizado
	}
	coordinador, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, coordinadorID)
	if err != nil || coordinador.Rol != RolCoordinador || !coordinador.Activo ||
		coordinador.FacultadID != actor.FacultadID {
		return ErrCoordinadorNoEncontrado
	}
	if err := servicio.RepositorioUsuarios.DesactivarUsuario(ctx, coordinador.ID); err != nil {
		return err
	}
	// El middleware comprueba Activo en cada petición; no hay sesión que borrar.
	return nil
}

func esCoordinadorGeneral(usuario *Usuario) bool {
	return usuario != nil && usuario.Rol == RolCoordinadorGeneral && usuario.Activo
}
