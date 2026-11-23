package panel

import (
	"time"

	modelo "uniapp/internal/domain/panel"
)

type ReservasMetricasJSON struct {
	Totales     int `json:"totales"`
	Futuras     int `json:"futuras"`
	Finalizadas int `json:"finalizadas"`
}

type AsistenciaMetricasJSON struct {
	SesionesFinalizadas  int     `json:"sesiones_finalizadas"`
	ConAsistenciaEscolar int     `json:"con_asistencia_escolar"`
	ConAsistenciaTutor   int     `json:"con_asistencia_tutor"`
	TasaEscolar          float64 `json:"tasa_escolar_porcentaje"`
	TasaTutor            float64 `json:"tasa_tutor_porcentaje"`
}

type HorasMetricasJSON struct {
	Internas float64 `json:"internas"`
}

type GestionMetricasJSON struct {
	ExpedientesPendientes        int `json:"expedientes_pendientes"`
	IncidentesPendientes         int `json:"incidentes_pendientes"`
	CursosPendientes             int `json:"cursos_pendientes"`
	SolicitudesColegioPendientes int `json:"solicitudes_colegio_pendientes"`
}

type MetricasJSON struct {
	Reservas      ReservasMetricasJSON   `json:"reservas"`
	Asistencia    AsistenciaMetricasJSON `json:"asistencia"`
	Horas         HorasMetricasJSON      `json:"horas"`
	Gestion       GestionMetricasJSON    `json:"gestion"`
	ActualizadoEn time.Time              `json:"actualizado_en"`
}

type EventoAuditoriaJSON struct {
	ID          int64          `json:"id"`
	OcurridoEn  time.Time      `json:"ocurrido_en"`
	ActorID     string         `json:"actor_id,omitempty"`
	ActorNombre string         `json:"actor_nombre,omitempty"`
	ActorRol    string         `json:"actor_rol"`
	Accion      string         `json:"accion"`
	Entidad     string         `json:"entidad"`
	EntidadID   string         `json:"entidad_id"`
	Detalle     map[string]any `json:"detalle"`
}

func PresentarMetricas(metricas *modelo.Metricas) MetricasJSON {
	if metricas == nil {
		return MetricasJSON{}
	}
	return MetricasJSON{
		Reservas: ReservasMetricasJSON{
			Totales:     metricas.Reservas.Totales,
			Futuras:     metricas.Reservas.Futuras,
			Finalizadas: metricas.Reservas.Finalizadas,
		},
		Asistencia: AsistenciaMetricasJSON{
			SesionesFinalizadas:  metricas.Asistencia.SesionesFinalizadas,
			ConAsistenciaEscolar: metricas.Asistencia.ConAsistenciaEscolar,
			ConAsistenciaTutor:   metricas.Asistencia.ConAsistenciaTutor,
			TasaEscolar:          metricas.Asistencia.TasaEscolarPorcentaje,
			TasaTutor:            metricas.Asistencia.TasaTutorPorcentaje,
		},
		Horas: HorasMetricasJSON{Internas: metricas.Horas.Internas},
		Gestion: GestionMetricasJSON{
			ExpedientesPendientes:        metricas.Gestion.ExpedientesPendientes,
			IncidentesPendientes:         metricas.Gestion.IncidentesPendientes,
			CursosPendientes:             metricas.Gestion.CursosPendientes,
			SolicitudesColegioPendientes: metricas.Gestion.SolicitudesColegioPendientes,
		},
		ActualizadoEn: metricas.ActualizadoEn,
	}
}

func PresentarAuditoria(eventos []modelo.EventoAuditoria) []EventoAuditoriaJSON {
	presentados := make([]EventoAuditoriaJSON, 0, len(eventos))
	for _, evento := range eventos {
		detalle := evento.Detalle
		if detalle == nil {
			detalle = map[string]any{}
		}
		presentados = append(presentados, EventoAuditoriaJSON{
			ID:          evento.ID,
			OcurridoEn:  evento.OcurridoEn,
			ActorID:     evento.ActorID,
			ActorNombre: evento.ActorNombre,
			ActorRol:    evento.ActorRol,
			Accion:      evento.Accion,
			Entidad:     evento.Entidad,
			EntidadID:   evento.EntidadID,
			Detalle:     detalle,
		})
	}
	return presentados
}
