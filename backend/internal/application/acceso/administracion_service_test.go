package acceso

import (
	"context"
	"errors"
	"testing"
)

func TestAdministradorGestionaGeneralesYConsultaLogs(t *testing.T) {
	servicio, _, _, _ := nuevo(t)
	ctx := context.Background()
	admin, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "admin@uni.edu.pe", NombreUso: "Administrador", Clave: claveOK, Rol: RolAdministrador,
	})
	if err != nil {
		t.Fatal(err)
	}
	generalCiencias, err := servicio.Registro.CrearAdministrado(ctx, SolicitudRegistro{
		Email: "general.ciencias@example.com", NombreUso: "General Ciencias", Clave: claveOK,
		Rol: RolCoordinadorGeneral, FacultadID: "1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.Administracion.ListarGenerales(ctx, generalCiencias); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("un general no debe administrar el sistema: %v", err)
	}
	creado, err := servicio.Administracion.CrearGeneral(ctx, admin, DatosNuevoGeneral{
		Email: "general.civil@example.com", NombreUso: "General Civil", Clave: claveOK, FacultadID: "2",
	}, "203.0.113.10")
	if err != nil || creado.Rol != RolCoordinadorGeneral || creado.FacultadID != "2" {
		t.Fatalf("crear general: %+v, %v", creado, err)
	}
	lista, err := servicio.Administracion.ListarGenerales(ctx, admin)
	if err != nil || len(lista) != 2 {
		t.Fatalf("listar generales: %+v, %v", lista, err)
	}
	if err := servicio.Administracion.EliminarGeneral(ctx, admin, creado.ID, "203.0.113.11"); err != nil {
		t.Fatal(err)
	}
	logs, err := servicio.Administracion.Logs(ctx, admin, 10)
	if err != nil || len(logs) < 2 {
		t.Fatalf("logs administrativos: %+v, %v", logs, err)
	}
	if logs[0].IP == "" || logs[0].Categoria != CategoriaAdministracion {
		t.Fatalf("log administrativo incompleto: %+v", logs[0])
	}
}
