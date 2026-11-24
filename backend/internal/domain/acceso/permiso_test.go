package acceso

import "testing"

func TestPermisosPorRol(testingT *testing.T) {
	if !RolCoordinador.TienePermiso(PermisoCoordinacionAcceder) {
		testingT.Fatal("el coordinador debe tener coordinacion:acceder")
	}
	if !RolCoordinadorGeneral.TienePermiso(PermisoCoordinacionAcceder) ||
		!RolCoordinadorGeneral.TienePermiso(PermisoCoordinadoresGestionar) {
		testingT.Fatal("el coordinador general debe conservar coordinación y gestionar coordinadores")
	}
	if RolCoordinador.TienePermiso(PermisoCoordinadoresGestionar) {
		testingT.Fatal("el coordinador no debe gestionar cuentas de coordinación")
	}
	if !RolAdministrador.TienePermiso(PermisoAdministracionAcceder) {
		testingT.Fatal("el administrador debe acceder a administración")
	}
	if RolAdministrador.TienePermiso(PermisoCoordinacionAcceder) {
		testingT.Fatal("el administrador no debe mezclarse con coordinación")
	}
	for _, rol := range []Rol{RolEscolar, RolApoderado, RolTutor} {
		if rol.TienePermiso(PermisoCoordinacionAcceder) {
			testingT.Fatalf("%s no debe tener coordinacion:acceder", rol)
		}
	}
	if len(RolEscolar.Permisos()) == 0 {
		testingT.Fatal("el escolar debe conservar sus permisos")
	}
	if len(Rol("desconocido").Permisos()) != 0 {
		testingT.Fatal("un rol desconocido no debe tener permisos")
	}
}

func TestPermisosDevuelveCopia(testingT *testing.T) {
	permisosAlterados := RolCoordinador.Permisos()
	permisosAlterados[0] = "alterado"
	if !RolCoordinador.TienePermiso(PermisoCoordinacionAcceder) {
		testingT.Fatal("modificar la copia no debe alterar la política")
	}
}
