package incidente

import (
	"context"
	"testing"
	"time"

	modeloAcceso "uniapp/internal/application/acceso"
)

var evidenciaPDF = []byte("%PDF-1.4\n%evidencia\n")

type almacenPrueba struct{ datos map[string][]byte }

func (a *almacenPrueba) Guardar(_ context.Context, clave string, contenido []byte) error {
	if a.datos == nil {
		a.datos = map[string][]byte{}
	}
	a.datos[clave] = append([]byte(nil), contenido...)
	return nil
}
func (a *almacenPrueba) Leer(_ context.Context, clave string) ([]byte, error) {
	contenido, existe := a.datos[clave]
	if !existe {
		return nil, ErrAlmacen
	}
	return contenido, nil
}
func (a *almacenPrueba) Eliminar(_ context.Context, clave string) error {
	delete(a.datos, clave)
	return nil
}

func TestServicioIncidenteRegistrarListarYRevisar(t *testing.T) {
	ahora := time.Now().UTC()
	inicio := ahora.Add(-2 * time.Hour)
	repo := &repositorioPrueba{
		tutoria: &Tutoria{
			ReservaID: "r-1", EscolarID: "e-1", EscolarNombre: "Escolar", TutorID: "t-1",
			TutorNombre: "Tutor", TutorFacultadID: "fac-1", Curso: "Álgebra",
			Inicio: inicio, Fin: inicio.Add(time.Hour),
		},
		escolares: []EscolarVinculado{{ID: "e-1", Nombre: "Escolar"}},
	}
	almacen := &almacenPrueba{}
	servicio := NuevoServicio(repo, almacen)
	servicio.FijarReloj(func() time.Time { return ahora })
	datos := DatosIncidente{
		EscolarID: "e-1", ReservaID: "r-1", Tipo: TipoTratoInadecuado,
		Descripcion: "Descripción suficientemente larga", Evidencias: []Archivo{{Nombre: "a.pdf", Contenido: evidenciaPDF}},
	}

	creado, err := servicio.Registrar(context.Background(), Solicitante{ID: "e-1", Rol: modeloAcceso.RolEscolar}, datos)
	if err != nil {
		t.Fatalf("registrar falló: %v", err)
	}
	if len(creado.Evidencias) != 1 || creado.Evidencias[0].ClaveAlmacenamiento == "" || len(almacen.datos) != 1 {
		t.Fatalf("evidencia no almacenada: %+v", creado)
	}
	if _, err := servicio.Registrar(context.Background(), Solicitante{ID: "e-1", Rol: modeloAcceso.RolTutor}, datos); err != ErrNoAutorizado {
		t.Fatalf("un tutor no debe registrar incidentes: %v", err)
	}
	if _, err := servicio.Registrar(context.Background(), Solicitante{ID: "a-1", Rol: modeloAcceso.RolApoderado}, datos); err != nil {
		t.Fatal(err)
	}

	incidente := &Incidente{
		ID: "i-1", EscolarID: "e-1", FacultadID: "fac-1", Estado: EstadoPendiente,
		Evidencias: []Evidencia{{ID: "ev-1", ClaveAlmacenamiento: creado.Evidencias[0].ClaveAlmacenamiento}},
	}
	repo.incidente = incidente
	repo.evidencia = incidente.Evidencias[0]
	listado, err := servicio.Listar(context.Background(), Solicitante{ID: "e-1", Rol: modeloAcceso.RolEscolar})
	if err != nil || len(listado) != 1 || repo.filtro.EscolarID != "e-1" {
		t.Fatalf("listar escolar inesperado: %+v filtro=%+v err=%v", listado, repo.filtro, err)
	}
	if _, err := servicio.Listar(context.Background(), Solicitante{ID: "a-1", Rol: modeloAcceso.RolApoderado}); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.Listar(context.Background(), Solicitante{ID: "c-1", Rol: modeloAcceso.RolCoordinador, FacultadID: "fac-1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := servicio.PorID(context.Background(), Solicitante{ID: "e-2", Rol: modeloAcceso.RolEscolar}, "i-1"); err != ErrNoAutorizado {
		t.Fatalf("un escolar ajeno no debe ver el incidente: %v", err)
	}
	_, evidencia, contenido, err := servicio.AbrirEvidencia(context.Background(), Solicitante{ID: "e-1", Rol: modeloAcceso.RolEscolar}, "i-1", "ev-1")
	if err != nil || evidencia.ID != "ev-1" || string(contenido) != string(evidenciaPDF) {
		t.Fatalf("abrir evidencia inesperado: %+v %q err=%v", evidencia, contenido, err)
	}
	revisado, err := servicio.Revisar(context.Background(), Solicitante{ID: "c-1", Rol: modeloAcceso.RolCoordinador, FacultadID: "fac-1"}, "i-1", DatosRevision{
		Resultado: ResultadoProcede, Comentario: "Comentario válido",
	})
	if err != nil || revisado.Estado != EstadoRevisado {
		t.Fatalf("revisar inesperado: %+v err=%v", revisado, err)
	}
	if _, err := servicio.Revisar(context.Background(), Solicitante{ID: "e-1", Rol: modeloAcceso.RolEscolar}, "i-1", DatosRevision{}); err != ErrNoAutorizado {
		t.Fatalf("un escolar no debe revisar: %v", err)
	}
}

func TestServicioIncidenteSinAlmacen(t *testing.T) {
	repo := &repositorioPrueba{}
	servicio := NuevoServicio(repo, nil)
	servicio.FijarReloj(time.Now)
	if _, err := servicio.Registrar(context.Background(), Solicitante{ID: "e-1", Rol: modeloAcceso.RolEscolar}, DatosIncidente{}); err == nil {
		t.Fatal("sin almacén no debe registrarse")
	}
}
