package panel

import (
	"errors"
	"time"
)

const (
	LimiteAuditoriaPorDefecto = 50
	LimiteAuditoriaMaximo     = 200
)

var ErrNoAutorizado = errors.New("no autorizado para consultar el panel")

type ReservasMetricas struct {
	Totales     int
	Futuras     int
	Finalizadas int
}

type AsistenciaMetricas struct {
	SesionesFinalizadas   int
	ConAsistenciaEscolar  int
	ConAsistenciaTutor    int
	TasaEscolarPorcentaje float64
	TasaTutorPorcentaje   float64
}

type HorasMetricas struct {
	Internas float64
}

type GestionMetricas struct {
	ExpedientesPendientes        int
	IncidentesPendientes         int
	CursosPendientes             int
	SolicitudesColegioPendientes int
}

type Metricas struct {
	Reservas      ReservasMetricas
	Asistencia    AsistenciaMetricas
	Horas         HorasMetricas
	Gestion       GestionMetricas
	ActualizadoEn time.Time
}

type EventoAuditoria struct {
	ID          int64
	OcurridoEn  time.Time
	ActorID     string
	ActorNombre string
	ActorRol    string
	Accion      string
	Entidad     string
	EntidadID   string
	Detalle     map[string]any
}

type FiltroAuditoria struct {
	Limite int
}
