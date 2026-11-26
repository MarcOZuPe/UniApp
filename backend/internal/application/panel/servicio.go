package panel

import (
	"context"
	"time"

	modelo "uniapp/internal/domain/panel"
)

type Repositorio interface {
	Metricas(ctx context.Context, ahora time.Time, facultadID string) (*modelo.Metricas, error)
	Auditoria(ctx context.Context, filtro modelo.FiltroAuditoria) ([]modelo.EventoAuditoria, error)
}

type Servicio struct {
	repositorio Repositorio
	ahora       func() time.Time
}

func NuevoServicio(repositorio Repositorio) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		ahora:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Servicio) FijarReloj(f func() time.Time) { s.ahora = f }

func (s *Servicio) Metricas(
	ctx context.Context,
	solicitante Solicitante,
) (*modelo.Metricas, error) {
	if !solicitante.Autorizado() {
		return nil, modelo.ErrNoAutorizado
	}
	return s.repositorio.Metricas(ctx, s.ahora().UTC(), solicitante.FacultadID)
}

func (s *Servicio) Auditoria(
	ctx context.Context,
	solicitante Solicitante,
	filtro modelo.FiltroAuditoria,
) ([]modelo.EventoAuditoria, error) {
	if !solicitante.Autorizado() {
		return nil, modelo.ErrNoAutorizado
	}
	filtro.FacultadID = solicitante.FacultadID
	return s.repositorio.Auditoria(ctx, filtro.Normalizado())
}
