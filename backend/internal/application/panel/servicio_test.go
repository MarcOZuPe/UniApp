package panel

import (
	"context"
	"errors"
	"testing"
	"time"

	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/domain/panel"
)

type repositorioPrueba struct {
	metricas  *modelo.Metricas
	eventos   []modelo.EventoAuditoria
	filtro    modelo.FiltroAuditoria
	consultas int
}

func (r *repositorioPrueba) Metricas(context.Context, time.Time) (*modelo.Metricas, error) {
	r.consultas++
	return r.metricas, nil
}

func (r *repositorioPrueba) Auditoria(_ context.Context, filtro modelo.FiltroAuditoria) ([]modelo.EventoAuditoria, error) {
	r.consultas++
	r.filtro = filtro
	return r.eventos, nil
}

func TestServicioPanelSoloCoordinacion(t *testing.T) {
	repositorio := &repositorioPrueba{}
	servicio := NuevoServicio(repositorio)

	if _, err := servicio.Metricas(context.Background(), modeloAcceso.RolEscolar); !errors.Is(err, modelo.ErrNoAutorizado) {
		t.Fatalf("escolar no debe consultar métricas: %v", err)
	}
	if _, err := servicio.Auditoria(context.Background(), modeloAcceso.RolTutor, modelo.FiltroAuditoria{}); !errors.Is(err, modelo.ErrNoAutorizado) {
		t.Fatalf("tutor no debe consultar auditoría: %v", err)
	}
	if _, err := servicio.Metricas(context.Background(), modeloAcceso.RolCoordinadorGeneral); err != nil {
		t.Fatalf("coordinación general debe consultar métricas: %v", err)
	}
	if repositorio.consultas != 1 {
		t.Fatalf("no debe consultar el repositorio sin permiso: %d", repositorio.consultas)
	}
}

func TestServicioPanelAplicaLimitesDeAuditoria(t *testing.T) {
	repositorio := &repositorioPrueba{}
	servicio := NuevoServicio(repositorio)
	servicio.FijarReloj(func() time.Time { return time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC) })

	if _, err := servicio.Auditoria(context.Background(), modeloAcceso.RolCoordinador, modelo.FiltroAuditoria{}); err != nil {
		t.Fatal(err)
	}
	if repositorio.filtro.Limite != modelo.LimiteAuditoriaPorDefecto {
		t.Fatalf("límite por defecto: %d", repositorio.filtro.Limite)
	}
	if _, err := servicio.Auditoria(context.Background(), modeloAcceso.RolCoordinador, modelo.FiltroAuditoria{Limite: 1000}); err != nil {
		t.Fatal(err)
	}
	if repositorio.filtro.Limite != modelo.LimiteAuditoriaMaximo {
		t.Fatalf("límite máximo: %d", repositorio.filtro.Limite)
	}
}
