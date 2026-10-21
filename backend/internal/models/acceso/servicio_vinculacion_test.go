package acceso

import (
	"context"
	"errors"
	"testing"
	"time"
)

func crearCuenta(t *testing.T, s *Servicio, email, nombre string, rol Rol) *Usuario {
	t.Helper()
	usuario, err := s.CrearAdministrado(context.Background(), DatosRegistro{
		Email: email, NombreUso: nombre, Clave: claveOK, Rol: rol,
	})
	if err != nil {
		t.Fatal(err)
	}
	return usuario
}

func capturarTokenVinculacion(bus *Bus) *string {
	token := new(string)
	bus.Registrar(VinculacionSolicitada, NuevoObservadorFunc("captura-vinculacion", func(e Event) {
		*token, _ = e.Payload["token"].(string)
	}))
	return token
}

func TestVinculacionFlujoCompleto(t *testing.T) {
	s, bus, _, _ := nuevo(t)
	ctx := context.Background()
	apoderado := crearCuenta(t, s, "apoderado@example.com", "Apoderado", RolApoderado)
	escolar := crearCuenta(t, s, "escolar@example.com", "Escolar", RolEscolar)
	token := capturarTokenVinculacion(bus)

	creada, err := s.SolicitarVinculacion(ctx, apoderado, " Escolar@Example.COM ")
	if err != nil {
		t.Fatal(err)
	}
	if creada.Estado != EstadoVinculacionPendiente || *token == "" {
		t.Fatalf("solicitud pendiente con token: %+v", creada)
	}

	consultada, err := s.ConsultarVinculacion(ctx, escolar, *token)
	if err != nil || consultada.ApoderadoEmail != apoderado.Email {
		t.Fatalf("consulta: %+v (%v)", consultada, err)
	}

	aceptada, err := s.AceptarVinculacion(ctx, escolar, *token)
	if err != nil {
		t.Fatal(err)
	}
	if aceptada.Estado != EstadoVinculacionAceptada || aceptada.AceptadaEn == nil {
		t.Fatalf("aceptada: %+v", aceptada)
	}
	if _, err := s.AceptarVinculacion(ctx, escolar, *token); !errors.Is(err, ErrTokenInvalido) {
		t.Fatalf("el enlace es de un solo uso: %v", err)
	}

	actual, err := s.VinculacionDeEscolar(ctx, escolar)
	if err != nil || actual == nil || actual.ApoderadoID != apoderado.ID {
		t.Fatalf("vínculo actual: %+v (%v)", actual, err)
	}

	otro := crearCuenta(t, s, "otro@example.com", "Otro", RolApoderado)
	if _, err := s.SolicitarVinculacion(ctx, otro, escolar.Email); !errors.Is(err, ErrYaVinculado) {
		t.Fatalf("un escolar solo admite un apoderado: %v", err)
	}

	vinculaciones, err := s.VinculacionesDeApoderado(ctx, apoderado)
	if err != nil || len(vinculaciones) != 1 || vinculaciones[0].EscolarEmail != escolar.Email {
		t.Fatalf("listado del apoderado: %+v (%v)", vinculaciones, err)
	}
}

