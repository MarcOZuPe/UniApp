package incidente

import (
	"context"
	"testing"
	"time"

	modeloAcceso "uniapp/internal/application/acceso"
)

type repositorioPrueba struct {
	escolares         []EscolarVinculado
	tutorias          map[string][]Tutoria
	consultasTutorias int
}

func (r *repositorioPrueba) TutoriasDeEscolar(_ context.Context, escolarID string, _ time.Time) ([]Tutoria, error) {
	r.consultasTutorias++
	return r.tutorias[escolarID], nil
}

func (r *repositorioPrueba) TutoriaDeEscolar(context.Context, string, string, time.Time) (*Tutoria, error) {
	return nil, ErrTutoriaNoEncontrada
}

func (r *repositorioPrueba) EscolaresDeApoderado(context.Context, string) ([]EscolarVinculado, error) {
	return r.escolares, nil
}

func (r *repositorioPrueba) EsApoderadoDe(_ context.Context, _, escolarID string) (bool, error) {
	for _, escolar := range r.escolares {
		if escolar.ID == escolarID {
			return true, nil
		}
	}
	return false, nil
}

func (r *repositorioPrueba) Crear(context.Context, *Incidente) (*Incidente, error) {
	return nil, nil
}

func (r *repositorioPrueba) PorID(context.Context, string) (*Incidente, error) {
	return nil, ErrIncidenteNoEncontrado
}

func (r *repositorioPrueba) Listar(context.Context, Filtro) ([]Incidente, error) {
	return nil, nil
}

func (r *repositorioPrueba) Revisar(context.Context, string, string, Resultado, string, time.Time) (*Incidente, error) {
	return nil, nil
}

func (r *repositorioPrueba) EvidenciaDe(context.Context, string, string) (Evidencia, error) {
	return Evidencia{}, ErrEvidenciaNoEncontrada
}

func TestNuevoAlcanceRechazaRolesSinAccesoAIncidentes(t *testing.T) {
	for _, rol := range []modeloAcceso.Rol{
		modeloAcceso.RolTutor,
		modeloAcceso.RolAdministrador,
		modeloAcceso.Rol("desconocido"),
	} {
		if _, err := nuevoAlcance(Solicitante{ID: "u-1", Rol: rol}, &repositorioPrueba{}, time.Now()); err == nil {
			t.Fatalf("el rol %q no debería tener alcance de incidentes", rol)
		}
	}
}

func TestAlcanceEscolarSoloVeIncidentesPropios(t *testing.T) {
	alcance, err := nuevoAlcance(
		Solicitante{ID: "escolar-1", Rol: modeloAcceso.RolEscolar},
		&repositorioPrueba{},
		time.Now(),
	)
	if err != nil {
		t.Fatalf("no se pudo crear el alcance escolar: %v", err)
	}

	visible, err := alcance.Visible(context.Background(), &Incidente{EscolarID: "escolar-1"})
	if err != nil || !visible {
		t.Fatalf("el escolar debería ver su propio incidente: visible=%v err=%v", visible, err)
	}

	visible, err = alcance.Visible(context.Background(), &Incidente{EscolarID: "escolar-2"})
	if err != nil || visible {
		t.Fatalf("el escolar no debería ver incidentes ajenos: visible=%v err=%v", visible, err)
	}
}

func TestAlcanceApoderadoFiltraSusEscolaresVinculados(t *testing.T) {
	repositorio := &repositorioPrueba{escolares: []EscolarVinculado{{ID: "escolar-1"}, {ID: "escolar-2"}}}
	alcance, err := nuevoAlcance(
		Solicitante{ID: "apoderado-1", Rol: modeloAcceso.RolApoderado},
		repositorio,
		time.Now(),
	)
	if err != nil {
		t.Fatalf("no se pudo crear el alcance del apoderado: %v", err)
	}

	filtro, aplicable, err := alcance.Filtrar(context.Background())
	if err != nil || !aplicable {
		t.Fatalf("el apoderado con escolares debería tener filtro: aplicable=%v err=%v", aplicable, err)
	}
	if len(filtro.EscolaresID) != 2 {
		t.Fatalf("se esperaban 2 escolares vinculados, se obtuvo %d", len(filtro.EscolaresID))
	}

	visible, err := alcance.Visible(context.Background(), &Incidente{EscolarID: "escolar-2"})
	if err != nil || !visible {
		t.Fatalf("el apoderado debería ver incidentes de sus escolares: visible=%v err=%v", visible, err)
	}

	visible, err = alcance.Visible(context.Background(), &Incidente{EscolarID: "escolar-3"})
	if err != nil || visible {
		t.Fatalf("el apoderado no debería ver incidentes ajenos: visible=%v err=%v", visible, err)
	}
}

