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

type OrigenSolicitud struct {
	Actor *Usuario
	IP    string
}

type EventoAdministracion struct {
	Origen        OrigenSolicitud
	Nombre        string
	CoordinadorID string
	FacultadID    string
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
	origen OrigenSolicitud,
	datos DatosNuevoGeneral,
) (*Usuario, error) {
	if !esAdministrador(origen.Actor) {
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
	servicio.registrarLog(ctx, EventoAdministracion{
		Origen: origen, Nombre: "coordinador_general.creado",
		CoordinadorID: creado.ID, FacultadID: facultad.ID,
	})
	return creado, nil
}

func (servicio *AdministracionService) EliminarGeneral(
	ctx context.Context,
	origen OrigenSolicitud,
	generalID string,
) error {
	if !esAdministrador(origen.Actor) {
		return ErrNoAutorizado
	}
	usuario, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, generalID)
	if err != nil || usuario.Rol != RolCoordinadorGeneral || !usuario.Activo {
		return ErrCoordinadorNoEncontrado
	}
	if err := servicio.RepositorioUsuarios.DesactivarUsuario(ctx, usuario.ID); err != nil {
		return err
	}
	// El middleware comprueba Activo en cada petición; no hay sesión que borrar.
	servicio.registrarLog(ctx, EventoAdministracion{
		Origen: origen, Nombre: "coordinador_general.eliminado",
		CoordinadorID: usuario.ID, FacultadID: usuario.FacultadID,
	})
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

func (servicio *AdministracionService) registrarLog(ctx context.Context, evento EventoAdministracion) {
	actor := evento.Origen.Actor
	_ = servicio.RepositorioLogs.RegistrarLogSistema(ctx, LogSistema{
		Categoria: CategoriaAdministracion,
		Evento:    evento.Nombre,
		ActorID:   actor.ID,
		ActorRol:  string(actor.Rol),
		IP:        evento.Origen.IP,
		Detalle: map[string]any{
			"coordinador_general_id": evento.CoordinadorID,
			"facultad_id":            evento.FacultadID,
		},
	})
}

func esAdministrador(usuario *Usuario) bool {
	return usuario != nil && usuario.Rol == RolAdministrador && usuario.Activo
}
