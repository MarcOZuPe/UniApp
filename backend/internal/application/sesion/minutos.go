package sesion

import (
	"context"
	"time"

	dominio "uniapp/internal/domain/sesion"
)

// ServicioMinutos orquesta los casos de uso de I17. Las reglas viven en el
// dominio; aqui solo se combinan la sesion, el estado de la reserva (puerto
// hacia el flujo D) y la persistencia de los minutos.
type ServicioMinutos struct {
	Sesiones Repositorio
	Minutos  RepositorioMinutos
	Reservas EstadoReservas
}

// TotalMinutos resume lo que suma un tutor. Solo cuentan los verificados.
type TotalMinutos struct {
	TutorID     string
	Verificados int
	Registros   []*RegistroMinutos
}

// Proponer crea el registro de minutos de una sesion (lo hace el tutor).
func (s ServicioMinutos) Proponer(ctx context.Context, sesionID, tutorID string, minutos int, ahora time.Time) (*RegistroMinutos, error) {
	sesion, err := s.Sesiones.SesionPorID(ctx, sesionID)
	if err != nil {
		return nil, err
	}
	cancelada, err := s.Reservas.Cancelada(ctx, sesion.ReservaID())
	if err != nil {
		return nil, err
	}
	if cancelada {
		return nil, ErrSesionCancelada
	}
	registro, err := dominio.ProponerMinutos(sesion, tutorID, minutos, ahora)
	if err != nil {
		return nil, err
	}
	if err := s.Minutos.CrearMinutos(ctx, registro); err != nil {
		return nil, err
	}
	return registro, nil
}

// Consultar devuelve el registro de una sesion, anulandolo antes si la
// reserva fue cancelada para que nunca se muestre un valor que no suma.
func (s ServicioMinutos) Consultar(ctx context.Context, sesionID string, ahora time.Time) (*RegistroMinutos, error) {
	registro, err := s.Minutos.MinutosPorSesion(ctx, sesionID)
	if err != nil {
		return nil, err
	}
	return s.sincronizar(ctx, registro, ahora)
}

// Vista devuelve el registro tal como debe leerse ahora, sin escribir nada:
// si la reserva se cancelo, anula una copia en memoria. La usan las lecturas
// masivas (reporte de I18) para no repetir el efecto lateral de Consultar.
func (s ServicioMinutos) Vista(ctx context.Context, sesionID string, ahora time.Time) (*RegistroMinutos, error) {
	registro, err := s.Minutos.MinutosPorSesion(ctx, sesionID)
	if err != nil {
		return nil, err
	}
	if registro.Estado() == dominio.MinutosAnulados {
		return registro, nil
	}
	cancelada, err := s.Reservas.Cancelada(ctx, registro.ReservaID())
	if err != nil {
		return nil, err
	}
	if !cancelada {
		return registro, nil
	}
	copia := registro.Clonar()
	if err := copia.Anular(ahora); err != nil {
		return nil, err
	}
	return copia, nil
}

func (s ServicioMinutos) Aprobar(ctx context.Context, sesionID, coordinadorID string, ahora time.Time) (*RegistroMinutos, error) {
	return s.actuar(ctx, sesionID, ahora, func(r *RegistroMinutos) error {
		return r.Aprobar(coordinadorID, ahora)
	})
}

func (s ServicioMinutos) Ajustar(ctx context.Context, sesionID, coordinadorID string, minutos int, motivo string, ahora time.Time) (*RegistroMinutos, error) {
	return s.actuar(ctx, sesionID, ahora, func(r *RegistroMinutos) error {
		return r.Ajustar(coordinadorID, minutos, motivo, ahora)
	})
}

func (s ServicioMinutos) Disputar(ctx context.Context, sesionID, usuarioID, motivo string, ahora time.Time) (*RegistroMinutos, error) {
	return s.actuar(ctx, sesionID, ahora, func(r *RegistroMinutos) error {
		return r.Disputar(usuarioID, motivo, ahora)
	})
}

func (s ServicioMinutos) Resolver(ctx context.Context, sesionID, coordinadorID string, minutos int, motivo string, ahora time.Time) (*RegistroMinutos, error) {
	return s.actuar(ctx, sesionID, ahora, func(r *RegistroMinutos) error {
		return r.Resolver(coordinadorID, minutos, motivo, ahora)
	})
}

// Pendientes es la bandeja de coordinacion: propuestos y en disputa. Los que
// pertenecen a reservas canceladas se anulan y salen de la bandeja.
func (s ServicioMinutos) Pendientes(ctx context.Context, ahora time.Time) ([]*RegistroMinutos, error) {
	registros, err := s.Minutos.MinutosPendientes(ctx)
	if err != nil {
		return nil, err
	}
	pendientes := make([]*RegistroMinutos, 0, len(registros))
	for _, registro := range registros {
		actual, err := s.sincronizar(ctx, registro, ahora)
		if err != nil {
			return nil, err
		}
		if actual.Estado() == dominio.MinutosPropuestos || actual.Estado() == dominio.MinutosEnDisputa {
			pendientes = append(pendientes, actual)
		}
	}
	return pendientes, nil
}

// Total suma los minutos verificados de un tutor (insumo del reporte I18).
func (s ServicioMinutos) Total(ctx context.Context, tutorID string, ahora time.Time) (TotalMinutos, error) {
	registros, err := s.Minutos.MinutosDeTutor(ctx, tutorID)
	if err != nil {
		return TotalMinutos{}, err
	}
	total := TotalMinutos{TutorID: tutorID, Registros: make([]*RegistroMinutos, 0, len(registros))}
	for _, registro := range registros {
		actual, err := s.sincronizar(ctx, registro, ahora)
		if err != nil {
			return TotalMinutos{}, err
		}
		total.Verificados += actual.MinutosComputables()
		total.Registros = append(total.Registros, actual)
	}
	return total, nil
}

// actuar aplica un comando de forma atomica. Si la reserva se cancelo, el
// registro se anula (y se persiste) y el comando se rechaza.
func (s ServicioMinutos) actuar(ctx context.Context, sesionID string, ahora time.Time, comando func(*RegistroMinutos) error) (*RegistroMinutos, error) {
	registro, err := s.Minutos.MinutosPorSesion(ctx, sesionID)
	if err != nil {
		return nil, err
	}
	actual, err := s.sincronizar(ctx, registro, ahora)
	if err != nil {
		return nil, err
	}
	if actual.Estado() == dominio.MinutosAnulados {
		return nil, ErrSesionCancelada
	}
	return s.Minutos.ActualizarMinutos(ctx, sesionID, comando)
}

func (s ServicioMinutos) sincronizar(ctx context.Context, registro *RegistroMinutos, ahora time.Time) (*RegistroMinutos, error) {
	if registro.Estado() == dominio.MinutosAnulados {
		return registro, nil
	}
	cancelada, err := s.Reservas.Cancelada(ctx, registro.ReservaID())
	if err != nil {
		return nil, err
	}
	if !cancelada {
		return registro, nil
	}
	return s.Minutos.ActualizarMinutos(ctx, registro.SesionID(), func(r *RegistroMinutos) error {
		return r.Anular(ahora)
	})
}
