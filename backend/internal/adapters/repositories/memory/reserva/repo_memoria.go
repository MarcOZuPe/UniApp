package reserva

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// RepositorioMemoria sirve para pruebas y demos sin base de datos. Reproduce las mismas
// reglas que la transacción de PostgreSQL: unicidad por slot, solapamiento del escolar,
// cuota de D07 e idempotencia.
type RepositorioMemoria struct {
	mu                 sync.Mutex
	secuencia          int
	secuenciaAdjunto   int
	reservas           map[string]*Reserva
	slots              map[string]string
	idempotencia       map[string]string
	adjuntos           map[string]*Adjunto
	adjuntosPorReserva map[string][]string
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		reservas:           map[string]*Reserva{},
		slots:              map[string]string{},
		idempotencia:       map[string]string{},
		adjuntos:           map[string]*Adjunto{},
		adjuntosPorReserva: map[string][]string{},
	}
}

func claveIdempotencia(escolarID, clave string) string { return escolarID + "\x00" + clave }

func (r *RepositorioMemoria) ReservarTransaccional(_ context.Context, reserva *Reserva, limites Limites) (*Reserva, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if reserva.ClaveIdempotencia != "" {
		if id, existe := r.idempotencia[claveIdempotencia(reserva.EscolarID, reserva.ClaveIdempotencia)]; existe {
			guardada := r.reservas[id]
			copia := *guardada
			copia.Adjuntos = r.copiarAdjuntosDe(guardada.ID)
			return &copia, false, nil
		}
	}
	if _, tomado := r.slots[reserva.SlotID]; tomado {
		return nil, false, ErrSlotOcupado
	}
	futuras := 0
	for _, guardada := range r.reservas {
		if guardada.EscolarID != reserva.EscolarID || guardada.Estado != EstadoConfirmada {
			continue
		}
		if guardada.Inicio.After(reserva.CreadaEn) {
			futuras++
		}
		if guardada.Inicio.Before(reserva.Fin) && reserva.Inicio.Before(guardada.Fin) {
			return nil, false, ErrSolapamiento
		}
	}
	if futuras >= limites.MaxFuturas {
		return nil, false, ErrCuotaEscolar
	}
	inicioSemana, finSemana := SemanaLima(reserva.Inicio)
	delCurso := 0
	for _, guardada := range r.reservas {
		if guardada.EscolarID == reserva.EscolarID && guardada.Curso == reserva.Curso &&
			guardada.Estado == EstadoConfirmada &&
			!guardada.Inicio.Before(inicioSemana) && guardada.Inicio.Before(finSemana) {
			delCurso++
		}
	}
	if delCurso >= limites.MaxPorCursoSemana {
		return nil, false, ErrCuotaCursoSemana
	}

	r.secuencia++
	guardada := *reserva
	guardada.ID = fmt.Sprintf("r-%d", r.secuencia)
	guardada.Adjuntos = nil
	r.reservas[guardada.ID] = &guardada
	r.slots[guardada.SlotID] = guardada.ID
	if guardada.ClaveIdempotencia != "" {
		r.idempotencia[claveIdempotencia(guardada.EscolarID, guardada.ClaveIdempotencia)] = guardada.ID
	}
	copia := guardada
	return &copia, true, nil
}

func (r *RepositorioMemoria) PorClaveIdempotencia(_ context.Context, escolarID, clave string) (*Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	id, existe := r.idempotencia[claveIdempotencia(escolarID, clave)]
	if !existe {
		return nil, ErrReservaNoEncontrada
	}
	guardada := r.reservas[id]
	copia := *guardada
	copia.Adjuntos = r.copiarAdjuntosDe(guardada.ID)
	return &copia, nil
}

func (r *RepositorioMemoria) PorID(_ context.Context, id string) (*Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	guardada, existe := r.reservas[id]
	if !existe {
		return nil, ErrReservaNoEncontrada
	}
	copia := *guardada
	copia.Adjuntos = r.copiarAdjuntosDe(guardada.ID)
	return &copia, nil
}

func (r *RepositorioMemoria) PorEscolar(_ context.Context, escolarID string) ([]Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var resultado []Reserva
	for _, guardada := range r.reservas {
		if guardada.EscolarID != escolarID {
			continue
		}
		copia := *guardada
		copia.Adjuntos = r.copiarAdjuntosDe(guardada.ID)
		resultado = append(resultado, copia)
	}
	sort.Slice(resultado, func(i, j int) bool { return resultado[i].Inicio.Before(resultado[j].Inicio) })
	return resultado, nil
}

func (r *RepositorioMemoria) PorTutor(_ context.Context, tutorID string) ([]Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	resultado := []Reserva{}

	for _, guardada := range r.reservas {
		if guardada.TutorID != tutorID {
			continue
		}

		copia := *guardada
		copia.Adjuntos = nil
		resultado = append(resultado, copia)
	}

	sort.Slice(resultado, func(i, j int) bool {
		if resultado[i].Inicio.Equal(resultado[j].Inicio) {
			return resultado[i].ID < resultado[j].ID
		}
		return resultado[i].Inicio.Before(resultado[j].Inicio)
	})

	return resultado, nil
}

func (r *RepositorioMemoria) SlotTomado(_ context.Context, slotID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, tomado := r.slots[slotID]
	return tomado, nil
}

func (r *RepositorioMemoria) ContarAdjuntos(_ context.Context, reservaID string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.adjuntosPorReserva[reservaID]), nil
}

func (r *RepositorioMemoria) AgregarAdjunto(_ context.Context, reservaID string, adjunto *Adjunto) (*Adjunto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, existe := r.reservas[reservaID]; !existe {
		return nil, ErrReservaNoEncontrada
	}
	r.secuenciaAdjunto++
	guardado := *adjunto
	guardado.ID = fmt.Sprintf("a-%d", r.secuenciaAdjunto)
	if guardado.CreadoEn.IsZero() {
		guardado.CreadoEn = time.Now().UTC()
	}
	r.adjuntos[guardado.ID] = &guardado
	r.adjuntosPorReserva[reservaID] = append(r.adjuntosPorReserva[reservaID], guardado.ID)
	copia := guardado
	return &copia, nil
}

func (r *RepositorioMemoria) AdjuntosDe(_ context.Context, reservaID string) ([]Adjunto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.copiarAdjuntosDe(reservaID), nil
}

func (r *RepositorioMemoria) AdjuntoPorID(_ context.Context, reservaID, adjuntoID string) (*Adjunto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	guardado, existe := r.adjuntos[adjuntoID]
	if !existe || !contieneID(r.adjuntosPorReserva[reservaID], adjuntoID) {
		return nil, ErrAdjuntoNoEncontrado
	}
	copia := *guardado
	return &copia, nil
}

func (r *RepositorioMemoria) ActualizarAdjunto(_ context.Context, adjuntoID string, estado EstadoAdjunto, motivo string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	guardado, existe := r.adjuntos[adjuntoID]
	if !existe {
		return ErrAdjuntoNoEncontrado
	}
	guardado.Estado = estado
	guardado.Motivo = motivo
	return nil
}

func (r *RepositorioMemoria) copiarAdjuntosDe(reservaID string) []Adjunto {
	ids := r.adjuntosPorReserva[reservaID]
	if len(ids) == 0 {
		return nil
	}
	copia := make([]Adjunto, 0, len(ids))
	for _, id := range ids {
		copia = append(copia, *r.adjuntos[id])
	}
	return copia
}

func contieneID(ids []string, buscado string) bool {
	for _, id := range ids {
		if id == buscado {
			return true
		}
	}
	return false
}
