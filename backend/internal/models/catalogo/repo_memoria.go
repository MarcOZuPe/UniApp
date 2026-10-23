package catalogo

import (
	"context"
	"sort"
	"sync"
)

type RepositorioMemoria struct {
	mu                sync.RWMutex
	siguienteImportID int64
	siguienteCursoID  int64
	siguienteSolID    int64
	importaciones     []*Importacion
	cursos            map[int64]Curso
	solicitudes       map[int64]SolicitudColegio
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		cursos:      make(map[int64]Curso),
		solicitudes: make(map[int64]SolicitudColegio),
	}
}

func (r *RepositorioMemoria) Importar(_ context.Context, importacion *Importacion) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.siguienteImportID++
	importacion.ID = r.siguienteImportID
	for indice := range importacion.Cursos {
		r.siguienteCursoID++
		importacion.Cursos[indice].ID = r.siguienteCursoID
	}
	copia := copiarImportacion(importacion)
	r.importaciones = append(r.importaciones, copia)
	for _, curso := range copia.Cursos {
		r.cursos[curso.ID] = curso
	}
	for id, solicitud := range r.solicitudes {
		for _, colegio := range copia.Colegios {
			if claveNormalizada(solicitud.Nombre)+"\x00"+claveNormalizada(solicitud.Distrito) ==
				claveNormalizada(colegio.Nombre)+"\x00"+claveNormalizada(colegio.Distrito) {
				solicitud.Estado = EstadoSolicitudIncorporada
				r.solicitudes[id] = solicitud
				break
			}
		}
	}
	return nil
}

func (r *RepositorioMemoria) ImportacionActual(_ context.Context) (*Importacion, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.importaciones) == 0 {
		return nil, ErrNoEncontrado
	}
	return copiarImportacion(r.importaciones[len(r.importaciones)-1]), nil
}

func (r *RepositorioMemoria) Curso(_ context.Context, id int64) (Curso, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	curso, existe := r.cursos[id]
	if !existe || len(r.importaciones) == 0 || curso.ID <= 0 {
		return Curso{}, ErrNoEncontrado
	}
	actual := r.importaciones[len(r.importaciones)-1]
	for _, cursoActual := range actual.Cursos {
		if cursoActual.ID == id {
			return cursoActual, nil
		}
	}
	return Curso{}, ErrNoEncontrado
}

func (r *RepositorioMemoria) RevisarCurso(_ context.Context, curso Curso, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.importaciones) == 0 {
		return ErrNoEncontrado
	}
	estaEnImportacionActual := false
	for _, actual := range r.importaciones[len(r.importaciones)-1].Cursos {
		if actual.ID == curso.ID {
			estaEnImportacionActual = true
			break
		}
	}
	if !estaEnImportacionActual {
		return ErrEstadoInvalido
	}
	actual, existe := r.cursos[curso.ID]
	if !existe {
		return ErrNoEncontrado
	}
	if actual.Estado != EstadoPendiente {
		return ErrEstadoInvalido
	}
	r.cursos[curso.ID] = curso
	for indice := range r.importaciones {
		for cursoIndice := range r.importaciones[indice].Cursos {
			if r.importaciones[indice].Cursos[cursoIndice].ID == curso.ID {
				r.importaciones[indice].Cursos[cursoIndice] = curso
			}
		}
	}
	return nil
}

func (r *RepositorioMemoria) SolicitarColegio(_ context.Context, solicitud SolicitudColegio, _ string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clave := claveNormalizada(solicitud.Nombre) + "\x00" + claveNormalizada(solicitud.Distrito)
	for _, existente := range r.solicitudes {
		if existente.Estado == EstadoSolicitudPendiente &&
			claveNormalizada(existente.Nombre)+"\x00"+claveNormalizada(existente.Distrito) == clave {
			return ErrSolicitudDuplicada
		}
	}
	r.siguienteSolID++
	solicitud.ID = r.siguienteSolID
	solicitud.Estado = EstadoSolicitudPendiente
	r.solicitudes[solicitud.ID] = solicitud
	return nil
}

func (r *RepositorioMemoria) SolicitudesPendientes(_ context.Context) ([]SolicitudColegio, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resultado := make([]SolicitudColegio, 0, len(r.solicitudes))
	for _, solicitud := range r.solicitudes {
		if solicitud.Estado == EstadoSolicitudPendiente {
			resultado = append(resultado, solicitud)
		}
	}
	sort.Slice(resultado, func(i, j int) bool {
		if resultado[i].SolicitadoEn.Equal(resultado[j].SolicitadoEn) {
			return resultado[i].ID < resultado[j].ID
		}
		return resultado[i].SolicitadoEn.Before(resultado[j].SolicitadoEn)
	})
	return resultado, nil
}