func TestVinculacionSolicitudPendienteYUnApoderadoConVariosEscolares(t *testing.T) {
	s, bus, _, _ := nuevo(t)
	ctx := context.Background()
	apoderado := crearCuenta(t, s, "apoderado@example.com", "Apoderado", RolApoderado)
	escolarUno := crearCuenta(t, s, "uno@example.com", "Uno", RolEscolar)
	escolarDos := crearCuenta(t, s, "dos@example.com", "Dos", RolEscolar)
	token := capturarTokenVinculacion(bus)

	if _, err := s.SolicitarVinculacion(ctx, apoderado, escolarUno.Email); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SolicitarVinculacion(ctx, apoderado, escolarUno.Email); !errors.Is(err, ErrSolicitudPendiente) {
		t.Fatalf("no se duplica una solicitud pendiente: %v", err)
	}
	otro := crearCuenta(t, s, "otro@example.com", "Otro", RolApoderado)
	if _, err := s.SolicitarVinculacion(ctx, otro, escolarUno.Email); !errors.Is(err, ErrSolicitudPendiente) {
		t.Fatalf("otro apoderado no desplaza la solicitud pendiente: %v", err)
	}
	if _, err := s.SolicitarVinculacion(ctx, apoderado, escolarDos.Email); err != nil {
		t.Fatalf("un apoderado puede solicitar varios escolares: %v", err)
	}
	if _, err := s.AceptarVinculacion(ctx, escolarDos, *token); err != nil {
		t.Fatal(err)
	}
	vinculaciones, err := s.VinculacionesDeApoderado(ctx, apoderado)
	if err != nil || len(vinculaciones) != 2 {
		t.Fatalf("debía tener dos escolares: %+v (%v)", vinculaciones, err)
	}
}

func TestVinculacionVenceYSePuedeVolverASolicitar(t *testing.T) {
	s, bus, ahora, _ := nuevo(t)
	ctx := context.Background()
	apoderado := crearCuenta(t, s, "apoderado@example.com", "Apoderado", RolApoderado)
	escolar := crearCuenta(t, s, "escolar@example.com", "Escolar", RolEscolar)
	token := capturarTokenVinculacion(bus)

	if _, err := s.SolicitarVinculacion(ctx, apoderado, escolar.Email); err != nil {
		t.Fatal(err)
	}
	*ahora = ahora.Add(24*time.Hour + time.Second)
	if _, err := s.ConsultarVinculacion(ctx, escolar, *token); !errors.Is(err, ErrTokenInvalido) {
		t.Fatalf("consulta vencida: %v", err)
	}
	if _, err := s.AceptarVinculacion(ctx, escolar, *token); !errors.Is(err, ErrTokenInvalido) {
		t.Fatalf("aceptación vencida: %v", err)
	}
	vinculaciones, err := s.VinculacionesDeApoderado(ctx, apoderado)
	if err != nil || len(vinculaciones) != 0 {
		t.Fatalf("la solicitud vencida no debe listarse: %+v (%v)", vinculaciones, err)
	}
	if _, err := s.SolicitarVinculacion(ctx, apoderado, escolar.Email); err != nil {
		t.Fatalf("tras vencer se puede volver a solicitar: %v", err)
	}
}

func TestVinculacionValidaRolesYDestino(t *testing.T) {
	s, _, _, _ := nuevo(t)
	ctx := context.Background()
	apoderado := crearCuenta(t, s, "apoderado@example.com", "Apoderado", RolApoderado)
	escolar := crearCuenta(t, s, "escolar@example.com", "Escolar", RolEscolar)
	tutor := crearCuenta(t, s, "tutor@uni.edu.pe", "Tutor", RolTutor)

	if _, err := s.SolicitarVinculacion(ctx, escolar, apoderado.Email); !errors.Is(err, ErrNoEsApoderado) {
		t.Fatalf("solo apoderado solicita: %v", err)
	}
	if _, err := s.AceptarVinculacion(ctx, apoderado, "token"); !errors.Is(err, ErrNoEsEscolar) {
		t.Fatalf("solo escolar acepta: %v", err)
	}
	if _, err := s.SolicitarVinculacion(ctx, apoderado, "nadie@example.com"); !errors.Is(err, ErrEscolarNoEncontrado) {
		t.Fatalf("escolar inexistente: %v", err)
	}
	if _, err := s.SolicitarVinculacion(ctx, apoderado, tutor.Email); !errors.Is(err, ErrEscolarNoValido) {
		t.Fatalf("un tutor no es escolar: %v", err)
	}
	if _, err := s.SolicitarVinculacion(ctx, apoderado, ""); !errors.As(err, new(ErrValidacion)) {
		t.Fatalf("correo vacío: %v", err)
	}
}