func TestAlcanceApoderadoSinEscolaresNoAplicaFiltro(t *testing.T) {
	alcance, err := nuevoAlcance(
		Solicitante{ID: "apoderado-1", Rol: modeloAcceso.RolApoderado},
		&repositorioPrueba{},
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, aplicable, err := alcance.Filtrar(context.Background()); err != nil || aplicable {
		t.Fatalf("sin escolares no debería consultarse la lista: aplicable=%v err=%v", aplicable, err)
	}
}

func TestAlcanceApoderadoAgregaTutoriasDeSusEscolares(t *testing.T) {
	repositorio := &repositorioPrueba{
		escolares: []EscolarVinculado{{ID: "escolar-1"}, {ID: "escolar-2"}},
		tutorias: map[string][]Tutoria{
			"escolar-1": {{ReservaID: "r-1"}},
			"escolar-2": {{ReservaID: "r-2"}, {ReservaID: "r-3"}},
		},
	}
	alcance, err := nuevoAlcance(
		Solicitante{ID: "apoderado-1", Rol: modeloAcceso.RolApoderado},
		repositorio,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}
	tutorias, err := alcance.Tutorias(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tutorias) != 3 {
		t.Fatalf("se esperaban 3 tutorías, se obtuvieron %d", len(tutorias))
	}
}

// F-17: el apoderado con varios escolares no debe disparar una consulta de
// tutorías por cada fila. Debe resolverse con una sola consulta por lote.
func TestF17AlcanceApoderadoNoDebeConsultarTutoriasPorFila(t *testing.T) {
	repositorio := &repositorioPrueba{
		escolares: []EscolarVinculado{
			{ID: "escolar-1"},
			{ID: "escolar-2"},
			{ID: "escolar-3"},
		},
		tutorias: map[string][]Tutoria{
			"escolar-1": {{ReservaID: "r-1"}},
			"escolar-2": {{ReservaID: "r-2"}},
			"escolar-3": {{ReservaID: "r-3"}},
		},
	}
	alcance, err := nuevoAlcance(
		Solicitante{ID: "apoderado-1", Rol: modeloAcceso.RolApoderado},
		repositorio,
		time.Now(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := alcance.Tutorias(context.Background()); err != nil {
		t.Fatal(err)
	}
	if repositorio.consultasTutorias != 1 {
		t.Fatalf("F-17: se consultaron las tutorías %d veces para 3 escolares; debe ser una consulta por lote",
			repositorio.consultasTutorias)
	}
}

func TestAlcanceCoordinacionSeLimitaASuFacultad(t *testing.T) {
	sinFacultad, err := nuevoAlcance(
		Solicitante{ID: "coord-1", Rol: modeloAcceso.RolCoordinador},
		&repositorioPrueba{},
		time.Now(),
	)
	if err != nil {
		t.Fatalf("no se pudo crear el alcance de coordinación: %v", err)
	}
	if _, aplicable, _ := sinFacultad.Filtrar(context.Background()); aplicable {
		t.Fatal("coordinación sin facultad no debería consultar incidentes")
	}

	conFacultad, err := nuevoAlcance(
		Solicitante{ID: "coord-1", Rol: modeloAcceso.RolCoordinador, FacultadID: "fac-1"},
		&repositorioPrueba{},
		time.Now(),
	)
	if err != nil {
		t.Fatalf("no se pudo crear el alcance de coordinación: %v", err)
	}
	filtro, aplicable, err := conFacultad.Filtrar(context.Background())
	if err != nil || !aplicable || filtro.FacultadID != "fac-1" {
		t.Fatalf("el filtro debería usar la facultad del coordinador: %+v aplicable=%v err=%v", filtro, aplicable, err)
	}

	visible, err := conFacultad.Visible(context.Background(), &Incidente{FacultadID: "fac-1"})
	if err != nil || !visible {
		t.Fatalf("el coordinador debería ver incidentes de su facultad: visible=%v err=%v", visible, err)
	}
	visible, err = conFacultad.Visible(context.Background(), &Incidente{FacultadID: "fac-2"})
	if err != nil || visible {
		t.Fatalf("el coordinador no debería ver incidentes de otra facultad: visible=%v err=%v", visible, err)
	}
}
