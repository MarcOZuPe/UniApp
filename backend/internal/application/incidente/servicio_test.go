package incidente

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/incidente"
	modeloAcceso "uniapp/internal/application/acceso"
)

type almacenMemoria struct {
	archivos map[string][]byte
}

func (a *almacenMemoria) Guardar(_ context.Context, clave string, contenido []byte) error {
	a.archivos[clave] = append([]byte(nil), contenido...)
	return nil
}

func (a *almacenMemoria) Leer(_ context.Context, clave string) ([]byte, error) {
	contenido, existe := a.archivos[clave]
	if !existe {
		return nil, ErrAlmacen
	}
	return append([]byte(nil), contenido...), nil
}

func (a *almacenMemoria) Eliminar(_ context.Context, clave string) error {
	delete(a.archivos, clave)
	return nil
}

func entornoIncidente(t *testing.T) (*Servicio, *memoria.RepositorioMemoria, *almacenMemoria, time.Time) {
	t.Helper()
	repositorio := memoria.NuevoRepositorioMemoria()
	almacen := &almacenMemoria{archivos: map[string][]byte{}}
	servicio := NuevoServicio(repositorio, almacen)
	ahora := time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC)
	servicio.FijarReloj(func() time.Time { return ahora })
	repositorio.SembrarTutoria(Tutoria{
		ReservaID: "reserva-1", EscolarID: "escolar-1", EscolarNombre: "Escolar Uno",
		TutorID: "tutor-1", TutorNombre: "Tutor Uno", Curso: "Álgebra",
		Inicio: ahora.Add(-2 * time.Hour), Fin: ahora.Add(-time.Hour), Asistio: true,
	})
	repositorio.SembrarVinculacion("apoderado-1", EscolarVinculado{ID: "escolar-1", Nombre: "Escolar Uno"})
	return servicio, repositorio, almacen, ahora
}

func datosIncidenteValido() DatosIncidente {
	return DatosIncidente{
		ReservaID: "reserva-1", Tipo: TipoTratoInadecuado,
		Descripcion: strings.Repeat("Descripción del incidente ", 2),
		Evidencias: []Archivo{
			{Nombre: "chat.pdf", Contenido: []byte("%PDF-1.4\n% evidencia\n")},
			{Nombre: "captura.png", Contenido: append([]byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, []byte("datos")...)},
		},
	}
}

func TestRegistroYVisibilidadDeIncidentesPorRol(t *testing.T) {
	servicio, repositorio, almacen, ahora := entornoIncidente(t)
	ctx := context.Background()

	tutorias, err := servicio.Tutorias(ctx, "escolar-1", modeloAcceso.RolEscolar)
	if err != nil || len(tutorias) != 1 {
		t.Fatalf("tutorías del escolar: %+v, error: %v", tutorias, err)
	}
	tutoriasApoderado, err := servicio.Tutorias(ctx, "apoderado-1", modeloAcceso.RolApoderado)
	if err != nil || len(tutoriasApoderado) != 1 || tutoriasApoderado[0].EscolarNombre != "Escolar Uno" {
		t.Fatalf("tutorías del apoderado: %+v, error: %v", tutoriasApoderado, err)
	}

	creado, err := servicio.Registrar(ctx, "escolar-1", modeloAcceso.RolEscolar, datosIncidenteValido())
	if err != nil {
		t.Fatal(err)
	}
	if creado.Estado != EstadoPendiente || len(creado.Evidencias) != 2 || len(almacen.archivos) != 2 {
		t.Fatalf("incidente creado: %+v, archivos: %d", creado, len(almacen.archivos))
	}

	if _, err := servicio.Listar(ctx, "tutor-1", modeloAcceso.RolTutor, ""); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("el tutor no debe listar: %v", err)
	}
	if _, err := servicio.PorID(ctx, "tutor-1", modeloAcceso.RolTutor, "", creado.ID); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("el tutor no debe ver el detalle: %v", err)
	}
	if _, _, _, err := servicio.AbrirEvidencia(ctx, "tutor-1", modeloAcceso.RolTutor, "", creado.ID, creado.Evidencias[0].ID); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("el tutor no debe descargar evidencia: %v", err)
	}

	propios, err := servicio.Listar(ctx, "escolar-1", modeloAcceso.RolEscolar, "")
	if err != nil || len(propios) != 1 {
		t.Fatalf("listado del escolar: %+v, error: %v", propios, err)
	}
	delApoderado, err := servicio.Listar(ctx, "apoderado-1", modeloAcceso.RolApoderado, "")
	if err != nil || len(delApoderado) != 1 {
		t.Fatalf("listado del apoderado: %+v, error: %v", delApoderado, err)
	}
	ajenos, err := servicio.Listar(ctx, "escolar-2", modeloAcceso.RolEscolar, "")
	if err != nil || len(ajenos) != 0 {
		t.Fatalf("otro escolar no debe ver incidentes ajenos: %+v, error: %v", ajenos, err)
	}
	if _, err := servicio.PorID(ctx, "escolar-2", modeloAcceso.RolEscolar, "", creado.ID); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("otro escolar no debe ver el detalle: %v", err)
	}

	if _, evidencia, contenido, err := servicio.AbrirEvidencia(ctx, "apoderado-1", modeloAcceso.RolApoderado, "", creado.ID, creado.Evidencias[0].ID); err != nil || evidencia.Nombre != "chat.pdf" || len(contenido) == 0 {
		t.Fatalf("descarga del apoderado: %+v, error: %v", evidencia, err)
	}
	if _, _, _, err := servicio.AbrirEvidencia(ctx, "coord-1", modeloAcceso.RolCoordinador, "1", creado.ID, creado.Evidencias[1].ID); err != nil {
		t.Fatalf("descarga de coordinación: %v", err)
	}

	if _, err := servicio.Revisar(ctx, "escolar-1", modeloAcceso.RolEscolar, "", creado.ID, DatosRevision{
		Resultado: ResultadoProcede, Comentario: "Intento de revisión",
	}); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("el escolar no debe revisar: %v", err)
	}
	revisado, err := servicio.Revisar(ctx, "coord-1", modeloAcceso.RolCoordinador, "1", creado.ID, DatosRevision{
		Resultado: ResultadoProcede, Comentario: "La evidencia sustenta el reporte.",
	})
	if err != nil {
		t.Fatal(err)
	}
	if revisado.Estado != EstadoRevisado || revisado.Resultado != ResultadoProcede || revisado.RevisadoPor != "coord-1" || revisado.RevisadoEn == nil {
		t.Fatalf("revisión: %+v", revisado)
	}
	if _, err := servicio.Revisar(ctx, "coord-1", modeloAcceso.RolCoordinador, "1", creado.ID, DatosRevision{
		Resultado: ResultadoNoProcede, Comentario: "Segunda revisión.",
	}); !errors.Is(err, ErrYaRevisado) {
		t.Fatalf("segunda revisión: %v", err)
	}

	// La tutoría de la semilla sigue siendo visible después de la revisión.
	tutorias, _ = repositorio.TutoriasDeEscolar(ctx, "escolar-1", ahora)
	if len(tutorias) != 1 {
		t.Fatalf("tutorías persistentes: %+v", tutorias)
	}
}

