package acceso

import (
	"testing"
)

func TestRolesYPermisos(t *testing.T) {
	casos := []struct {
		rol           Rol
		coordinacion  bool
		requiereFacul bool
		permiso       Permiso
	}{
		{RolEscolar, false, false, PermisoReservaCrear},
		{RolApoderado, false, false, PermisoAutorizacionOtorgar},
		{RolTutor, false, true, PermisoExpedientePresentar},
		{RolCoordinador, true, true, PermisoIncidenteRevisar},
		{RolCoordinadorGeneral, true, true, PermisoCoordinadoresGestionar},
		{RolAdministrador, false, false, PermisoAdministracionAcceder},
	}
	for _, caso := range casos {
		t.Run(string(caso.rol), func(t *testing.T) {
			if obtenido := caso.rol.EsCoordinacion(); obtenido != caso.coordinacion {
				t.Fatalf("EsCoordinacion()=%v, se esperaba %v", obtenido, caso.coordinacion)
			}
			if obtenido := caso.rol.RequiereFacultad(); obtenido != caso.requiereFacul {
				t.Fatalf("RequiereFacultad()=%v, se esperaba %v", obtenido, caso.requiereFacul)
			}
			if !caso.rol.TienePermiso(caso.permiso) {
				t.Fatalf("el rol %q debería tener el permiso %q", caso.rol, caso.permiso)
			}
			fabrica, err := FabricaParaRol(caso.rol)
			if err != nil || fabrica == nil {
				t.Fatalf("no se pudo obtener la fábrica: %v", err)
			}
			if fabrica.CrearPerfil("uni.edu.pe") == nil {
				t.Fatalf("la fábrica devolvió un perfil nulo")
			}
		})
	}
	if Rol("desconocido").EsCoordinacion() || Rol("desconocido").RequiereFacultad() {
		t.Fatal("un rol desconocido no debería tener capacidades")
	}
	if _, err := FabricaParaRol(Rol("desconocido")); err == nil {
		t.Fatal("un rol desconocido debería rechazarse")
	}
}

func TestPermisosDevuelveUnaCopia(t *testing.T) {
	originales := RolEscolar.Permisos()
	if len(originales) == 0 {
		t.Fatal("el rol escolar debería tener permisos")
	}
	originales[0] = "permiso:alterado"
	if RolEscolar.TienePermiso("permiso:alterado") {
		t.Fatal("mutar la copia no debe alterar los permisos del rol")
	}
}

func TestPerfilesYValidacionesDeCorreo(t *testing.T) {
	if perfil, err := ObtenerPerfil(RolEscolar, "uni.edu.pe"); err != nil || perfil.Rol() != RolEscolar || !perfil.EsPublico() {
		t.Fatalf("perfil escolar inesperado: %+v err=%v", perfil, err)
	}
	tutor, err := ObtenerPerfil(RolTutor, "uni.edu.pe,uni.pe")
	if err != nil || tutor.Rol() != RolTutor || !tutor.EsPublico() {
		t.Fatalf("perfil tutor inesperado: %+v err=%v", tutor, err)
	}
	if campos := tutor.ValidarExtra(SolicitudRegistro{Email: "tutor@uni.edu.pe"}); len(campos) != 0 {
		t.Fatalf("el correo institucional debería aceptarse: %v", campos)
	}
	if campos := tutor.ValidarExtra(SolicitudRegistro{Email: "tutor@gmail.com"}); len(campos) == 0 {
		t.Fatal("un correo no institucional debería rechazarse")
	}
	if _, err := ObtenerPerfil(Rol("desconocido"), ""); err == nil {
		t.Fatal("un rol desconocido no debería tener perfil")
	}
}

func TestPoliticasDeClave(t *testing.T) {
	basica := PoliticaBasica{Min: 10}
	if observaciones := basica.Validar("Clave-1234", "otro@example.test"); len(observaciones) != 0 {
		t.Fatalf("la clave válida no debería tener observaciones: %v", observaciones)
	}
	if observaciones := basica.Validar("corta", "otro@example.test"); len(observaciones) == 0 {
		t.Fatal("la clave corta debería rechazarse")
	}
	if observaciones := basica.Validar("usuario", "usuario@example.test"); len(observaciones) == 0 {
		t.Fatal("la clave igual al correo debería rechazarse")
	}
	estricta := PoliticaEstricta{Min: 10}
	if observaciones := estricta.Validar("Clave-1234", "otro@example.test"); len(observaciones) != 0 {
		t.Fatalf("la clave estricta válida no debería tener observaciones: %v", observaciones)
	}
	if observaciones := estricta.Validar("clave1234", "otro@example.test"); len(observaciones) == 0 {
		t.Fatal("la clave estricta sin mayúscula y símbolo debería rechazarse")
	}
	validador := NuevoValidadorClave(basica)
	if observaciones := validador.Validar("Clave-1234", "otro@example.test"); len(observaciones) != 0 {
		t.Fatalf("el validador debería delegar en la política: %v", observaciones)
	}
	validador.FijarPolitica(estricta)
	if observaciones := validador.Validar("clave1234", "otro@example.test"); len(observaciones) == 0 {
		t.Fatal("el validador debería usar la nueva política")
	}
}

func TestEstadosDeAutorizacionYVinculacion(t *testing.T) {
	if EstadoParaAutorizacion(true) != AutorizacionOtorgada {
		t.Fatal("autorizar debería otorgar el permiso")
	}
	if EstadoParaAutorizacion(false) != AutorizacionRevocada {
		t.Fatal("retirar debería revocar el permiso")
	}
	if !(Autorizacion{Estado: AutorizacionOtorgada}).Vigente() {
		t.Fatal("la autorización otorgada debería estar vigente")
	}
	if (Autorizacion{Estado: AutorizacionRevocada}).Vigente() {
		t.Fatal("la autorización revocada no debería estar vigente")
	}
	if errorValidacion := (ErrValidacion{Campos: map[string]string{"x": "y"}}).Error(); errorValidacion != "datos inválidos" {
		t.Fatalf("mensaje inesperado: %q", errorValidacion)
	}
}

func TestFacultadesYLogSistema(t *testing.T) {
	facultad := Facultad{ID: "fac-1", Nombre: "Ingeniería"}
	if facultad.ID != "fac-1" || facultad.Nombre != "Ingeniería" {
		t.Fatal("la facultad no conserva sus datos")
	}
	log := LogSistema{Categoria: CategoriaAdministracion, Evento: "evento", Detalle: map[string]any{"x": 1}}
	if log.Categoria != CategoriaAdministracion || log.Detalle["x"] != 1 {
		t.Fatal("el log no conserva sus datos")
	}
}
