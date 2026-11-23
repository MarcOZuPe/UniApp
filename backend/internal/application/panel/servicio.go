package panel

import (
	"context"
	"time"

	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/domain/panel"
)

type Repositorio interface {
	Metricas(ctx context.Context, ahora time.Time) (*modelo.Metricas, error)
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

func (s *Servicio) Metricas(ctx context.Context, rol modeloAcceso.Rol) (*modelo.Metricas, error) {
	if !rol.EsCoordinacion() {
		return nil, modelo.ErrNoAutorizado
	}
	return s.repositorio.Metricas(ctx, s.ahora().UTC())
}

func (s *Servicio) Auditoria(
	ctx context.Context,
	rol modeloAcceso.Rol,
	filtro modelo.FiltroAuditoria,
) ([]modelo.EventoAuditoria, error) {
	if !rol.EsCoordinacion() {
		return nil, modelo.ErrNoAutorizado
	}
	if filtro.Limite <= 0 {
		filtro.Limite = modelo.LimiteAuditoriaPorDefecto
	}
	if filtro.Limite > modelo.LimiteAuditoriaMaximo {
		filtro.Limite = modelo.LimiteAuditoriaMaximo
	}
	return s.repositorio.Auditoria(ctx, filtro)
}
