package incidente

import (
	"context"
	"strings"
	"time"

	modeloAcceso "uniapp/internal/application/acceso"
)

// Solicitante agrupa la identidad y el alcance del usuario que consulta o
// registra un incidente. Evita repetir la terna ID, rol y facultad en cada caso
// de uso y mantiene las firmas del servicio pequeñas.
type Solicitante struct {
	ID         string
	Rol        modeloAcceso.Rol
	FacultadID string
}

func (solicitante Solicitante) identificado() bool {
	return strings.TrimSpace(solicitante.ID) != ""
}

// alcance aplica la Strategy de visibilidad correspondiente al rol. Cada rol
// decide qué tutorías e incidentes puede consultar, sin condicionales repetidos
// dentro del servicio.
type alcance interface {
	Tutorias(ctx context.Context) ([]Tutoria, error)
	Filtrar(ctx context.Context) (Filtro, bool, error)
	Visible(ctx context.Context, incidente *Incidente) (bool, error)
}

type alcanceBase struct {
	repositorio Repositorio
	solicitante Solicitante
	ahora       time.Time
}

func nuevoAlcance(solicitante Solicitante, repositorio Repositorio, ahora time.Time) (alcance, error) {
	base := alcanceBase{repositorio: repositorio, solicitante: solicitante, ahora: ahora}
	switch solicitante.Rol {
	case modeloAcceso.RolEscolar:
		return alcanceEscolar{base}, nil
	case modeloAcceso.RolApoderado:
		return alcanceApoderado{base}, nil
	case modeloAcceso.RolCoordinador, modeloAcceso.RolCoordinadorGeneral:
		return alcanceCoordinacion{base}, nil
	default:
		return nil, ErrNoAutorizado
	}
}

type alcanceEscolar struct{ alcanceBase }

func (alcance alcanceEscolar) Tutorias(ctx context.Context) ([]Tutoria, error) {
	return alcance.repositorio.TutoriasDeEscolar(ctx, alcance.solicitante.ID, alcance.ahora)
}

func (alcance alcanceEscolar) Filtrar(context.Context) (Filtro, bool, error) {
	return Filtro{EscolarID: alcance.solicitante.ID}, true, nil
}

func (alcance alcanceEscolar) Visible(_ context.Context, incidente *Incidente) (bool, error) {
	return incidente.EscolarID == alcance.solicitante.ID, nil
}

type alcanceApoderado struct{ alcanceBase }

func (alcance alcanceApoderado) Tutorias(ctx context.Context) ([]Tutoria, error) {
	escolares, err := alcance.escolares(ctx)
	if err != nil {
		return nil, err
	}
	tutorias := make([]Tutoria, 0)
	for _, escolar := range escolares {
		delEscolar, err := alcance.repositorio.TutoriasDeEscolar(ctx, escolar.ID, alcance.ahora)
		if err != nil {
			return nil, err
		}
		tutorias = append(tutorias, delEscolar...)
	}
	return tutorias, nil
}

func (alcance alcanceApoderado) Filtrar(ctx context.Context) (Filtro, bool, error) {
	escolares, err := alcance.escolares(ctx)
	if err != nil {
		return Filtro{}, false, err
	}
	if len(escolares) == 0 {
		return Filtro{}, false, nil
	}
	ids := make([]string, 0, len(escolares))
	for _, escolar := range escolares {
		ids = append(ids, escolar.ID)
	}
	return Filtro{EscolaresID: ids}, true, nil
}

func (alcance alcanceApoderado) Visible(ctx context.Context, incidente *Incidente) (bool, error) {
	return alcance.repositorio.EsApoderadoDe(ctx, alcance.solicitante.ID, incidente.EscolarID)
}

func (alcance alcanceApoderado) escolares(ctx context.Context) ([]EscolarVinculado, error) {
	return alcance.repositorio.EscolaresDeApoderado(ctx, alcance.solicitante.ID)
}

type alcanceCoordinacion struct{ alcanceBase }

func (alcanceCoordinacion) Tutorias(context.Context) ([]Tutoria, error) {
	return []Tutoria{}, nil
}

func (alcance alcanceCoordinacion) Filtrar(context.Context) (Filtro, bool, error) {
	if alcance.solicitante.FacultadID == "" {
		return Filtro{}, false, nil
	}
	return Filtro{FacultadID: alcance.solicitante.FacultadID}, true, nil
}

func (alcance alcanceCoordinacion) Visible(_ context.Context, incidente *Incidente) (bool, error) {
	return alcance.solicitante.FacultadID != "" &&
		incidente.FacultadID == alcance.solicitante.FacultadID, nil
}
