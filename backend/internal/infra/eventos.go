package infra

import (
	"log"
	"sync"

	modelo "uniapp/internal/models/acceso"
)

type Bus struct {
	mu           sync.RWMutex
	observadores map[string][]modelo.Observador
}

func NuevoBus() *Bus { return &Bus{observadores: map[string][]modelo.Observador{}} }

func (b *Bus) Registrar(nombre string, observador modelo.Observador) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.observadores[nombre] = append(b.observadores[nombre], observador)
}

func (b *Bus) Desregistrar(nombre string, observador modelo.Observador) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.observadores[nombre] = sinObservador(b.observadores[nombre], observador)
}

func (b *Bus) Notificar(evento modelo.Event) {
	b.mu.RLock()
	suscritos := append([]modelo.Observador(nil), b.observadores[evento.Nombre]...)
	b.mu.RUnlock()
	for _, observador := range suscritos {
		notificarConAislamiento(observador, evento)
	}
}

func sinObservador(observadores []modelo.Observador, aQuitar modelo.Observador) []modelo.Observador {
	restantes := make([]modelo.Observador, 0, len(observadores))
	for _, observador := range observadores {
		if observador.ID() != aQuitar.ID() {
			restantes = append(restantes, observador)
		}
	}
	return restantes
}

func notificarConAislamiento(observador modelo.Observador, evento modelo.Event) {
	defer func() {
		if causa := recover(); causa != nil {
			log.Printf("observador %s falló en %s: %v", observador.ID(), evento.Nombre, causa)
		}
	}()
	observador.Actualizar(evento)
}

type ObservadorFunc struct {
	identificador string
	alRecibir     func(modelo.Event)
}

func NuevoObservadorFunc(identificador string, alRecibir func(modelo.Event)) *ObservadorFunc {
	return &ObservadorFunc{identificador: identificador, alRecibir: alRecibir}
}

func (o *ObservadorFunc) Actualizar(evento modelo.Event) { o.alRecibir(evento) }
func (o *ObservadorFunc) ID() string                     { return o.identificador }
