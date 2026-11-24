package incidente

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type RepositorioMemoria struct {
	mu         sync.RWMutex
	secuencia  int
	tutorias   []Tutoria
	escolares  map[string][]EscolarVinculado
	incidentes map[string]*Incidente
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		escolares:  map[string][]EscolarVinculado{},
		incidentes: map[string]*Incidente{},
	}
}

// SembrarTutoria permite a pruebas y demostraciones preparar el historial del escolar
// sin depender del flujo de reservas.
func (r *RepositorioMemoria) SembrarTutoria(tutoria Tutoria) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tutorias = append(r.tutorias, tutoria)
}

func (r *RepositorioMemoria) SembrarVinculacion(apoderadoID string, escolar EscolarVinculado) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.escolares[apoderadoID] = append(r.escolares[apoderadoID], escolar)
}

func (r *RepositorioMemoria) TutoriasDeEscolar(_ context.Context, escolarID string, ahora time.Time) ([]Tutoria, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := make([]Tutoria, 0)
	for _, tutoria := range r.tutorias {
		if tutoria.EscolarID == escolarID && !tutoria.Fin.After(ahora) {
			resultado = append(resultado, tutoria)
		}
	}
	sort.Slice(resultado, func(i, j int) bool {
		if resultado[i].Inicio.Equal(resultado[j].Inicio) {
			return resultado[i].ReservaID < resultado[j].ReservaID
		}
		return resultado[i].Inicio.After(resultado[j].Inicio)
	})
	return resultado, nil
}

func (r *RepositorioMemoria) TutoriaDeEscolar(_ context.Context, escolarID, reservaID string, ahora time.Time) (*Tutoria, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, tutoria := range r.tutorias {
		if tutoria.EscolarID == escolarID && tutoria.ReservaID == reservaID && !tutoria.Fin.After(ahora) {
			copia := tutoria
			return &copia, nil
		}
	}
	return nil, ErrTutoriaNoEncontrada
}

func (r *RepositorioMemoria) EscolaresDeApoderado(_ context.Context, apoderadoID string) ([]EscolarVinculado, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]EscolarVinculado(nil), r.escolares[apoderadoID]...), nil
}

func (r *RepositorioMemoria) EsApoderadoDe(_ context.Context, apoderadoID, escolarID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, escolar := range r.escolares[apoderadoID] {
		if escolar.ID == escolarID {
			return true, nil
		}
	}
	return false, nil
}

func (r *RepositorioMemoria) Crear(_ context.Context, incidente *Incidente) (*Incidente, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.secuencia++
	copia := copiarIncidente(incidente)
	copia.ID = fmt.Sprintf("incidente-%d", r.secuencia)
	if copia.FacultadID == "" {
		copia.FacultadID = "1"
	}
	if copia.CreadoEn.IsZero() {
		copia.CreadoEn = time.Now().UTC()
	}
	for indice := range copia.Evidencias {
		copia.Evidencias[indice].ID = fmt.Sprintf("evidencia-%d-%d", r.secuencia, indice+1)
		if copia.Evidencias[indice].CreadoEn.IsZero() {
			copia.Evidencias[indice].CreadoEn = copia.CreadoEn
		}
	}
	r.incidentes[copia.ID] = copia
	return copiarIncidente(copia), nil
}

func (r *RepositorioMemoria) PorID(_ context.Context, incidenteID string) (*Incidente, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	incidente, existe := r.incidentes[incidenteID]
	if !existe {
		return nil, ErrIncidenteNoEncontrado
	}
	return copiarIncidente(incidente), nil
}

func (r *RepositorioMemoria) Listar(_ context.Context, filtro Filtro) ([]Incidente, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := make([]Incidente, 0)
	for _, incidente := range r.incidentes {
		if !coincideFiltro(incidente, filtro) {
			continue
		}
		resultado = append(resultado, *copiarIncidente(incidente))
	}
	sort.Slice(resultado, func(i, j int) bool {
		if resultado[i].CreadoEn.Equal(resultado[j].CreadoEn) {
			return resultado[i].ID > resultado[j].ID
		}
		return resultado[i].CreadoEn.After(resultado[j].CreadoEn)
	})
	return resultado, nil
}

func (r *RepositorioMemoria) Revisar(
	_ context.Context,
	incidenteID, revisorID string,
	resultado Resultado,
	comentario string,
	ahora time.Time,
) (*Incidente, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	incidente, existe := r.incidentes[incidenteID]
	if !existe {
		return nil, ErrIncidenteNoEncontrado
	}
	if incidente.Estado != EstadoPendiente {
		return nil, ErrYaRevisado
	}
	revisadoEn := ahora.UTC()
	incidente.Estado = EstadoRevisado
	incidente.Resultado = resultado
	incidente.ComentarioRevision = comentario
	incidente.RevisadoPor = revisorID
	incidente.RevisadoEn = &revisadoEn
	return copiarIncidente(incidente), nil
}

func (r *RepositorioMemoria) EvidenciaDe(_ context.Context, incidenteID, evidenciaID string) (Evidencia, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	incidente, existe := r.incidentes[incidenteID]
	if !existe {
		return Evidencia{}, ErrIncidenteNoEncontrado
	}
	for _, evidencia := range incidente.Evidencias {
		if evidencia.ID == evidenciaID {
			return evidencia, nil
		}
	}
	return Evidencia{}, ErrEvidenciaNoEncontrada
}

func coincideFiltro(incidente *Incidente, filtro Filtro) bool {
	if filtro.FacultadID != "" {
		return incidente.FacultadID == filtro.FacultadID
	}
	if filtro.Todos {
		return true
	}
	if filtro.EscolarID != "" {
		return incidente.EscolarID == filtro.EscolarID
	}
	for _, escolarID := range filtro.EscolaresID {
		if incidente.EscolarID == escolarID {
			return true
		}
	}
	return false
}

func copiarIncidente(incidente *Incidente) *Incidente {
	if incidente == nil {
		return nil
	}
	copia := *incidente
	copia.Evidencias = append([]Evidencia(nil), incidente.Evidencias...)
	if incidente.RevisadoEn != nil {
		revisadoEn := *incidente.RevisadoEn
		copia.RevisadoEn = &revisadoEn
	}
	return &copia
}
