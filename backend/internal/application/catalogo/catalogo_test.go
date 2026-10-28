package catalogo

import (
	"context"
	"errors"
	"testing"
	"time"
	memoria "uniapp/internal/adapters/repositories/memory/catalogo"
)

func TestImportacionValidaYRegistraFuenteYFecha(t *testing.T) {
	ahora := time.Date(2026, 10, 22, 12, 0, 0, 0, time.FixedZone("Lima", -5*60*60))
	importacion, err := NuevaImportacion(datosValidosCatalogo(), ahora)
	if err != nil {
		t.Fatal(err)
	}
	if importacion.Fuente != "sintetico: prueba" || !importacion.ImportadoEn.Equal(ahora.UTC()) {
		t.Fatalf("metadatos de importación: %+v", importacion)
	}
	if importacion.Cursos[0].Estado != EstadoPendiente {
		t.Fatalf("curso importado sin revisión debe quedar pendiente: %+v", importacion.Cursos[0])
	}

	datos := datosValidosCatalogo()
	datos.Cursos = append(datos.Cursos, CursoEntrada{Nombre: " álgebra ", Grado: "1.° de secundaria"})
	if _, err := NuevaImportacion(datos, ahora); err == nil {
		t.Fatal("debe rechazarse un curso repetido para el mismo grado")
	}
}

func TestCatalogoPublicoIncluyeSoloCursosAprobadosDeUltimaImportacion(t *testing.T) {
	ctx := context.Background()
	repositorio := memoria.NuevoRepositorioMemoria()
	servicio := NuevoServicio(repositorio)
	servicio.ahora = func() time.Time { return time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC) }

	importada, err := servicio.Importar(ctx, datosValidosCatalogo())
	if err != nil {
		t.Fatal(err)
	}
	publica, err := servicio.CatalogoPublico(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(publica.Cursos) != 0 || len(publica.Colegios) != 1 || publica.Fuente != importada.Fuente {
		t.Fatalf("la vista pública inicial debe ocultar cursos pendientes y conservar el padrón: %+v", publica)
	}

	aprobado, err := servicio.RevisarCurso(ctx, importada.Cursos[0].ID, "coordinacion-1", EstadoAprobado, "")
	if err != nil {
		t.Fatal(err)
	}
	if aprobado.Grado != "1.° de secundaria" {
		t.Fatalf("el grado aprobado se debe conservar: %+v", aprobado)
	}
	if aprobado.RevisadoPor != "coordinacion-1" || aprobado.RevisadoEn == nil {
		t.Fatalf("la aprobación debe auditar responsable y fecha: %+v", aprobado)
	}
	publica, err = servicio.CatalogoPublico(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(publica.Cursos) != 1 || publica.Cursos[0].ID != aprobado.ID {
		t.Fatalf("solo debe publicarse el curso aprobado: %+v", publica.Cursos)
	}
	if _, err := servicio.RevisarCurso(ctx, aprobado.ID, "coordinacion-1", EstadoAprobado, ""); !errors.Is(err, ErrEstadoInvalido) {
		t.Fatalf("un curso solo debe revisarse una vez: %v", err)
	}
	if _, err := servicio.RevisarCurso(ctx, importada.Cursos[1].ID, "coordinacion-1", EstadoRechazado, "No pertenece al grado"); err != nil {
		t.Fatal(err)
	}
	publica, err = servicio.CatalogoPublico(ctx)
	if err != nil || len(publica.Cursos) != 1 || publica.Cursos[0].ID != aprobado.ID {
		t.Fatalf("un curso rechazado tampoco debe ser público: %+v, error: %v", publica.Cursos, err)
	}

	if _, err := servicio.Importar(ctx, datosValidosCatalogo()); err != nil {
		t.Fatal(err)
	}
	publica, err = servicio.CatalogoPublico(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(publica.Cursos) != 0 {
		t.Fatalf("la última importación pendiente no debe heredar aprobaciones previas: %+v", publica.Cursos)
	}
}

func TestSolicitudColegioDuplicadaYResueltaAlImportar(t *testing.T) {
	ctx := context.Background()
	repositorio := memoria.NuevoRepositorioMemoria()
	servicio := NuevoServicio(repositorio)
	servicio.ahora = func() time.Time { return time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC) }
	if _, err := servicio.Importar(ctx, datosValidosCatalogo()); err != nil {
		t.Fatal(err)
	}

	entrada := SolicitudColegioEntrada{Nombre: " Colegio Nuevo ", Distrito: "Distrito Nuevo"}
	if _, err := servicio.SolicitarColegio(ctx, "escolar-1", entrada); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.SolicitarColegio(ctx, "apoderado-1", entrada); !errors.Is(err, ErrSolicitudDuplicada) {
		t.Fatalf("no se deben duplicar solicitudes pendientes: %v", err)
	}
	pendientes, err := servicio.SolicitudesPendientes(ctx)
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("solicitudes pendientes: %d, error: %v", len(pendientes), err)
	}

	datos := datosValidosCatalogo()
	datos.Colegios = append(datos.Colegios, ColegioEntrada{Nombre: "Colegio Nuevo", Distrito: "Distrito Nuevo"})
	if _, err := servicio.Importar(ctx, datos); err != nil {
		t.Fatal(err)
	}
	pendientes, err = servicio.SolicitudesPendientes(ctx)
	if err != nil || len(pendientes) != 0 {
		t.Fatalf("una solicitud incluida en el nuevo padrón debe resolverse: %d, error: %v", len(pendientes), err)
	}
	if _, err := servicio.SolicitarColegio(ctx, "escolar-1", SolicitudColegioEntrada{
		Nombre: "Colegio Sintético", Distrito: "Distrito Sintético",
	}); !errors.Is(err, ErrColegioYaExiste) {
		t.Fatalf("no se deben solicitar colegios ya importados: %v", err)
	}
}

