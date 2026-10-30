package acceso_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	postgresRepo "uniapp/internal/adapters/repositories/postgres/acceso"
	app "uniapp/internal/application/acceso"
	modelo "uniapp/internal/domain/acceso"
	infra "uniapp/internal/infrastructure"
	"uniapp/migrations"
)

func TestRepositorioPostgresFlujoCompleto(testingT *testing.T) {
	urlBaseDatos := os.Getenv("TEST_DATABASE_URL")
	if urlBaseDatos == "" {
		testingT.Skip("TEST_DATABASE_URL no definida")
	}
	contexto := context.Background()
	baseDatos, err := migrations.Abrir(urlBaseDatos)
	if err != nil {
		testingT.Fatal(err)
	}
	defer baseDatos.Close()
	if err := migrations.Migrar(contexto, baseDatos, migrations.FS); err != nil {
		testingT.Fatal(err)
	}
	if err := migrations.Migrar(contexto, baseDatos, migrations.FS); err != nil {
		testingT.Fatal(err)
	}
	_, _ = baseDatos.Exec(`TRUNCATE usuarios CASCADE`)

	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	repositorio := postgresRepo.RepositorioPostgres{DB: baseDatos}
	busEventos := infra.NuevoBus()
	var tokenConfirmacion string
	busEventos.Registrar(modelo.ConfirmacionSolicitada, infra.NuevoObservadorFunc("captura-confirmacion", func(evento modelo.Event) {
		tokenConfirmacion, _ = evento.Payload["token"].(string)
	}))
	servicio := app.NuevoServicio(app.Dependencias{
		Repositorio: repositorio,
		Fabrica:     app.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         busEventos,
	})
	registrado, err := servicio.Registro.Registrar(contexto, modelo.SolicitudRegistro{Email: "pg@example.com", NombreUso: "Pg", Clave: "clave-larga-ok", Rol: modelo.RolEscolar})
	if err != nil {
		testingT.Fatal(err)
	}
	if registrado.EmailConfirmado || tokenConfirmacion == "" {
		testingT.Fatal("el registro público debe nacer sin confirmar y emitir el enlace")
	}
	if _, err := servicio.Registro.Registrar(contexto, modelo.SolicitudRegistro{Email: "PG@example.com", NombreUso: "Pg", Clave: "clave-larga-ok", Rol: modelo.RolEscolar}); !errors.Is(err, modelo.ErrEmailEnUso) {
		testingT.Fatalf("unicidad: %v", err)
	}
	if _, _, err := servicio.Sesion.Ingresar(contexto, app.CredencialesIngreso{Email: "pg@example.com", Clave: "clave-larga-ok"}); !errors.Is(err, modelo.ErrCredenciales) {
		testingT.Fatalf("sin confirmar debe ser invisible: %v", err)
	}
	if err := servicio.Confirmacion.ConfirmarCorreo(contexto, tokenConfirmacion); err != nil {
		testingT.Fatal(err)
	}
	if err := servicio.Confirmacion.ConfirmarCorreo(contexto, tokenConfirmacion); !errors.Is(err, modelo.ErrTokenInvalido) {
		testingT.Fatal("el enlace de confirmación es de un solo uso")
	}
	confirmado, err := repositorio.UsuarioPorEmail(contexto, "pg@example.com")
	if err != nil || !confirmado.EmailConfirmado {
		testingT.Fatalf("correo_confirmado debe persistir: %v", err)
	}
	tokenSesion, usuario, err := servicio.Sesion.Ingresar(contexto, app.CredencialesIngreso{Email: "pg@example.com", Clave: "clave-larga-ok"})
	if err != nil || usuario.Rol != modelo.RolEscolar {
		testingT.Fatal(err)
	}
	if usuarioSesion, err := servicio.Sesion.UsuarioDeSesion(contexto, tokenSesion); err != nil || usuarioSesion.UltimoIngreso == nil {
		testingT.Fatalf("sesión: %v", err)
	}

	var tokenReset string
	busEventos.Registrar(modelo.RecuperacionSolicitada, infra.NuevoObservadorFunc("captura-token-reset", func(evento modelo.Event) {
		tokenReset, _ = evento.Payload["token"].(string)
	}))
	_ = servicio.Recuperacion.SolicitarRecuperacion(contexto, "pg@example.com")
	if err := servicio.Recuperacion.RestablecerClave(contexto, app.SolicitudRestablecimiento{Token: tokenReset, NuevaClave: "nueva-clave-larga"}); err != nil {
		testingT.Fatal(err)
	}
	if err := servicio.Recuperacion.RestablecerClave(contexto, app.SolicitudRestablecimiento{Token: tokenReset, NuevaClave: "nueva-clave-larga"}); !errors.Is(err, modelo.ErrTokenInvalido) {
		testingT.Fatal("un solo uso")
	}
	if _, err := servicio.Sesion.UsuarioDeSesion(contexto, tokenSesion); err == nil {
		testingT.Fatal("sesiones cerradas tras restablecer")
	}

	if _, err := servicio.Registro.Registrar(contexto, modelo.SolicitudRegistro{Email: "vencido@example.com", NombreUso: "Vencido", Clave: "clave-larga-ok", Rol: modelo.RolEscolar}); err != nil {
		testingT.Fatal(err)
	}
	liberados, err := repositorio.EliminarUsuariosSinConfirmar(contexto, time.Now().Add(31*time.Minute))
	if err != nil || liberados != 1 {
		testingT.Fatalf("debía liberarse el registro vencido: %d (%v)", liberados, err)
	}
	if _, err := repositorio.UsuarioPorEmail(contexto, "vencido@example.com"); !errors.Is(err, modelo.ErrNoEncontrado) {
		testingT.Fatalf("el correo debe quedar libre: %v", err)
	}
	if confirmado, err := repositorio.UsuarioPorEmail(contexto, "pg@example.com"); err != nil || !confirmado.EmailConfirmado {
		testingT.Fatalf("no debe borrar cuentas confirmadas: %v", err)
	}
}
func TestRepositorioPostgresVinculacion(testingT *testing.T) {
	urlBaseDatos := os.Getenv("TEST_DATABASE_URL")
	if urlBaseDatos == "" {
		testingT.Skip("TEST_DATABASE_URL no definida")
	}
	contexto := context.Background()
	baseDatos, err := migrations.Abrir(urlBaseDatos)
	if err != nil {
		testingT.Fatal(err)
	}
	defer baseDatos.Close()
	if err := migrations.Migrar(contexto, baseDatos, migrations.FS); err != nil {
		testingT.Fatal(err)
	}
	_, _ = baseDatos.Exec(`TRUNCATE usuarios CASCADE`)

	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	busEventos := infra.NuevoBus()
	var tokenVinculacion string
	busEventos.Registrar(modelo.VinculacionSolicitada, infra.NuevoObservadorFunc("captura-vinculacion-pg", func(evento modelo.Event) {
		tokenVinculacion, _ = evento.Payload["token"].(string)
	}))
	servicio := app.NuevoServicio(app.Dependencias{
		Repositorio: postgresRepo.RepositorioPostgres{DB: baseDatos},
		Fabrica:     app.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         busEventos,
	})
	apoderado, err := servicio.Registro.CrearAdministrado(contexto, modelo.SolicitudRegistro{
		Email: "apoderado.pg@example.com", NombreUso: "Apoderado PG", Clave: "clave-larga-ok", Rol: modelo.RolApoderado,
	})
	if err != nil {
		testingT.Fatal(err)
	}
	escolar, err := servicio.Registro.CrearAdministrado(contexto, modelo.SolicitudRegistro{
		Email: "escolar.pg@example.com", NombreUso: "Escolar PG", Clave: "clave-larga-ok", Rol: modelo.RolEscolar,
	})
	if err != nil {
		testingT.Fatal(err)
	}

	creada, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, escolar.Email)
	if err != nil {
		testingT.Fatal(err)
	}
	if creada.Estado != modelo.EstadoVinculacionPendiente || tokenVinculacion == "" {
		testingT.Fatalf("solicitud: %+v", creada)
	}
	consultada, err := servicio.Vinculacion.ConsultarVinculacion(contexto, escolar, tokenVinculacion)
	if err != nil || consultada.ApoderadoEmail != apoderado.Email {
		testingT.Fatalf("consulta: %+v (%v)", consultada, err)
	}
	aceptada, err := servicio.Vinculacion.AceptarVinculacion(contexto, escolar, tokenVinculacion)
	if err != nil || aceptada.Estado != modelo.EstadoVinculacionAceptada || aceptada.AceptadaEn == nil {
		testingT.Fatalf("aceptar: %+v (%v)", aceptada, err)
	}
	if aceptada.Autorizacion == nil || !aceptada.Autorizacion.Vigente() {
		testingT.Fatalf("autorización inicial: %+v", aceptada.Autorizacion)
	}
	if autorizado, err := servicio.Autorizacion.AutorizacionVigente(contexto, escolar.ID); err != nil || !autorizado {
		testingT.Fatalf("autorización vigente: %v (%v)", autorizado, err)
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(contexto, app.SolicitudCambioAutorizacion{
		SolicitudAutorizacion: app.SolicitudAutorizacion{Apoderado: apoderado, EscolarID: escolar.ID},
		Autorizado:            false,
	}); err != nil {
		testingT.Fatal(err)
	}
	if autorizado, err := servicio.Autorizacion.AutorizacionVigente(contexto, escolar.ID); err != nil || autorizado {
		testingT.Fatalf("revocación persistida: %v (%v)", autorizado, err)
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(contexto, app.SolicitudCambioAutorizacion{
		SolicitudAutorizacion: app.SolicitudAutorizacion{Apoderado: apoderado, EscolarID: escolar.ID},
		Autorizado:            true,
	}); err != nil {
		testingT.Fatal(err)
	}
	var cambios int
	if err := baseDatos.QueryRowContext(contexto,
		`SELECT count(*) FROM autorizaciones_apoderado WHERE escolar_id=$1 AND apoderado_id=$2`,
		escolar.ID, apoderado.ID).Scan(&cambios); err != nil || cambios != 3 {
		testingT.Fatalf("historial de cambios: %d (%v)", cambios, err)
	}
	historial, err := servicio.Autorizacion.HistorialAutorizaciones(contexto, app.SolicitudAutorizacion{Apoderado: apoderado, EscolarID: escolar.ID})
	if err != nil || len(historial) != 3 {
		testingT.Fatalf("historial del vínculo: %+v (%v)", historial, err)
	}
	esperado := []modelo.EstadoAutorizacion{
		modelo.AutorizacionOtorgada, modelo.AutorizacionRevocada, modelo.AutorizacionOtorgada,
	}
	for indice, estado := range esperado {
		if historial[indice].Estado != estado {
			testingT.Fatalf("historial[%d]=%s, se esperaba %s", indice, historial[indice].Estado, estado)
		}
		if historial[indice].CreadaEn.IsZero() {
			testingT.Fatalf("el cambio %d debe llevar fecha", indice)
		}
	}
	if _, err := servicio.Vinculacion.AceptarVinculacion(contexto, escolar, tokenVinculacion); !errors.Is(err, modelo.ErrTokenInvalido) {
		testingT.Fatalf("un solo uso: %v", err)
	}

	actual, err := servicio.Vinculacion.VinculacionDeEscolar(contexto, escolar)
	if err != nil || actual == nil || actual.ApoderadoID != apoderado.ID {
		testingT.Fatalf("vínculo: %+v (%v)", actual, err)
	}
	otro, err := servicio.Registro.CrearAdministrado(contexto, modelo.SolicitudRegistro{
		Email: "otro.pg@example.com", NombreUso: "Otro PG", Clave: "clave-larga-ok", Rol: modelo.RolApoderado,
	})
	if err != nil {
		testingT.Fatal(err)
	}
	if _, err := servicio.Autorizacion.HistorialAutorizaciones(contexto, app.SolicitudAutorizacion{Apoderado: otro, EscolarID: escolar.ID}); !errors.Is(err, modelo.ErrVinculacionAjena) {
		testingT.Fatalf("otro apoderado no ve el historial: %v", err)
	}
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, otro, escolar.Email); !errors.Is(err, modelo.ErrYaVinculado) {
		testingT.Fatalf("escolar ya vinculado: %v", err)
	}
	vinculaciones, err := servicio.Vinculacion.VinculacionesDeApoderado(contexto, apoderado)
	if err != nil || len(vinculaciones) != 1 {
		testingT.Fatalf("listado: %+v (%v)", vinculaciones, err)
	}
}