func TestApoderadoSoloReportaParaEscolaresVinculados(t *testing.T) {
	servicio, _, _, _ := entornoIncidente(t)
	ctx := context.Background()
	datos := datosIncidenteValido()
	datos.EscolarID = "escolar-2"
	if _, err := servicio.Registrar(ctx, "apoderado-1", modeloAcceso.RolApoderado, datos); !errors.Is(err, ErrEscolarNoVinculado) {
		t.Fatalf("escolar no vinculado: %v", err)
	}
	datos.EscolarID = "escolar-1"
	creado, err := servicio.Registrar(ctx, "apoderado-1", modeloAcceso.RolApoderado, datos)
	if err != nil || creado.ReportadoPor != "apoderado-1" || creado.EscolarID != "escolar-1" {
		t.Fatalf("reporte del apoderado: %+v, error: %v", creado, err)
	}
}

func TestNoSePuedeReportarSobreTutoriaAjenaOFutura(t *testing.T) {
	servicio, repositorio, _, ahora := entornoIncidente(t)
	ctx := context.Background()
	datos := datosIncidenteValido()
	datos.ReservaID = "reserva-ajena"
	if _, err := servicio.Registrar(ctx, "escolar-1", modeloAcceso.RolEscolar, datos); !errors.Is(err, ErrTutoriaNoEncontrada) {
		t.Fatalf("tutoría ajena: %v", err)
	}
	repositorio.SembrarTutoria(Tutoria{
		ReservaID: "reserva-futura", EscolarID: "escolar-1", EscolarNombre: "Escolar Uno",
		TutorID: "tutor-1", TutorNombre: "Tutor Uno", Curso: "Álgebra",
		Inicio: ahora.Add(time.Hour), Fin: ahora.Add(2 * time.Hour),
	})
	datos.ReservaID = "reserva-futura"
	if _, err := servicio.Registrar(ctx, "escolar-1", modeloAcceso.RolEscolar, datos); !errors.Is(err, ErrTutoriaNoEncontrada) {
		t.Fatalf("tutoría futura: %v", err)
	}
}

func TestCoordinadorNoAccedeAIncidentesDeOtraFacultad(t *testing.T) {
	servicio, _, _, _ := entornoIncidente(t)
	ctx := context.Background()
	creado, err := servicio.Registrar(ctx, "escolar-1", modeloAcceso.RolEscolar, datosIncidenteValido())
	if err != nil {
		t.Fatal(err)
	}
	if lista, err := servicio.Listar(ctx, "coord-2", modeloAcceso.RolCoordinador, "2"); err != nil || len(lista) != 0 {
		t.Fatalf("no debe listar incidentes de otra facultad: %+v, %v", lista, err)
	}
	if _, err := servicio.PorID(ctx, "coord-2", modeloAcceso.RolCoordinador, "2", creado.ID); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("no debe ver el detalle de otra facultad: %v", err)
	}
	if _, err := servicio.Revisar(ctx, "coord-2", modeloAcceso.RolCoordinador, "2", creado.ID, DatosRevision{
		Resultado: ResultadoProcede, Comentario: "Intento ajeno.",
	}); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("no debe revisar incidentes de otra facultad: %v", err)
	}
}
