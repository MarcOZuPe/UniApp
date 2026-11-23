package acceso

import (
	"context"
	"errors"
	"testing"
)

func TestPerfilCambiarClaveYConservarIdentidad(t *testing.T) {
	servicio, _, _, _ := nuevo(t)
	ctx := context.Background()
	usuario, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "perfil@example.com", NombreUso: "Perfil", Clave: claveOK, Rol: RolEscolar,
	})
	if err != nil {
		t.Fatal(err)
	}
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
		ClaveActual: claveOK, ClaveNueva: "Nueva-Clave-2026!",
	}); err != nil {
		t.Fatal(err)
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
