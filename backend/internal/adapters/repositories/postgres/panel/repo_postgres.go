package panel

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"database/sql"

	modelo "uniapp/internal/domain/panel"
)

type RepositorioPostgres struct {
	DB *sql.DB
}

func (r RepositorioPostgres) Metricas(ctx context.Context, ahora time.Time) (*modelo.Metricas, error) {
	var metricas modelo.Metricas
	err := r.DB.QueryRowContext(ctx, `
		SELECT
			(SELECT count(*) FROM reservas),
			(SELECT count(*) FROM reservas WHERE estado = 'confirmada' AND inicio > $1),
			(SELECT count(*) FROM reservas WHERE fin <= $1),
			(SELECT count(*)
			 FROM sesion_sesiones s
			 JOIN reservas r ON r.id::text = s.reserva_id
			 WHERE s.fin <= $1),
			(SELECT count(DISTINCT a.sesion_id)
			 FROM sesion_asistencias a
			 JOIN sesion_sesiones s ON s.id = a.sesion_id
			 JOIN reservas r ON r.id::text = s.reserva_id
			 WHERE a.rol = 'escolar' AND s.fin <= $1),
			(SELECT count(DISTINCT a.sesion_id)
			 FROM sesion_asistencias a
			 JOIN sesion_sesiones s ON s.id = a.sesion_id
			 JOIN reservas r ON r.id::text = s.reserva_id
			 WHERE a.rol = 'tutor' AND s.fin <= $1),
			(SELECT COALESCE(SUM(EXTRACT(EPOCH FROM (s.fin - s.inicio)) / 3600.0), 0)
			 FROM sesion_sesiones s
			 JOIN reservas r ON r.id::text = s.reserva_id
			 WHERE s.fin <= $1),
			(SELECT count(*) FROM tutor_expedientes WHERE estado = 'pendiente'),
			(SELECT count(*) FROM incidentes WHERE estado = 'pendiente'),
			(SELECT count(*) FROM catalogo_cursos WHERE estado = 'pendiente'),
			(SELECT count(*) FROM catalogo_solicitudes_colegio WHERE estado = 'pendiente')
	`, ahora.UTC()).Scan(
		&metricas.Reservas.Totales,
		&metricas.Reservas.Futuras,
		&metricas.Reservas.Finalizadas,
		&metricas.Asistencia.SesionesFinalizadas,
		&metricas.Asistencia.ConAsistenciaEscolar,
		&metricas.Asistencia.ConAsistenciaTutor,
		&metricas.Horas.Internas,
		&metricas.Gestion.ExpedientesPendientes,
		&metricas.Gestion.IncidentesPendientes,
		&metricas.Gestion.CursosPendientes,
		&metricas.Gestion.SolicitudesColegioPendientes,
	)
	if err != nil {
		return nil, fmt.Errorf("consultar métricas del panel: %w", err)
	}
	metricas.Asistencia.TasaEscolarPorcentaje = porcentaje(
		metricas.Asistencia.ConAsistenciaEscolar,
		metricas.Asistencia.SesionesFinalizadas,
	)
	metricas.Asistencia.TasaTutorPorcentaje = porcentaje(
		metricas.Asistencia.ConAsistenciaTutor,
		metricas.Asistencia.SesionesFinalizadas,
	)
	metricas.Horas.Internas = math.Round(metricas.Horas.Internas*100) / 100
	metricas.ActualizadoEn = ahora.UTC()
	return &metricas, nil
}

func (r RepositorioPostgres) Auditoria(
	ctx context.Context,
	filtro modelo.FiltroAuditoria,
) ([]modelo.EventoAuditoria, error) {
	filas, err := r.DB.QueryContext(ctx, `
		SELECT a.id, a.ocurrido_en, COALESCE(a.actor_id::text, ''),
		       COALESCE(u.nombre_uso, ''), a.actor_rol, a.accion, a.entidad,
		       a.entidad_id, a.detalle
		FROM auditoria_eventos a
		LEFT JOIN usuarios u ON u.id = a.actor_id
		ORDER BY a.ocurrido_en DESC, a.id DESC
		LIMIT $1`, filtro.Limite)
	if err != nil {
		return nil, fmt.Errorf("consultar auditoría: %w", err)
	}
	defer filas.Close()

	eventos := make([]modelo.EventoAuditoria, 0)
	for filas.Next() {
		var (
			evento modelo.EventoAuditoria
			crudo  []byte
		)
		if err := filas.Scan(
			&evento.ID, &evento.OcurridoEn, &evento.ActorID, &evento.ActorNombre,
			&evento.ActorRol, &evento.Accion, &evento.Entidad, &evento.EntidadID, &crudo,
		); err != nil {
			return nil, err
		}
		evento.Detalle = map[string]any{}
		if len(crudo) > 0 {
			if err := json.Unmarshal(crudo, &evento.Detalle); err != nil {
				return nil, fmt.Errorf("leer detalle de auditoría: %w", err)
			}
		}
		eventos = append(eventos, evento)
	}
	return eventos, filas.Err()
}

func porcentaje(parte, total int) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(float64(parte)*10000/float64(total)) / 100
}
