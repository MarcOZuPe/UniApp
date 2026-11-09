package acceso

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestVinculacionFlujoCompleto(testingT *testing.T) {
	servicio, bus, _, _ := nuevo(testingT)
	contexto := context.Background()
	apoderado := crearCuenta(testingT, servicio, datosCuenta{email: "apoderado@example.com", nombre: "Apoderado", rol: RolApoderado})
	escolar := crearCuenta(testingT, servicio, datosCuenta{email: "escolar@example.com", nombre: "Escolar", rol: RolEscolar})
	token := capturarTokenVinculacion(bus)

	creada, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, " Escolar@Example.COM ")
	if err != nil {
		testingT.Fatal(err)
	}
	if creada.Estado != EstadoVinculacionPendiente || *token == "" {
		testingT.Fatalf("solicitud pendiente con token: %+v", creada)
	}

	consultada, err := servicio.Vinculacion.ConsultarVinculacion(contexto, escolar, *token)
	if err != nil || consultada.ApoderadoEmail != apoderado.Email {
		testingT.Fatalf("consulta: %+v (%v)", consultada, err)
	}

	aceptada, err := servicio.Vinculacion.AceptarVinculacion(contexto, escolar, *token)
	if err != nil {
		testingT.Fatal(err)
	}
	if aceptada.Estado != EstadoVinculacionAceptada || aceptada.AceptadaEn == nil {
		testingT.Fatalf("aceptada: %+v", aceptada)
	}
	if aceptada.Autorizacion == nil || !aceptada.Autorizacion.Vigente() {
		testingT.Fatalf("al aceptar se otorga la autorización inicial: %+v", aceptada.Autorizacion)
	}
	if autorizado, err := servicio.Autorizacion.AutorizacionVigente(contexto, escolar.ID); err != nil || !autorizado {
		testingT.Fatalf("autorización inicial vigente: %v (%v)", autorizado, err)
	}
	if _, err := servicio.Vinculacion.AceptarVinculacion(contexto, escolar, *token); !errors.Is(err, ErrTokenInvalido) {
		testingT.Fatalf("el enlace es de un solo uso: %v", err)
	}

	revocada, err := servicio.Autorizacion.CambiarAutorizacion(contexto, cambioAutorizacion(apoderado, escolar.ID, false))
	if err != nil || revocada.Vigente() || revocada.Estado != AutorizacionRevocada {
		testingT.Fatalf("revocar: %+v (%v)", revocada, err)
	}
	if autorizado, err := servicio.Autorizacion.AutorizacionVigente(contexto, escolar.ID); err != nil || autorizado {
		testingT.Fatalf("tras revocar no puede reservar: %v (%v)", autorizado, err)
	}
	otorgada, err := servicio.Autorizacion.CambiarAutorizacion(contexto, cambioAutorizacion(apoderado, escolar.ID, true))
	if err != nil || !otorgada.Vigente() {
		testingT.Fatalf("volver a autorizar: %+v (%v)", otorgada, err)
	}
	if autorizado, err := servicio.Autorizacion.AutorizacionVigente(contexto, escolar.ID); err != nil || !autorizado {
		testingT.Fatalf("tras autorizar puede reservar: %v (%v)", autorizado, err)
	}

	actual, err := servicio.Vinculacion.VinculacionDeEscolar(contexto, escolar)
	if err != nil || actual == nil || actual.ApoderadoID != apoderado.ID {
		testingT.Fatalf("vínculo actual: %+v (%v)", actual, err)
	}

	otro := crearCuenta(testingT, servicio, datosCuenta{email: "otro@example.com", nombre: "Otro", rol: RolApoderado})
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, otro, escolar.Email); !errors.Is(err, ErrYaVinculado) {
		testingT.Fatalf("un escolar solo admite un apoderado: %v", err)
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(contexto, cambioAutorizacion(otro, escolar.ID, false)); !errors.Is(err, ErrVinculacionAjena) {
		testingT.Fatalf("otro apoderado no cambia la autorización: %v", err)
	}

	vinculaciones, err := servicio.Vinculacion.VinculacionesDeApoderado(contexto, apoderado)
	if err != nil || len(vinculaciones) != 1 || vinculaciones[0].EscolarEmail != escolar.Email {
		testingT.Fatalf("listado del apoderado: %+v (%v)", vinculaciones, err)
	}

	historial, err := servicio.Autorizacion.HistorialAutorizaciones(contexto, solicitudAutorizacion(apoderado, escolar.ID))
	if err != nil {
		testingT.Fatal(err)
	}
	if len(historial) != 3 {
		testingT.Fatalf("debía registrar aceptación, revocación y nueva autorización: %+v", historial)
	}
	esperado := []EstadoAutorizacion{AutorizacionOtorgada, AutorizacionRevocada, AutorizacionOtorgada}
	for i, estado := range esperado {
		if historial[i].Estado != estado {
			testingT.Fatalf("historial[%d]=%s, se esperaba %s", i, historial[i].Estado, estado)
		}
		if historial[i].CreadaEn.IsZero() {
			testingT.Fatalf("el cambio %d debe llevar fecha", i)
		}
	}
	if _, err := servicio.Autorizacion.HistorialAutorizaciones(contexto, solicitudAutorizacion(otro, escolar.ID)); !errors.Is(err, ErrVinculacionAjena) {
		testingT.Fatalf("otro apoderado no ve el historial: %v", err)
	}
	if _, err := servicio.Autorizacion.HistorialAutorizaciones(contexto, solicitudAutorizacion(escolar, escolar.ID)); !errors.Is(err, ErrNoEsApoderado) {
		testingT.Fatalf("solo el apoderado ve el historial: %v", err)
	}
}

