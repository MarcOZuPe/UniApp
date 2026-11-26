package acceso

import "context"

const (
	maximoLogsAdministracion  = 200
	logsAdministracionDefecto = 50
)

type AdministracionService struct {
	*Nucleo
}

type DatosNuevoGeneral struct {
	Email      string
	NombreUso  string
	Clave      string
	FacultadID string
}

func (servicio *AdministracionService) ListarGenerales(
	ctx context.Context,
	actor *Usuario,
) ([]*Usuario, error) {
	if !esAdministrador(actor) {
		return nil, ErrNoAutorizado
	}
	return servicio.RepositorioUsuarios.UsuariosActivosPorRol(ctx, RolCoordinadorGeneral, "")
}

func (servicio *AdministracionService) CrearGeneral(
	ctx context.Context,
	actor *Usuario,
	datos DatosNuevoGeneral,
	ip string,
) (*Usuario, error) {
	if !esAdministrador(actor) {
		return nil, ErrNoAutorizado
	}
	facultad, err := servicio.RepositorioFacultades.FacultadPorID(ctx, datos.FacultadID)
	if err != nil || facultad == nil {
		return nil, ErrValidacion{Campos: map[string]string{"facultad_id": "La facultad no existe."}}
	}
	usuario, err := servicio.CreadorUsuarios.CrearInterno(SolicitudRegistro{
		Email: datos.Email, NombreUso: datos.NombreUso, Clave: datos.Clave,
		Rol: RolCoordinadorGeneral, FacultadID: facultad.ID,
	})
	if err != nil {
		return nil, err
	}
	creado, err := servicio.RepositorioUsuarios.CrearUsuario(ctx, usuario)
	if err != nil {
		return nil, err
	}
	servicio.registrarLog(ctx, actor, ip, "coordinador_general.creado", creado.ID, facultad.ID)
	return creado, nil
}

func (servicio *AdministracionService) EliminarGeneral(
	ctx context.Context,
	actor *Usuario,
	generalID, ip string,
) error {
	if !esAdministrador(actor) {
		return ErrNoAutorizado
	}
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, generalID)
	if err != nil || usuario.Rol != RolCoordinadorGeneral || !usuario.Activo {
		return ErrCoordinadorNoEncontrado
	}
	if err := servicio.RepositorioUsuarios.DesactivarUsuario(ctx, usuario.ID); err != nil {
		return err
	}
	if err := servicio.RepositorioSesiones.EliminarSesionesDeUsuario(ctx, usuario.ID); err != nil {
		return err
	}
	servicio.registrarLog(ctx, actor, ip, "coordinador_general.eliminado", usuario.ID, usuario.FacultadID)
	return nil
}

func (servicio *AdministracionService) Logs(
	ctx context.Context,
	actor *Usuario,
	limite int,
) ([]LogSistema, error) {
	if !esAdministrador(actor) {
		return nil, ErrNoAutorizado
	}
	if limite <= 0 {
		limite = logsAdministracionDefecto
	}
	if limite > maximoLogsAdministracion {
		limite = maximoLogsAdministracion
	}
	return servicio.RepositorioLogs.LogsSistema(ctx, FiltroLogsSistema{Limite: limite})
}

func (servicio *AdministracionService) registrarLog(
	ctx context.Context,
	actor *Usuario,
	ip, evento, coordinadorID, facultadID string,
) {
	_ = servicio.RepositorioLogs.RegistrarLogSistema(ctx, LogSistema{
		Categoria: CategoriaAdministracion,
		Evento:    evento,
		ActorID:   actor.ID,
		ActorRol:  string(actor.Rol),
		IP:        ip,
		Detalle: map[string]any{
			"coordinador_general_id": coordinadorID,
			"facultad_id":            facultadID,
		},
	})
}

func esAdministrador(usuario *Usuario) bool {
	return usuario != nil && usuario.Rol == RolAdministrador && usuario.Activo
}
