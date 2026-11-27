package acceso

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

func TestTokenSesionEsAutocontenidoYVerificable(t *testing.T) {
	firmador := NuevoFirmadorSesion("clave-de-prueba-suficientemente-larga")
	usuario := &Usuario{ID: "u-1", ClaveHash: "hash-de-clave"}
	ahora := time.Now().UTC()

	token, err := firmador.Firmar(usuario, ahora.Add(time.Hour))
	if err != nil {
		t.Fatalf("no se pudo firmar el token: %v", err)
	}

	sesion, err := firmador.Verificar(token, ahora)
	if err != nil {
		t.Fatalf("el token recién emitido debería ser válido: %v", err)
	}
	if sesion.UsuarioID != usuario.ID {
		t.Fatalf("UsuarioID=%q, se esperaba %q", sesion.UsuarioID, usuario.ID)
	}
	if sesion.HuellaClave != huellaEsperada(usuario.ClaveHash) {
		t.Fatal("la huella de clave no coincide")
	}
}

func huellaEsperada(claveHash string) string {
	suma := sha256.Sum256([]byte(claveHash))
	return hex.EncodeToString(suma[:])
}

func TestTokenSesionRechazaManipulacionYVencimiento(t *testing.T) {
	firmador := NuevoFirmadorSesion("clave-de-prueba-suficientemente-larga")
	usuario := &Usuario{ID: "u-1", ClaveHash: "hash-de-clave"}
	ahora := time.Now().UTC()

	token, err := firmador.Firmar(usuario, ahora.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	alterado := token[:len(token)-1] + "x"
	if _, err := firmador.Verificar(alterado, ahora); err == nil {
		t.Fatal("un token alterado no debe aceptarse")
	}

	vencido, err := firmador.Firmar(usuario, ahora.Add(-time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := firmador.Verificar(vencido, ahora); err == nil {
		t.Fatal("un token vencido no debe aceptarse")
	}
	for _, invalido := range []string{"", "sin-firma", "a.b.c", "....", "!!!.###"} {
		if _, err := firmador.Verificar(invalido, ahora); err == nil {
			t.Fatalf("el token malformado %q no debe aceptarse", invalido)
		}
	}
}

func TestTokenSesionNoAceptaOtraClaveDeFirma(t *testing.T) {
	usuario := &Usuario{ID: "u-1", ClaveHash: "hash-de-clave"}
	token, err := NuevoFirmadorSesion("clave-emisora").Firmar(usuario, time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NuevoFirmadorSesion("clave-distinta").Verificar(token, time.Now().UTC()); err == nil {
		t.Fatal("un token firmado con otra clave no debe aceptarse")
	}
}

func TestSesionSinEstadoInvalidaTokenAlCambiarClaveODesactivar(t *testing.T) {
	repositorio := nuevoRepositorioPrueba()
	servicio, _ := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()

	hash, err := servicio.Hasher.Hash(clavePrueba)
	if err != nil {
		t.Fatal(err)
	}
	creado, err := repositorio.CrearUsuario(ctx, &Usuario{
		Email: "sesion@example.test", NombreUso: "Sesión", ClaveHash: hash,
		Rol: RolEscolar, Activo: true, EmailConfirmado: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	token, _, err := servicio.Sesion.Ingresar(ctx, CredencialesIngreso{
		Email: "sesion@example.test", Clave: clavePrueba,
	})
	if err != nil {
		t.Fatalf("no se pudo iniciar sesión: %v", err)
	}
	usuario, err := servicio.Sesion.UsuarioDeSesion(ctx, token)
	if err != nil || usuario.ID != creado.ID {
		t.Fatalf("el token debería identificar al usuario: usuario=%v err=%v", usuario, err)
	}

	nuevoHash, err := servicio.Hasher.Hash("Otra-Clave-2026!")
	if err != nil {
		t.Fatal(err)
	}
	if err := repositorio.ActualizarClave(ctx, creado.ID, nuevoHash); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.Sesion.UsuarioDeSesion(ctx, token); err == nil {
		t.Fatal("cambiar la clave debería invalidar los tokens anteriores")
	}

	// Un token nuevo sigue siendo válido; desactivar la cuenta lo invalida.
	token, _, err = servicio.Sesion.Ingresar(ctx, CredencialesIngreso{
		Email: "sesion@example.test", Clave: "Otra-Clave-2026!",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repositorio.DesactivarUsuario(ctx, creado.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.Sesion.UsuarioDeSesion(ctx, token); err == nil {
		t.Fatal("desactivar la cuenta debería invalidar sus tokens")
	}
}

// Design defect: con sesiones autocontenidas, cerrar sesión solo elimina la
// cookie del cliente. El token robado sigue siendo válido hasta expirar.
func TestDesignLogoutDeberiaInvalidarElToken(t *testing.T) {
	repositorio := nuevoRepositorioPrueba()
	servicio, _ := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()

	hash, err := servicio.Hasher.Hash(clavePrueba)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repositorio.CrearUsuario(ctx, &Usuario{
		Email: "logout@example.test", NombreUso: "Logout", ClaveHash: hash,
		Rol: RolEscolar, Activo: true, EmailConfirmado: true,
	}); err != nil {
		t.Fatal(err)
	}
	token, _, err := servicio.Sesion.Ingresar(ctx, CredencialesIngreso{
		Email: "logout@example.test", Clave: clavePrueba,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := servicio.Sesion.Salir(ctx, token); err != nil {
		t.Fatalf("cerrar sesión devolvió error: %v", err)
	}
	if _, err := servicio.Sesion.UsuarioDeSesion(ctx, token); err == nil {
		t.Fatal("D-01: el token sigue siendo válido después de cerrar sesión; un token robado no se revoca")
	}
}
