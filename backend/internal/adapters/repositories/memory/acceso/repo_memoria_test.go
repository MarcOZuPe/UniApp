package acceso

import (
	"context"
	"testing"
	"time"
)

func TestRepositorioMemoriaUsuariosSesionesYTokens(t *testing.T) {
	repo := NuevoRepositorioMemoria()
	ctx := context.Background()
	ahora := time.Now().UTC()
	if _, err := repo.UsuarioPorID(ctx, "inexistente"); err == nil {
		t.Fatal("un usuario inexistente debe fallar")
	}
	usuario, err := repo.CrearUsuario(ctx, &Usuario{
		Email: "user@example.test", NombreUso: "User", ClaveHash: "hash",
		Rol: RolEscolar, Activo: true, EmailConfirmado: true,
	})
	if err != nil || usuario.ID == "" {
		t.Fatalf("crear usuario falló: %+v err=%v", usuario, err)
	}
	if _, err := repo.CrearUsuario(ctx, &Usuario{Email: "user@example.test"}); err == nil {
		t.Fatal("un correo duplicado debe rechazarse")
	}
	if obtenido, err := repo.UsuarioPorEmail(ctx, "user@example.test"); err != nil || obtenido.ID != usuario.ID {
		t.Fatalf("usuario por email inesperado: %+v err=%v", obtenido, err)
	}
	activos, err := repo.UsuariosActivosPorRol(ctx, RolEscolar, "")
	if err != nil || len(activos) != 1 {
		t.Fatalf("usuarios activos inesperados: %+v err=%v", activos, err)
	}
	if err := repo.MarcarCorreoConfirmado(ctx, usuario.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.ActualizarClave(ctx, usuario.ID, "nuevo-hash"); err != nil {
		t.Fatal(err)
	}
	if err := repo.CambiarClaveConLimite(ctx, usuario.ID, "otro-hash", ahora, 5, 2); err != nil {
		t.Fatalf("cambio de clave falló: %v", err)
	}
	if err := repo.CambiarClaveConLimite(ctx, usuario.ID, "otro-hash", ahora, 5, 2); err != nil {
		t.Fatalf("segundo cambio de clave falló: %v", err)
	}
	if err := repo.CambiarClaveConLimite(ctx, usuario.ID, "otro-hash", ahora, 5, 2); err == nil {
		t.Fatal("el límite de cambios de clave debe aplicarse")
	}

	if err := repo.GuardarSesion(ctx, Sesion{TokenHash: []byte("sesion"), UsuarioID: usuario.ID, ExpiraEn: ahora.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	sesion, err := repo.SesionPorHash(ctx, []byte("sesion"))
	if err != nil || sesion.UsuarioID != usuario.ID {
		t.Fatalf("sesión inesperada: %+v err=%v", sesion, err)
	}
	if err := repo.EliminarSesion(ctx, []byte("sesion")); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SesionPorHash(ctx, []byte("sesion")); err == nil {
		t.Fatal("una sesión eliminada no debe encontrarse")
	}

	if err := repo.GuardarTokenRecuperacion(ctx, TokenRecuperacion{TokenHash: []byte("recuperacion"), UsuarioID: usuario.ID, ExpiraEn: ahora.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if id, err := repo.UsuarioDeTokenRecuperacion(ctx, []byte("recuperacion"), ahora); err != nil || id != usuario.ID {
		t.Fatalf("token de recuperación inesperado: %q err=%v", id, err)
	}
	if _, err := repo.ConsumirTokenRecuperacion(ctx, []byte("recuperacion"), ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ConsumirTokenRecuperacion(ctx, []byte("recuperacion"), ahora); err == nil {
		t.Fatal("un token consumido no debe reutilizarse")
	}

	if err := repo.GuardarTokenConfirmacion(ctx, TokenConfirmacion{TokenHash: []byte("confirmacion"), UsuarioID: usuario.ID, ExpiraEn: ahora.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if id, err := repo.ConsumirTokenConfirmacion(ctx, []byte("confirmacion"), ahora); err != nil || id != usuario.ID {
		t.Fatalf("token de confirmación inesperado: %q err=%v", id, err)
	}

	if err := repo.DesactivarUsuario(ctx, usuario.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UsuarioPorID(ctx, usuario.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.RegistrarIngreso(ctx, usuario.ID, ahora); err != nil {
		t.Fatal(err)
	}
	if err := repo.RegistrarLogSistema(ctx, LogSistema{Categoria: CategoriaSistema, Evento: "evento"}); err != nil {
		t.Fatal(err)
	}
	logs, err := repo.LogsSistema(ctx, FiltroLogsSistema{Limite: 1})
	if err != nil || len(logs) != 1 {
		t.Fatalf("logs inesperados: %+v err=%v", logs, err)
	}
	facultades, err := repo.Facultades(ctx)
	if err != nil || len(facultades) == 0 {
		t.Fatalf("facultades inesperadas: %+v err=%v", facultades, err)
	}
	if _, err := repo.FacultadPorID(ctx, "1"); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FacultadPorID(ctx, "inexistente"); err == nil {
		t.Fatal("una facultad inexistente debe fallar")
	}
	if _, err := repo.EliminarUsuariosSinConfirmar(ctx, ahora); err != nil {
		t.Fatal(err)
	}
}

func TestRepositorioMemoriaVinculacionesYAutorizaciones(t *testing.T) {
	repo := NuevoRepositorioMemoria()
	ctx := context.Background()
	ahora := time.Now().UTC()
	apoderado, _ := repo.CrearUsuario(ctx, &Usuario{Email: "a@example.test", NombreUso: "A", Rol: RolApoderado, Activo: true, ClaveHash: "h"})
	escolar, _ := repo.CrearUsuario(ctx, &Usuario{Email: "e@example.test", NombreUso: "E", Rol: RolEscolar, Activo: true, ClaveHash: "h"})
	if _, err := repo.SolicitudPendienteDeEscolar(ctx, escolar.ID, ahora); err == nil {
		t.Fatal("sin solicitud debe fallar")
	}
	solicitud, err := repo.GuardarSolicitudVinculacion(ctx, Vinculacion{
		ApoderadoID: apoderado.ID, EscolarID: escolar.ID, Estado: EstadoVinculacionPendiente,
		TokenHash: []byte("token"), ExpiraEn: ahora.Add(time.Hour),
	})
	if err != nil || solicitud.ID == "" {
		t.Fatalf("solicitud inesperada: %+v err=%v", solicitud, err)
	}
	if _, err := repo.SolicitudPendienteDeEscolar(ctx, escolar.ID, ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.SolicitudVinculacionPorToken(ctx, ConsultaVinculacionPorToken{TokenHash: []byte("token"), EscolarID: escolar.ID, Ahora: ahora}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ConsumirSolicitudVinculacion(ctx, ConsultaVinculacionPorToken{TokenHash: []byte("token"), EscolarID: escolar.ID, Ahora: ahora}); err != nil {
		t.Fatal(err)
	}
	aceptada, err := repo.VinculacionAceptadaDeEscolar(ctx, escolar.ID)
	if err != nil || aceptada.Estado != EstadoVinculacionAceptada || aceptada.Autorizacion == nil {
		t.Fatalf("vinculación aceptada inesperada: %+v err=%v", aceptada, err)
	}
	vinculaciones, err := repo.VinculacionesDeApoderado(ctx, apoderado.ID, ahora)
	if err != nil || len(vinculaciones) != 1 {
		t.Fatalf("vinculaciones inesperadas: %+v err=%v", vinculaciones, err)
	}
	if vigente, err := repo.AutorizacionVigenteDeEscolar(ctx, escolar.ID); err != nil || !vigente.Vigente() {
		t.Fatalf("autorización vigente inesperada: %+v err=%v", vigente, err)
	}
	cambio, err := repo.CambiarAutorizacion(ctx, CambioAutorizacion{
		VinculacionID: aceptada.ID, EscolarID: escolar.ID, ApoderadoID: apoderado.ID,
		Estado: AutorizacionRevocada, Momento: ahora.Add(time.Minute),
	})
	if err != nil || cambio.Vigente() {
		t.Fatalf("cambio de autorización inesperado: %+v err=%v", cambio, err)
	}
	historial, err := repo.AutorizacionesDeVinculacion(ctx, aceptada.ID)
	if err != nil || len(historial) != 2 {
		t.Fatalf("historial inesperado: %+v err=%v", historial, err)
	}
	if _, err := repo.SolicitudVinculacionPorToken(ctx, ConsultaVinculacionPorToken{TokenHash: []byte("malo"), EscolarID: escolar.ID, Ahora: ahora}); err == nil {
		t.Fatal("un token inválido debe rechazarse")
	}
}
