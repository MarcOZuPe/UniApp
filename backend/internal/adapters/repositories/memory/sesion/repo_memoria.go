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
