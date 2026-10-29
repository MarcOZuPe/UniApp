package sesion

import (
	"context"
	"sync"
	"time"
)

type RepositorioMemoria struct {
	mu       sync.Mutex
	sesiones map[string]*Sesion
	reservas map[string]bool
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		sesiones: make(map[string]*Sesion),
		reservas: make(map[string]bool),
	}
}

func (r *RepositorioMemoria) CrearSesion(
	_ context.Context,
	s *Sesion,
) error {
	if err := s.Validar(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.sesiones[s.ID()]; existe {
		return ErrSesionDuplicada
	}

	if r.reservas[s.ReservaID()] {
		return ErrSesionDuplicada
	}

	copia := s.Clonar()
	r.sesiones[s.ID()] = copia
	r.reservas[s.ReservaID()] = true

	return nil
}

func (r *RepositorioMemoria) SesionPorID(
	_ context.Context,
	id string,
) (*Sesion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, existe := r.sesiones[id]
	if !existe {
		return nil, ErrSesionNoEncontrada
	}

	return s.Clonar(), nil
}

func (r *RepositorioMemoria) SesionPorReservaID(
	_ context.Context,
	reservaID string,
) (*Sesion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, s := range r.sesiones {
		if s.ReservaID() == reservaID {
			return s.Clonar(), nil
		}
	}

	return nil, ErrSesionNoEncontrada
}

func (r *RepositorioMemoria) DeclararAsistencia(
	_ context.Context,
	sesionID string,
	usuarioID string,
	ahora time.Time,
) (Asistencia, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	s, existe := r.sesiones[sesionID]
	if !existe {
		return Asistencia{}, ErrSesionNoEncontrada
	}

	if err := s.DeclararAsistencia(usuarioID, ahora); err != nil {
		return Asistencia{}, err
	}

	asistencias := s.Asistencias()
	return asistencias[len(asistencias)-1], nil
}
