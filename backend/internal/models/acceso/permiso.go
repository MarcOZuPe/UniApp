package acceso

type Permiso string

const (
	PermisoCoordinacionAcceder Permiso = "coordinacion:acceder"
	PermisoExpedientePresentar Permiso = "expediente:presentar"
	PermisoReservaCrear        Permiso = "reserva:crear"
	PermisoAutorizacionOtorgar Permiso = "autorizacion:otorgar"
)

var permisosPorRol = map[Rol][]Permiso{
	RolEscolar:     {PermisoReservaCrear},
	RolApoderado:   {PermisoAutorizacionOtorgar},
	RolTutor:       {PermisoExpedientePresentar},
	RolCoordinador: {PermisoCoordinacionAcceder},
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
