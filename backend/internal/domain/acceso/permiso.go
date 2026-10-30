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

func (rol Rol) Permisos() []Permiso {
	permisosOriginales := permisosDeRol(rol)
	copia := make([]Permiso, len(permisosOriginales))
	copy(copia, permisosOriginales)
	return copia
}

func (rol Rol) TienePermiso(permisoBuscado Permiso) bool {
	for _, permiso := range permisosDeRol(rol) {
		if permiso == permisoBuscado {
			return true
		}
	}
	return false
}

func permisosDeRol(rol Rol) []Permiso {
	fabrica, err := FabricaParaRol(rol)
	if err != nil {
		return nil
	}
	return fabrica.Permisos()
}