func TestVinculacionSolicitudPendienteYUnApoderadoConVariosEscolares(testingT *testing.T) {
	servicio, bus, _, _ := nuevo(testingT)
	contexto := context.Background()
	apoderado := crearCuenta(testingT, servicio, datosCuenta{email: "apoderado@example.com", nombre: "Apoderado", rol: RolApoderado})
	escolarUno := crearCuenta(testingT, servicio, datosCuenta{email: "uno@example.com", nombre: "Uno", rol: RolEscolar})
	escolarDos := crearCuenta(testingT, servicio, datosCuenta{email: "dos@example.com", nombre: "Dos", rol: RolEscolar})
	token := capturarTokenVinculacion(bus)

	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, escolarUno.Email); err != nil {
		testingT.Fatal(err)
	}
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, escolarUno.Email); !errors.Is(err, ErrSolicitudPendiente) {
		testingT.Fatalf("no se duplica una solicitud pendiente: %v", err)
	}
	otro := crearCuenta(testingT, servicio, datosCuenta{email: "otro@example.com", nombre: "Otro", rol: RolApoderado})
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, otro, escolarUno.Email); !errors.Is(err, ErrSolicitudPendiente) {
		testingT.Fatalf("otro apoderado no desplaza la solicitud pendiente: %v", err)
	}
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, escolarDos.Email); err != nil {
		testingT.Fatalf("un apoderado puede solicitar varios escolares: %v", err)
	}
	if _, err := servicio.Vinculacion.AceptarVinculacion(contexto, escolarDos, *token); err != nil {
		testingT.Fatal(err)
	}
	vinculaciones, err := servicio.Vinculacion.VinculacionesDeApoderado(contexto, apoderado)
	if err != nil || len(vinculaciones) != 2 {
		testingT.Fatalf("debía tener dos escolares: %+v (%v)", vinculaciones, err)
	}
}

func TestVinculacionVenceYSePuedeVolverASolicitar(testingT *testing.T) {
	servicio, bus, ahora, _ := nuevo(testingT)
	contexto := context.Background()
	apoderado := crearCuenta(testingT, servicio, datosCuenta{email: "apoderado@example.com", nombre: "Apoderado", rol: RolApoderado})
	escolar := crearCuenta(testingT, servicio, datosCuenta{email: "escolar@example.com", nombre: "Escolar", rol: RolEscolar})
	token := capturarTokenVinculacion(bus)

	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, escolar.Email); err != nil {
		testingT.Fatal(err)
	}
	*ahora = ahora.Add(24*time.Hour + time.Second)
	if _, err := servicio.Vinculacion.ConsultarVinculacion(contexto, escolar, *token); !errors.Is(err, ErrTokenInvalido) {
		testingT.Fatalf("consulta vencida: %v", err)
	}
	if _, err := servicio.Vinculacion.AceptarVinculacion(contexto, escolar, *token); !errors.Is(err, ErrTokenInvalido) {
		testingT.Fatalf("aceptación vencida: %v", err)
	}
	vinculaciones, err := servicio.Vinculacion.VinculacionesDeApoderado(contexto, apoderado)
	if err != nil || len(vinculaciones) != 0 {
		testingT.Fatalf("la solicitud vencida no debe listarse: %+v (%v)", vinculaciones, err)
	}
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, escolar.Email); err != nil {
		testingT.Fatalf("tras vencer se puede volver a solicitar: %v", err)
	}
}

