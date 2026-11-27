package catalogo

import (
	"context"
	"testing"

	memoria "uniapp/internal/adapters/repositories/memory/catalogo"
)

func TestServicioCatalogoFlujoCompleto(t *testing.T) {
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria())
	ctx := context.Background()

	importacion, err := servicio.Importar(ctx, DatosImportacion{
		Fuente:   "origen",
		Cursos:   []CursoEntrada{{Nombre: "Álgebra", Grado: "5to"}},
		Colegios: []ColegioEntrada{{Nombre: "Colegio", Distrito: "Lima"}},
	})
	if err != nil {
		t.Fatalf("importar falló: %v", err)
	}
	if importacion.ID == 0 {
		t.Fatal("la importación debería tener ID")
	}
	publica, err := servicio.CatalogoPublico(ctx)
	if err != nil || len(publica.Cursos) != 0 {
		t.Fatalf("el catálogo público no debe mostrar cursos pendientes: %+v err=%v", publica, err)
	}
	pendientes, err := servicio.CursosPendientes(ctx)
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("pendientes inesperados: %+v err=%v", pendientes, err)
	}
	if _, err := servicio.RevisarCurso(ctx, pendientes[0].ID, "coord-1", EstadoAprobado, ""); err != nil {
		t.Fatalf("aprobar curso falló: %v", err)
	}
	publica, err = servicio.CatalogoPublico(ctx)
	if err != nil || len(publica.Cursos) != 1 || publica.Cursos[0].Estado != EstadoAprobado {
		t.Fatalf("el catálogo público debe mostrar el curso aprobado: %+v err=%v", publica, err)
	}
	if _, err := servicio.RevisarCurso(ctx, pendientes[0].ID, "coord-1", EstadoAprobado, ""); err == nil {
		t.Fatal("un curso ya revisado no debe revisarse otra vez")
	}
	if _, err := servicio.RevisarCurso(ctx, 9999, "coord-1", EstadoAprobado, ""); err == nil {
		t.Fatal("un curso inexistente debe fallar")
	}
}

func TestServicioSolicitudesDeColegio(t *testing.T) {
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria())
	ctx := context.Background()
	if _, err := servicio.Importar(ctx, DatosImportacion{
		Fuente:   "origen",
		Cursos:   []CursoEntrada{{Nombre: "Álgebra", Grado: "5to"}},
		Colegios: []ColegioEntrada{{Nombre: "Colegio", Distrito: "Lima"}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.SolicitarColegio(ctx, "", SolicitudColegioEntrada{Nombre: "Otro", Distrito: "Surco"}); err == nil {
		t.Fatal("un solicitante vacío debe rechazarse")
	}
	if _, err := servicio.SolicitarColegio(ctx, "u-1", SolicitudColegioEntrada{Nombre: "Colegio", Distrito: "Lima"}); err != ErrColegioYaExiste {
		t.Fatalf("se esperaba ErrColegioYaExiste, se obtuvo %v", err)
	}
	solicitud, err := servicio.SolicitarColegio(ctx, "u-1", SolicitudColegioEntrada{Nombre: "Otro", Distrito: "Surco"})
	if err != nil || solicitud.Estado != EstadoSolicitudPendiente {
		t.Fatalf("solicitud inesperada: %+v err=%v", solicitud, err)
	}
	if _, err := servicio.SolicitarColegio(ctx, "u-1", SolicitudColegioEntrada{Nombre: "Otro", Distrito: "Surco"}); err == nil {
		t.Fatal("una solicitud duplicada debe rechazarse")
	}
	pendientes, err := servicio.SolicitudesPendientes(ctx)
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("solicitudes pendientes inesperadas: %+v err=%v", pendientes, err)
	}
}
