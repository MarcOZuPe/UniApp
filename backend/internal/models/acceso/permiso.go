package acceso

type Permiso string

const (
	PermisoCoordinacionAcceder      Permiso = "coordinacion:acceder"
	PermisoExpedientePresentar      Permiso = "expediente:presentar"
	PermisoReservaCrear             Permiso = "reserva:crear"
	PermisoAutorizacionOtorgar      Permiso = "autorizacion:otorgar"
	PermisoCatalogoGestionar        Permiso = "catalogo:gestionar"
	PermisoCatalogoSolicitarColegio Permiso = "catalogo:solicitar_colegio"
)

var permisosPorRol = map[Rol][]Permiso{
	RolEscolar:     {PermisoReservaCrear, PermisoCatalogoSolicitarColegio},
	RolApoderado:   {PermisoAutorizacionOtorgar, PermisoCatalogoSolicitarColegio},
	RolTutor:       {PermisoExpedientePresentar},
	RolCoordinador: {PermisoCoordinacionAcceder, PermisoCatalogoGestionar},
}

func (r Rol) Permisos() []Permiso {
	propios := permisosPorRol[r]
	copia := make([]Permiso, len(propios))
	copy(copia, propios)
	return copia
}

func (r Rol) TienePermiso(permiso Permiso) bool {
	for _, propio := range permisosPorRol[r] {
		if propio == permiso {
			return true
		}
	}
	return false
}