func TestVinculacionValidaRolesYDestino(testingT *testing.T) {
	servicio, _, _, _ := nuevo(testingT)
	contexto := context.Background()
	apoderado := crearCuenta(testingT, servicio, datosCuenta{email: "apoderado@example.com", nombre: "Apoderado", rol: RolApoderado})
	escolar := crearCuenta(testingT, servicio, datosCuenta{email: "escolar@example.com", nombre: "Escolar", rol: RolEscolar})
	tutor := crearCuenta(testingT, servicio, datosCuenta{email: "tutor@uni.edu.pe", nombre: "Tutor", rol: RolTutor})

	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, escolar, apoderado.Email); !errors.Is(err, ErrNoEsApoderado) {
		testingT.Fatalf("solo apoderado solicita: %v", err)
	}
	if _, err := servicio.Vinculacion.AceptarVinculacion(contexto, apoderado, "token"); !errors.Is(err, ErrNoEsEscolar) {
		testingT.Fatalf("solo escolar acepta: %v", err)
	}
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, "nadie@example.com"); !errors.Is(err, ErrEscolarNoEncontrado) {
		testingT.Fatalf("escolar inexistente: %v", err)
	}
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, tutor.Email); !errors.Is(err, ErrEscolarNoValido) {
		testingT.Fatalf("un tutor no es escolar: %v", err)
	}
	if _, err := servicio.Vinculacion.SolicitarVinculacion(contexto, apoderado, ""); !errors.As(err, new(ErrValidacion)) {
		testingT.Fatalf("correo vacío: %v", err)
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(contexto, cambioAutorizacion(apoderado, escolar.ID, false)); !errors.Is(err, ErrSinApoderado) {
		testingT.Fatalf("sin vínculo no hay autorización que cambiar: %v", err)
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(contexto, cambioAutorizacion(escolar, escolar.ID, false)); !errors.Is(err, ErrNoEsApoderado) {
		testingT.Fatalf("solo el apoderado cambia la autorización: %v", err)
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(contexto, cambioAutorizacion(apoderado, "", false)); !errors.As(err, new(ErrValidacion)) {
		testingT.Fatalf("escolar vacío: %v", err)
	}
}

func solicitudAutorizacion(apoderado *Usuario, escolarID string) SolicitudAutorizacion {
	return SolicitudAutorizacion{Apoderado: apoderado, EscolarID: escolarID}
}

func cambioAutorizacion(apoderado *Usuario, escolarID string, autorizado bool) SolicitudCambioAutorizacion {
	return SolicitudCambioAutorizacion{
		SolicitudAutorizacion: solicitudAutorizacion(apoderado, escolarID),
		Autorizado:            autorizado,
	}
}

type datosCuenta struct {
	email  string
	nombre string
	rol    Rol
}

func crearCuenta(testingT *testing.T, servicio *Servicio, datos datosCuenta) *Usuario {
	testingT.Helper()
	usuario, err := servicio.Registro.CrearAdministrado(context.Background(), SolicitudRegistro{
		Email: datos.email, NombreUso: datos.nombre, Clave: claveOK, Rol: datos.rol,
	})
	if err != nil {
		testingT.Fatal(err)
	}
	return usuario
}

func capturarTokenVinculacion(bus *busDePrueba) *string {
	token := new(string)
	bus.Registrar(VinculacionSolicitada, nuevoObservadorDePrueba("captura-vinculacion", func(evento Event) {
		*token, _ = evento.Payload["token"].(string)
	}))
	return token
}
