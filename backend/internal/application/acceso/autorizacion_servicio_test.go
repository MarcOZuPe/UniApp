package acceso

import (
	"context"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
)

func prepararApoderadoVinculado(t *testing.T) (*Servicio, *memoria.RepositorioMemoria, *Usuario, string) {
	t.Helper()
	repositorio := nuevoRepositorioPrueba()
	servicio, _ := servicioPrueba(t, repositorio, time.Now)
	ctx := context.Background()
	apoderado, err := repositorio.CrearUsuario(ctx, &Usuario{
		Email: "apoderado@example.test", NombreUso: "Apoderado", ClaveHash: "hash",
		Rol: RolApoderado, Activo: true, EmailConfirmado: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	escolar, err := repositorio.CrearUsuario(ctx, &Usuario{
		Email: "escolar@example.test", NombreUso: "Escolar", ClaveHash: "hash",
		Rol: RolEscolar, Activo: true, EmailConfirmado: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	vinculacion, err := repositorio.GuardarSolicitudVinculacion(ctx, Vinculacion{
		ApoderadoID: apoderado.ID, EscolarID: escolar.ID, Estado: EstadoVinculacionPendiente,
		TokenHash: []byte("token"), ExpiraEn: servicio.Ahora().Add(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repositorio.ConsumirSolicitudVinculacion(ctx, ConsultaVinculacionPorToken{
		TokenHash: []byte("token"), EscolarID: escolar.ID, Ahora: servicio.Ahora(),
	}); err != nil {
		t.Fatal(err)
	}
	if vinculacion.ID == "" {
		t.Fatal("la vinculación debería tener ID")
	}
	return servicio, repositorio, apoderado, escolar.ID
}

func TestAutorizacionServiceCambiarYConsultar(t *testing.T) {
	servicio, _, apoderado, escolarID := prepararApoderadoVinculado(t)
	ctx := context.Background()

	if vigente, err := servicio.Autorizacion.AutorizacionVigente(ctx, escolarID); err != nil || !vigente {
		t.Fatalf("la autorización inicial debería estar vigente: %v %v", vigente, err)
	}
	revocada, err := servicio.Autorizacion.CambiarAutorizacion(ctx, SolicitudCambioAutorizacion{
		SolicitudAutorizacion: SolicitudAutorizacion{Apoderado: apoderado, EscolarID: escolarID},
		Autorizado:            false,
	})
	if err != nil || revocada.Vigente() {
		t.Fatalf("la autorización debía revocarse: %+v err=%v", revocada, err)
	}
	if vigente, err := servicio.Autorizacion.AutorizacionVigente(ctx, escolarID); err != nil || vigente {
		t.Fatalf("la autorización no debería estar vigente: %v %v", vigente, err)
	}
	historial, err := servicio.Autorizacion.HistorialAutorizaciones(ctx, SolicitudAutorizacion{
		Apoderado: apoderado, EscolarID: escolarID,
	})
	if err != nil || len(historial) < 2 {
		t.Fatalf("el historial debería tener al menos dos cambios: %+v err=%v", historial, err)
	}
	otorgada, err := servicio.Autorizacion.CambiarAutorizacion(ctx, SolicitudCambioAutorizacion{
		SolicitudAutorizacion: SolicitudAutorizacion{Apoderado: apoderado, EscolarID: escolarID},
		Autorizado:            true,
	})
	if err != nil || !otorgada.Vigente() {
		t.Fatalf("la autorización debía otorgarse: %+v err=%v", otorgada, err)
	}
}

func TestAutorizacionServiceRechazaSolicitudesInvalidas(t *testing.T) {
	servicio, _, apoderado, escolarID := prepararApoderadoVinculado(t)
	ctx := context.Background()
	if _, err := servicio.Autorizacion.CambiarAutorizacion(ctx, SolicitudCambioAutorizacion{}); err == nil {
		t.Fatal("un apoderado nulo debe rechazarse")
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(ctx, SolicitudCambioAutorizacion{
		SolicitudAutorizacion: SolicitudAutorizacion{Apoderado: &Usuario{ID: "otro", Rol: RolApoderado, Activo: true}, EscolarID: escolarID},
	}); err == nil {
		t.Fatal("un apoderado ajeno debe rechazarse")
	}
	if _, err := servicio.Autorizacion.CambiarAutorizacion(ctx, SolicitudCambioAutorizacion{
		SolicitudAutorizacion: SolicitudAutorizacion{Apoderado: apoderado, EscolarID: "  "},
	}); err == nil {
		t.Fatal("un escolar vacío debe rechazarse")
	}
	if vigente, err := servicio.Autorizacion.AutorizacionVigente(ctx, "  "); err != nil || vigente {
		t.Fatalf("un escolar vacío no debe tener autorización: %v %v", vigente, err)
	}
	if vigente, err := servicio.Autorizacion.AutorizacionVigente(ctx, "u-inexistente"); err != nil || vigente {
		t.Fatalf("un escolar inexistente no debe tener autorización: %v %v", vigente, err)
	}
}
