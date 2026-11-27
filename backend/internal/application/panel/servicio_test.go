package panel

import (
	"context"
	"errors"
	"testing"
	"time"

	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/domain/panel"
)

type repositorioPanelPrueba struct {
	metricas  *modelo.Metricas
	eventos   []modelo.EventoAuditoria
	filtro    modelo.FiltroAuditoria
	errorRepo error
}

func (r *repositorioPanelPrueba) Metricas(context.Context, time.Time, string) (*modelo.Metricas, error) {
	return r.metricas, r.errorRepo
}

func (r *repositorioPanelPrueba) Auditoria(_ context.Context, filtro modelo.FiltroAuditoria) ([]modelo.EventoAuditoria, error) {
	r.filtro = filtro
	return r.eventos, r.errorRepo
}

func TestServicioPanelAutorizaYNormaliza(t *testing.T) {
	repo := &repositorioPanelPrueba{metricas: &modelo.Metricas{}, eventos: []modelo.EventoAuditoria{{ID: 1}}}
	servicio := NuevoServicio(repo)
	servicio.FijarReloj(func() time.Time { return time.Date(2026, 11, 26, 12, 0, 0, 0, time.UTC) })
	ctx := context.Background()

	if _, err := servicio.Metricas(ctx, Solicitante{}); !errors.Is(err, modelo.ErrNoAutorizado) {
		t.Fatalf("se esperaba no autorizado: %v", err)
	}
	metricas, err := servicio.Metricas(ctx, Solicitante{Rol: modeloAcceso.RolCoordinador, FacultadID: "fac-1"})
	if err != nil || metricas == nil {
		t.Fatalf("métricas inesperadas: %v %v", metricas, err)
	}
	eventos, err := servicio.Auditoria(ctx, Solicitante{Rol: modeloAcceso.RolCoordinador, FacultadID: "fac-1"}, modelo.FiltroAuditoria{})
	if err != nil || len(eventos) != 1 {
		t.Fatalf("eventos inesperados: %v %v", eventos, err)
	}
	if repo.filtro.FacultadID != "fac-1" || repo.filtro.Limite != modelo.LimiteAuditoriaPorDefecto {
		t.Fatalf("filtro no normalizado: %+v", repo.filtro)
	}
	eventos, err = servicio.Auditoria(ctx, Solicitante{Rol: modeloAcceso.RolCoordinadorGeneral, FacultadID: "fac-1"}, modelo.FiltroAuditoria{Limite: 999})
	if err != nil || len(eventos) != 1 || repo.filtro.Limite != modelo.LimiteAuditoriaMaximo {
		t.Fatalf("límite no acotado: %+v err=%v", repo.filtro, err)
	}
	if _, err := servicio.Auditoria(ctx, Solicitante{Rol: modeloAcceso.RolEscolar}, modelo.FiltroAuditoria{}); !errors.Is(err, modelo.ErrNoAutorizado) {
		t.Fatalf("se esperaba no autorizado: %v", err)
	}
}
