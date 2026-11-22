package sesion

import (
	"context"
	"sort"
	"sync"
)

// RepositorioReportesMemoria implementa el puerto de I18 sobre las sesiones
// en memoria. Las constancias solo se agregan; nunca se modifican.
type RepositorioReportesMemoria struct {
	sesiones  *RepositorioMemoria
	mutex     sync.Mutex
	registros []RegistroReporte
}

func NuevoRepositorioReportesMemoria(sesiones *RepositorioMemoria) *RepositorioReportesMemoria {
	return &RepositorioReportesMemoria{sesiones: sesiones}
}

func (r *RepositorioReportesMemoria) SesionesEntre(_ context.Context, filtro FiltroReporte) ([]*Sesion, error) {
	r.sesiones.mutex.Lock()
	defer r.sesiones.mutex.Unlock()

	resultado := []*Sesion{}
	for _, sesion := range r.sesiones.sesionesPorID {
		if filtro.Incluye(sesion) {
			resultado = append(resultado, sesion.Clonar())
		}
	}
	sort.Slice(resultado, func(i, j int) bool {
		if !resultado[i].Inicio().Equal(resultado[j].Inicio()) {
			return resultado[i].Inicio().Before(resultado[j].Inicio())
		}
		return resultado[i].ID() < resultado[j].ID()
	})
	return resultado, nil
}

func (r *RepositorioReportesMemoria) RegistrarReporte(_ context.Context, registro RegistroReporte) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, existente := range r.registros {
		if existente.ID == registro.ID {
			return ErrReporteInvalido
		}
	}
	r.registros = append(r.registros, registro)
	return nil
}

func (r *RepositorioReportesMemoria) ReportesRecientes(_ context.Context, limite int) ([]RegistroReporte, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	copia := append([]RegistroReporte(nil), r.registros...)
	sort.SliceStable(copia, func(i, j int) bool { return copia[i].GeneradoEn.After(copia[j].GeneradoEn) })
	if limite > 0 && len(copia) > limite {
		copia = copia[:limite]
	}
	return copia, nil
}
