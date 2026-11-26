package acceso

import "testing"

func TestEstadoParaAutorizacion(t *testing.T) {
	if estado := EstadoParaAutorizacion(true); estado != AutorizacionOtorgada {
		t.Fatalf("autorizar debería otorgar el permiso, se obtuvo %q", estado)
	}
	if estado := EstadoParaAutorizacion(false); estado != AutorizacionRevocada {
		t.Fatalf("retirar debería revocar el permiso, se obtuvo %q", estado)
	}
	if !(Autorizacion{Estado: AutorizacionOtorgada}).Vigente() {
		t.Fatal("una autorización otorgada debería estar vigente")
	}
	if (Autorizacion{Estado: AutorizacionRevocada}).Vigente() {
		t.Fatal("una autorización revocada no debería estar vigente")
	}
}

func TestFabricaParaRolDevuelveLaFabricaDelRol(t *testing.T) {
	for _, rol := range []Rol{
		RolEscolar,
		RolApoderado,
		RolTutor,
		RolCoordinador,
		RolCoordinadorGeneral,
		RolAdministrador,
	} {
		fabrica, err := FabricaParaRol(rol)
		if err != nil || fabrica == nil {
			t.Fatalf("el rol %q debería tener fábrica: %v", rol, err)
		}
	}
	if _, err := FabricaParaRol(Rol("desconocido")); err == nil {
		t.Fatal("un rol desconocido debería rechazarse")
	}
}

func TestFabricaAdministradorSoloTieneAccesoALaAdministracion(t *testing.T) {
	fabrica, err := FabricaParaRol(RolAdministrador)
	if err != nil {
		t.Fatalf("no se pudo obtener la fábrica de administrador: %v", err)
	}
	permisos := fabrica.Permisos()
	if len(permisos) != 1 || permisos[0] != PermisoAdministracionAcceder {
		t.Fatalf("permisos inesperados para administrador: %v", permisos)
	}
}
