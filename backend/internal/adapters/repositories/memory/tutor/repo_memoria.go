package tutor

import (
	"context"
	"sort"
	"sync"
	"time"
)

type RepositorioMemoria struct {
	mu          sync.RWMutex
	expedientes map[string]*Expediente
	propietario map[string]string
	documentos  map[string]Documento
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		expedientes: map[string]*Expediente{},
		propietario: map[string]string{},
		documentos:  map[string]Documento{},
	}
}

func (r *RepositorioMemoria) Guardar(_ context.Context, expediente *Expediente) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copia := Copiar(expediente)
	anterior := r.expedientes[expediente.TutorID]
	if anterior != nil {
		copia.Documentos = append(copia.Documentos, anterior.Documentos...)
	}
	for _, documento := range expediente.Documentos {
		r.propietario[documento.ID] = expediente.TutorID
		r.documentos[documento.ID] = documento
	}
	r.expedientes[expediente.TutorID] = copia
	return nil
}

func (r *RepositorioMemoria) PorTutor(_ context.Context, tutorID string) (*Expediente, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	expediente, ok := r.expedientes[tutorID]
	if !ok {
		return nil, ErrNoEncontrado
	}
	return Copiar(expediente), nil
}

func (r *RepositorioMemoria) Pendientes(_ context.Context) ([]Expediente, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := make([]Expediente, 0)
	for _, expediente := range r.expedientes {
		if expediente.Estado == EstadoPendiente {
			resultado = append(resultado, *Copiar(expediente))
		}
	}
	return resultado, nil
}

func (r *RepositorioMemoria) Revisar(_ context.Context, expediente *Expediente) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.expedientes[expediente.TutorID]; !ok {
		return ErrNoEncontrado
	}
	r.expedientes[expediente.TutorID] = Copiar(expediente)
	return nil
}

func (r *RepositorioMemoria) Documento(_ context.Context, id string) (Documento, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	documento, ok := r.documentos[id]
	if !ok {
		return Documento{}, "", ErrNoEncontrado
	}
	return documento, r.propietario[id], nil
}

func (r *RepositorioMemoria) PerfilesPublicos(_ context.Context, ahora time.Time) ([]PerfilListado, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := make([]PerfilListado, 0)
	for _, expediente := range r.expedientes {
		if expediente.Publicable(ahora) {
			resultado = append(resultado, PerfilListado{
				TutorID: expediente.TutorID, Presentacion: expediente.Presentacion,
				Cursos: expediente.Cursos, Experiencia: expediente.Experiencia,
			})
		}
	}
	sort.Slice(resultado, func(i, j int) bool { return resultado[i].TutorID < resultado[j].TutorID })
	return resultado, nil
}
