package acceso

import "testing"

func TestFabricaParaRolCompleta(testingT *testing.T) {
	roles := []Rol{RolEscolar, RolApoderado, RolTutor, RolCoordinador}
	for _, rol := range roles {
		fabrica, err := FabricaParaRol(rol)
		if err != nil {
			testingT.Fatalf("rol %s: %v", rol, err)
		}
		perfil := fabrica.CrearPerfil("uni.edu.pe,uni.pe")
		if perfil == nil {
			testingT.Fatalf("el rol %s no tiene perfil", rol)
		}
		if perfil.Rol() != rol {
			testingT.Fatalf("el perfil de %s devolvió el rol %s", rol, perfil.Rol())
		}
		if len(fabrica.Permisos()) == 0 {
			testingT.Fatalf("el rol %s no tiene permisos", rol)
		}
		if len(rol.Permisos()) == 0 {
			testingT.Fatalf("el rol %s no expone permisos", rol)
		}
	}

	if _, err := FabricaParaRol(Rol("desconocido")); err == nil {
		testingT.Fatal("un rol desconocido debe fallar")
	}
}
