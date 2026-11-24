package acceso

import (
	"context"
	"errors"
	"testing"
)

func TestCoordinadorGeneralGestionaCoordinadores(t *testing.T) {
	servicio, _, _, _ := nuevo(t)
	ctx := context.Background()
	general, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "general@example.com", NombreUso: "General", Clave: claveOK, Rol: RolCoordinadorGeneral,
	})
	if err != nil {
		t.Fatal(err)
	}
	regular, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "coord@example.com", NombreUso: "Coord", Clave: claveOK, Rol: RolCoordinador,
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := servicio.Coordinadores.Listar(ctx, regular); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("un coordinador regular no debe gestionar cuentas: %v", err)
	}
	creado, err := servicio.Coordinadores.Crear(ctx, general, SolicitudRegistro{
		Email: "nuevo@example.com", NombreUso: "Nuevo", Clave: claveOK,
	})
	if err != nil || creado.Rol != RolCoordinador || !creado.EmailConfirmado {
		t.Fatalf("crear coordinador: %+v, %v", creado, err)
	}
	listado, err := servicio.Coordinadores.Listar(ctx, general)
	if err != nil || len(listado) != 2 {
		t.Fatalf("listar coordinadores: %+v, %v", listado, err)
	}
	if err := servicio.Coordinadores.Eliminar(ctx, general, creado.ID); err != nil {
		t.Fatal(err)
	}
	eliminado, err := servicio.RepositorioUsuarios.UsuarioPorID(ctx, creado.ID)
	if err != nil || eliminado.Activo {
		t.Fatalf("el coordinador eliminado debe quedar inactivo: %+v, %v", eliminado, err)
	}
	if err := servicio.Coordinadores.Eliminar(ctx, general, general.ID); !errors.Is(err, ErrCoordinadorNoEncontrado) {
		t.Fatalf("no se debe eliminar una cuenta general por esta vía: %v", err)
	}

	otraFacultad, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "general2@example.com", NombreUso: "General Dos", Clave: claveOK,
		Rol: RolCoordinadorGeneral, FacultadID: "2",
	})
	if err != nil {
		t.Fatal(err)
	}
	if listado, err := servicio.Coordinadores.Listar(ctx, otraFacultad); err != nil || len(listado) != 0 {
		t.Fatalf("no debe listar coordinadores de otra facultad: %+v, %v", listado, err)
	}
	if err := servicio.Coordinadores.Eliminar(ctx, otraFacultad, regular.ID); !errors.Is(err, ErrCoordinadorNoEncontrado) {
		t.Fatalf("no debe eliminar coordinadores de otra facultad: %v", err)
	}
}