func TestDatosInvalidosYRechazoRequiereMotivo(t *testing.T) {
	ahora := time.Date(2026, 10, 22, 12, 0, 0, 0, time.UTC)
	datos := datosValidosCatalogo()
	datos.Fuente = "  "
	if _, err := NuevaImportacion(datos, ahora); err == nil {
		t.Fatal("debe requerirse una fuente")
	}
	datos = datosValidosCatalogo()
	datos.Cursos[0].Grado = "1.\nsecundaria"
	if _, err := NuevaImportacion(datos, ahora); err == nil {
		t.Fatal("el grado no debe contener caracteres de control")
	}
	if _, err := NuevaSolicitudColegio(SolicitudColegioEntrada{Distrito: "Lima"}, ahora); err == nil {
		t.Fatal("debe requerirse el nombre del colegio")
	}

	importacion, err := NuevaImportacion(datosValidosCatalogo(), ahora)
	if err != nil {
		t.Fatal(err)
	}
	curso := &importacion.Cursos[0]
	curso.ID = 1
	if err := curso.Revisar(EstadoRechazado, "coord-1", "  ", ahora); err == nil {
		t.Fatal("un rechazo debe exigir motivo")
	}
	if err := curso.Revisar(EstadoAprobado, "", "", ahora); err == nil {
		t.Fatal("debe requerirse el revisor")
	}
	if err := curso.Revisar(EstadoAprobado, "coord-1", "", ahora); err != nil {
		t.Fatal(err)
	}
	if err := curso.Revisar(EstadoRechazado, "coord-1", "duplicado", ahora); !errors.Is(err, ErrEstadoInvalido) {
		t.Fatalf("no se debe revisar dos veces: %v", err)
	}
}

func datosValidosCatalogo() DatosImportacion {
	return DatosImportacion{
		Fuente: "sintetico: prueba",
		Cursos: []CursoEntrada{
			{Nombre: "Álgebra", Grado: "1.° de secundaria"},
			{Nombre: "Comunicación", Grado: "2.° de secundaria"},
		},
		Colegios: []ColegioEntrada{
			{Nombre: "Colegio Sintético", Distrito: "Distrito Sintético"},
		},
	}
}
