package sesion

import (
	"context"
	"sync"
	"time"
)

type RepositorioMemoria struct {
	mutex            sync.Mutex
	sesionesPorID    map[string]*Sesion
	reservasOcupadas map[string]bool
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		sesionesPorID:    make(map[string]*Sesion),
		reservasOcupadas: make(map[string]bool),
	}
}

func (repositorio *RepositorioMemoria) CrearSesion(
	_ context.Context,
	sesion *Sesion,
) error {
	if err := sesion.Validar(); err != nil {
		return err
	}

	repositorio.mutex.Lock()
	defer repositorio.mutex.Unlock()

	if _, existe := repositorio.sesionesPorID[sesion.ID()]; existe {
		return ErrSesionDuplicada
	}
	if repositorio.reservasOcupadas[sesion.ReservaID()] {
		return ErrSesionDuplicada
	}

	copiaSesion := sesion.Clonar()
	repositorio.sesionesPorID[sesion.ID()] = copiaSesion
	repositorio.reservasOcupadas[sesion.ReservaID()] = true
	return nil
}

func (repositorio *RepositorioMemoria) SesionPorID(
	_ context.Context,
	sesionID string,
) (*Sesion, error) {
	repositorio.mutex.Lock()
	defer repositorio.mutex.Unlock()

	sesion, existe := repositorio.sesionesPorID[sesionID]
	if !existe {
		return nil, ErrSesionNoEncontrada
	}
	return sesion.Clonar(), nil
}

func (repositorio *RepositorioMemoria) SesionPorReservaID(
	_ context.Context,
	reservaID string,
) (*Sesion, error) {
	repositorio.mutex.Lock()
	defer repositorio.mutex.Unlock()

	for _, sesion := range repositorio.sesionesPorID {
		if sesion.ReservaID() == reservaID {
			return sesion.Clonar(), nil
		}
	}
	return nil, ErrSesionNoEncontrada
}

func (repositorio *RepositorioMemoria) DeclararAsistencia(
	_ context.Context,
	sesionID string,
	usuarioID string,
	momento time.Time,
) (Asistencia, error) {
	repositorio.mutex.Lock()
	defer repositorio.mutex.Unlock()

	sesion, existe := repositorio.sesionesPorID[sesionID]
	if !existe {
		return Asistencia{}, ErrSesionNoEncontrada
	}
	if err := sesion.DeclararAsistencia(usuarioID, momento); err != nil {
		return Asistencia{}, err
	}

	asistencias := sesion.Asistencias()
	return asistencias[len(asistencias)-1], nil
}

func (r *RepositorioMemoria) ReportarAusencia(
	_ context.Context, sesionID string, usuarioID string, ahora time.Time,
) (Ausencia, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	s, existe := r.sesionesPorID[sesionID]
	if !existe {
		return Ausencia{}, ErrSesionNoEncontrada
	}
	if err := s.ReportarAusencia(usuarioID, ahora); err != nil {
		return Ausencia{}, err
	}
	ausencias := s.Ausencias()
	return ausencias[len(ausencias)-1], nil
}

func (r *RepositorioMemoria) RegistrarResultado(
	_ context.Context, sesionID string, usuarioID string, resumen string, ahora time.Time,
) (Resultado, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	s, existe := r.sesionesPorID[sesionID]
	if !existe {
		return Resultado{}, ErrSesionNoEncontrada
	}
	if err := s.RegistrarResultado(usuarioID, resumen, ahora); err != nil {
		return Resultado{}, err
	}
	return *s.Resultado(), nil
}
