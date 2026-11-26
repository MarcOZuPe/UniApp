package acceso

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPerfilCambiarClaveConservaIdentidadYRevocaSesiones(t *testing.T) {
	servicio, bus, _, _ := nuevo(t)
	ctx := context.Background()
	usuario, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "perfil@example.com", NombreUso: "Perfil", Clave: claveOK, Rol: RolEscolar,
	})
	if err != nil {
		t.Fatal(err)
	}
	tokenAnterior, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso(usuario.Email, claveOK))
	if err != nil {
		t.Fatal(err)
	}
	correos := make([]string, 0, 1)
	bus.Registrar(ClaveCambiada, nuevoObservadorDePrueba("correo-cambio", func(evento Event) {
		correos = append(correos, evento.Payload["email"].(string))
	}))

	if err := servicio.Perfil.CambiarClave(ctx, usuario.ID, DatosCambioClave{
		ClaveActual: "incorrecta", ClaveNueva: "Nueva-Clave-2026!",
	}); !errors.Is(err, ErrClaveActualIncorrecta) {
		t.Fatalf("clave actual incorrecta: %v", err)
	}
	if err := servicio.Perfil.CambiarClave(ctx, usuario.ID, DatosCambioClave{
		ClaveActual: claveOK, ClaveNueva: "corta",
	}); err == nil {
		t.Fatal("la nueva clave debe cumplir la política")
	}
	if err := servicio.Perfil.CambiarClave(ctx, usuario.ID, DatosCambioClave{
		ClaveActual: claveOK, ClaveNueva: "Nueva-Clave-2026!", IP: "198.51.100.7",
	}); err != nil {
		t.Fatal(err)
	}
	logs, err := servicio.RepositorioLogs.LogsSistema(ctx, FiltroLogsSistema{Limite: 10})
	if err != nil || len(logs) != 1 || logs[0].Evento != "clave.cambiada" || logs[0].IP != "198.51.100.7" {
		t.Fatalf("el cambio de clave debe registrar IP: %+v, %v", logs, err)
	}
	if _, err := servicio.Sesion.UsuarioDeSesion(ctx, tokenAnterior); err == nil {
		t.Fatal("el cambio de clave debe revocar las sesiones existentes")
	}
	if len(correos) != 1 || correos[0] != usuario.Email {
		t.Fatalf("debe notificar al correo del usuario: %v", correos)
	}
	if _, _, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso(usuario.Email, claveOK)); !errors.Is(err, ErrCredenciales) {
		t.Fatalf("la clave anterior no debe funcionar: %v", err)
	}
	if _, actualizado, err := servicio.Sesion.Ingresar(ctx, credencialesIngreso(usuario.Email, "Nueva-Clave-2026!")); err != nil {
		t.Fatal(err)
	} else if actualizado.Email != usuario.Email || actualizado.NombreUso != usuario.NombreUso {
		t.Fatalf("correo y nombre no deben cambiar: %+v", actualizado)
	}
}

func TestLimiteCambiosClave(t *testing.T) {
	servicio, _, ahora, _ := nuevo(t)
	ctx := context.Background()
	usuario, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "limite@example.com", NombreUso: "Límite", Clave: claveOK, Rol: RolEscolar,
	})
	if err != nil {
		t.Fatal(err)
	}
	cambiar := func(momento time.Time) error {
		return servicio.RepositorioPerfil.CambiarClaveConLimite(
			ctx, usuario.ID, "hash-nuevo", momento, 3, 1,
		)
	}
	if err := cambiar(*ahora); err != nil {
		t.Fatal(err)
	}
	if err := cambiar(ahora.Add(30 * time.Second)); !errors.Is(err, ErrDemasiadosCambiosClave) {
		t.Fatalf("debe limitar a un cambio por minuto: %v", err)
	}
	if err := cambiar(ahora.Add(2 * time.Minute)); err != nil {
		t.Fatalf("pasado un minuto debe permitir otro cambio: %v", err)
	}
	if err := cambiar(ahora.Add(4 * time.Minute)); err != nil {
		t.Fatalf("tercer cambio dentro de la hora: %v", err)
	}
	if err := cambiar(ahora.Add(6 * time.Minute)); !errors.Is(err, ErrDemasiadosCambiosClave) {
		t.Fatalf("debe limitar a tres cambios por hora: %v", err)
	}
}
