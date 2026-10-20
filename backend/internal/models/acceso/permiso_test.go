package acceso

import "testing"

func TestPermisosPorRol(t *testing.T) {
	if !RolCoordinador.TienePermiso(PermisoCoordinacionAcceder) {
		t.Fatal("el coordinador debe tener coordinacion:acceder")
	}
	for _, rol := range []Rol{RolEscolar, RolApoderado, RolTutor} {
		if rol.TienePermiso(PermisoCoordinacionAcceder) {
			t.Fatalf("%s no debe tener coordinacion:acceder", rol)
		}
	}
	if len(RolEscolar.Permisos()) == 0 {
		t.Fatal("el escolar debe conservar sus permisos")
	}
	if len(Rol("desconocido").Permisos()) != 0 {
		t.Fatal("un rol desconocido no debe tener permisos")
	}
}

func TestPermisosDevuelveCopia(t *testing.T) {
	permisos := RolCoordinador.Permisos()
	permisos[0] = "alterado"
	if !RolCoordinador.TienePermiso(PermisoCoordinacionAcceder) {
		t.Fatal("modificar la copia no debe alterar la política")
	}
}
