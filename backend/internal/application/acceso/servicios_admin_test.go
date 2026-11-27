package acceso

import (
	"context"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
)

func usuarioActivo(t *testing.T, repo *memoria.RepositorioMemoria, email string, rol Rol, facultad string) *Usuario {
	t.Helper()
	usuario, err := repo.CrearUsuario(context.Background(), &Usuario{
		Email: email, NombreUso: "Usuario", ClaveHash: "hash", Rol: rol,
		FacultadID: facultad, Activo: true, EmailConfirmado: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return usuario
}

func TestAdministracionService(t *testing.T) {
	repo := nuevoRepositorioPrueba()
	servicio, _ := servicioPrueba(t, repo, time.Now)
	ctx := context.Background()
	admin := usuarioActivo(t, repo, "admin@example.test", RolAdministrador, "")
	if _, err := servicio.Administracion.ListarGenerales(ctx, usuarioActivo(t, repo, "escolar@example.test", RolEscolar, "")); err != ErrNoAutorizado {
		t.Fatalf("se esperaba no autorizado: %v", err)
	}
	if _, err := servicio.Administracion.ListarGenerales(ctx, admin); err != nil {
		t.Fatal(err)
	}
	creado, err := servicio.Administracion.CrearGeneral(ctx, OrigenSolicitud{Actor: admin, IP: "127.0.0.1"}, DatosNuevoGeneral{
		Email: "general.nuevo@example.test", NombreUso: "General Nuevo", Clave: clavePrueba, FacultadID: "1",
	})
	if err != nil || creado.Rol != RolCoordinadorGeneral || creado.FacultadID != "1" {
		t.Fatalf("crear general inesperado: %+v err=%v", creado, err)
	}
	if _, err := servicio.Administracion.CrearGeneral(ctx, OrigenSolicitud{Actor: admin}, DatosNuevoGeneral{
		Email: "otro@example.test", NombreUso: "Otro", Clave: clavePrueba, FacultadID: "inexistente",
	}); err == nil {
		t.Fatal("una facultad inexistente debe rechazarse")
	}
	if _, err := servicio.Administracion.Logs(ctx, admin, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.Administracion.Logs(ctx, admin, 9999); err != nil {
		t.Fatal(err)
	}
	if err := servicio.Administracion.EliminarGeneral(ctx, OrigenSolicitud{Actor: admin}, creado.ID); err != nil {
		t.Fatal(err)
	}
	if err := servicio.Administracion.EliminarGeneral(ctx, OrigenSolicitud{Actor: admin}, creado.ID); err == nil {
		t.Fatal("eliminar dos veces debe fallar")
	}
}

func TestCoordinadoresService(t *testing.T) {
	repo := nuevoRepositorioPrueba()
	servicio, _ := servicioPrueba(t, repo, time.Now)
	ctx := context.Background()
	general := usuarioActivo(t, repo, "general@example.test", RolCoordinadorGeneral, "1")
	if _, err := servicio.Coordinadores.Listar(ctx, usuarioActivo(t, repo, "escolar@example.test", RolEscolar, "")); err != ErrNoAutorizado {
		t.Fatalf("se esperaba no autorizado: %v", err)
	}
	if _, err := servicio.Coordinadores.Listar(ctx, general); err != nil {
		t.Fatal(err)
	}
	creado, err := servicio.Coordinadores.Crear(ctx, general, SolicitudRegistro{
		Email: "coord.nuevo@example.test", NombreUso: "Coord Nuevo", Clave: clavePrueba, Rol: RolCoordinador,
	})
	if err != nil || creado.Rol != RolCoordinador || creado.FacultadID != "1" {
		t.Fatalf("crear coordinador inesperado: %+v err=%v", creado, err)
	}
	if _, err := servicio.Coordinadores.Crear(ctx, general, SolicitudRegistro{Email: "mal@example.test", NombreUso: "X", Clave: "corta"}); err == nil {
		t.Fatal("una clave inválida debe rechazarse")
	}
	if err := servicio.Coordinadores.Eliminar(ctx, general, "inexistente"); err == nil {
		t.Fatal("un coordinador inexistente debe fallar")
	}
	if err := servicio.Coordinadores.Eliminar(ctx, general, creado.ID); err != nil {
		t.Fatal(err)
	}
}

func TestPerfilServiceCambiarClave(t *testing.T) {
	repo := nuevoRepositorioPrueba()
	servicio, _ := servicioPrueba(t, repo, time.Now)
	ctx := context.Background()
	usuario := usuarioActivo(t, repo, "perfil@example.test", RolEscolar, "")
	hash, err := servicio.Hasher.Hash(clavePrueba)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ActualizarClave(ctx, usuario.ID, hash); err != nil {
		t.Fatal(err)
	}
	if err := servicio.Perfil.CambiarClave(ctx, usuario.ID, DatosCambioClave{ClaveActual: "incorrecta", ClaveNueva: "Otra-Clave-2026!"}); err != ErrClaveActualIncorrecta {
		t.Fatalf("se esperaba clave actual incorrecta: %v", err)
	}
	if err := servicio.Perfil.CambiarClave(ctx, usuario.ID, DatosCambioClave{ClaveActual: clavePrueba, ClaveNueva: clavePrueba}); err == nil {
		t.Fatal("la clave nueva debe ser distinta")
	}
	if err := servicio.Perfil.CambiarClave(ctx, usuario.ID, DatosCambioClave{ClaveActual: clavePrueba, ClaveNueva: "corta"}); err == nil {
		t.Fatal("la política debe rechazar una clave débil")
	}
	if err := servicio.Perfil.CambiarClave(ctx, usuario.ID, DatosCambioClave{ClaveActual: clavePrueba, ClaveNueva: "Otra-Clave-2026!", IP: "127.0.0.1"}); err != nil {
		t.Fatalf("cambio de clave falló: %v", err)
	}
	if err := servicio.Perfil.CambiarClave(ctx, "inexistente", DatosCambioClave{}); err != ErrNoEncontrado {
		t.Fatalf("se esperaba no encontrado: %v", err)
	}
}

func TestRegistroServiceFacultadesYRegistro(t *testing.T) {
	repo := nuevoRepositorioPrueba()
	servicio, _ := servicioPrueba(t, repo, time.Now)
	ctx := context.Background()
	if facultades, err := servicio.Registro.Facultades(ctx); err != nil || len(facultades) == 0 {
		t.Fatalf("facultades inesperadas: %+v err=%v", facultades, err)
	}
	registrado, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{
		Email: "nuevo@example.test", NombreUso: "Nuevo", Clave: clavePrueba, Rol: RolEscolar,
	})
	if err != nil || registrado.ID == "" {
		t.Fatalf("registro inesperado: %+v err=%v", registrado, err)
	}
	if _, err := servicio.Registro.Registrar(ctx, SolicitudRegistro{Email: "nuevo@example.test", NombreUso: "Nuevo", Clave: clavePrueba, Rol: RolEscolar}); err != ErrEmailEnUso {
		t.Fatalf("se esperaba correo en uso: %v", err)
	}
	if _, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "tutor.nuevo@uni.edu.pe", NombreUso: "Tutor", Clave: clavePrueba, Rol: RolTutor, FacultadID: "1",
	}); err != nil {
		t.Fatal(err)
	}
}
