package panel

import (
	"testing"
	"time"

	modelo "uniapp/internal/domain/panel"
)

func TestPresentarMetricasYAuditoria(t *testing.T) {
	metricas := &modelo.Metricas{
		Reservas:      modelo.ReservasMetricas{Totales: 10, Futuras: 4, Finalizadas: 6},
		Asistencia:    modelo.AsistenciaMetricas{SesionesFinalizadas: 6, ConAsistenciaEscolar: 5, ConAsistenciaTutor: 4, TasaEscolarPorcentaje: 83.3, TasaTutorPorcentaje: 66.6},
		Horas:         modelo.HorasMetricas{Internas: 6},
		Gestion:       modelo.GestionMetricas{ExpedientesPendientes: 2, IncidentesPendientes: 1, CursosPendientes: 3, SolicitudesColegioPendientes: 4},
		ActualizadoEn: time.Now(),
	}
	presentadas := PresentarMetricas(metricas)
	if presentadas.Reservas.Totales != 10 || presentadas.Asistencia.TasaEscolar != 83.3 || presentadas.Gestion.CursosPendientes != 3 {
		t.Fatalf("métricas presentadas inesperadas: %+v", presentadas)
	}
	if PresentarMetricas(nil) != (MetricasJSON{}) {
		t.Fatal("las métricas nulas deben devolver un valor vacío")
	}

	eventos := PresentarAuditoria([]modelo.EventoAuditoria{{
		ID: 1, ActorID: "u-1", ActorNombre: "Usuario", ActorRol: "coordinador",
		Accion: "incidente.revisado", Entidad: "incidente", EntidadID: "i-1", Detalle: nil,
	}})
	if len(eventos) != 1 || eventos[0].Detalle == nil || eventos[0].Accion != "incidente.revisado" {
		t.Fatalf("auditoría presentada inesperada: %+v", eventos)
	}
}
