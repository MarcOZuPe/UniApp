package reserva

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// RepositorioMemoria sirve para pruebas y para demos sin base de datos. Conserva la
// misma regla de unicidad por slot que la tabla reservas.
type RepositorioMemoria struct {
	mu        sync.Mutex
	secuencia int
	reservas  map[string]*Reserva
	slots     map[string]string
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{reservas: map[string]*Reserva{}, slots: map[string]string{}}
}

func (r *RepositorioMemoria) Crear(_ context.Context, reserva *Reserva) (*Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, tomado := r.slots[reserva.SlotID]; tomado {
		return nil, ErrSlotOcupado
	}
	r.secuencia++
	guardada := *reserva
	guardada.ID = fmt.Sprintf("r-%d", r.secuencia)
	guardada.Adjuntos = copiarAdjuntos(reserva.Adjuntos)
	r.reservas[guardada.ID] = &guardada
	r.slots[guardada.SlotID] = guardada.ID
	copia := guardada
	return &copia, nil
}

func (r *RepositorioMemoria) PorID(_ context.Context, id string) (*Reserva, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	guardada, ok := r.reservas[id]
	if !ok {
		return nil, ErrReservaNoEncontrada
	}
	copia := *guardada
	copia.Adjuntos = copiarAdjuntos(guardada.Adjuntos)
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
		copia.Adjuntos = copiarAdjuntos(guardada.Adjuntos)
		resultado = append(resultado, copia)
	}
	sort.Slice(resultado, func(i, j int) bool { return resultado[i].Inicio.Before(resultado[j].Inicio) })
	return resultado, nil
}

func (r *RepositorioMemoria) SlotTomado(_ context.Context, slotID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, tomado := r.slots[slotID]
	return tomado, nil
}

func copiarAdjuntos(adjuntos []Adjunto) []Adjunto {
	if len(adjuntos) == 0 {
		return nil
	}
	copia := make([]Adjunto, len(adjuntos))
	copy(copia, adjuntos)
	return copia
}
