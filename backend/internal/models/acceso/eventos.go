package acceso

import (
	"log"
	"sync"
)

const (
	UsuarioRegistrado      = "usuario.registrado"
	IngresoExitoso         = "ingreso.exitoso"
	IngresoFallido         = "ingreso.fallido"
	RecuperacionSolicitada = "clave.recuperacion_solicitada"
	ClaveRestablecida      = "clave.restablecida"
	ConfirmacionSolicitada = "correo.confirmacion_solicitada"
	CorreoConfirmado       = "correo.confirmado"
	VinculacionSolicitada  = "vinculacion.solicitada"
	VinculacionAceptada    = "vinculacion.aceptada"
)

type Event struct {
	Nombre  string
	Payload map[string]any
}

type Observador interface {
	Actualizar(Event)
	ID() string
}

type Sujeto interface {
	Registrar(nombre string, observador Observador)
	Desregistrar(nombre string, observador Observador)
	Notificar(Event)
}

type Bus struct {
	mu           sync.RWMutex
	observadores map[string][]Observador
}

func NuevoBus() *Bus { return &Bus{observadores: map[string][]Observador{}} }

func (b *Bus) Registrar(nombre string, observador Observador) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.observadores[nombre] = append(b.observadores[nombre], observador)
}

func (b *Bus) Desregistrar(nombre string, observador Observador) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.observadores[nombre] = sinObservador(b.observadores[nombre], observador)
}

func (b *Bus) Notificar(evento Event) {
	b.mu.RLock()
	suscritos := append([]Observador(nil), b.observadores[evento.Nombre]...)
	b.mu.RUnlock()
	for _, observador := range suscritos {
		notificarConAislamiento(observador, evento)
	}
}

func sinObservador(observadores []Observador, aQuitar Observador) []Observador {
	restantes := make([]Observador, 0, len(observadores))
	for _, observador := range observadores {
		if observador.ID() != aQuitar.ID() {
			restantes = append(restantes, observador)
		}
	}
	return restantes
}

func notificarConAislamiento(observador Observador, evento Event) {
	defer func() {
		if causa := recover(); causa != nil {
			log.Printf("observador %s falló en %s: %v", observador.ID(), evento.Nombre, causa)
		}
	}()
	observador.Actualizar(evento)
}

type ObservadorFunc struct {
	identificador string
	alRecibir     func(Event)
}

func NuevoObservadorFunc(identificador string, alRecibir func(Event)) *ObservadorFunc {
	return &ObservadorFunc{identificador: identificador, alRecibir: alRecibir}
}

func (o *ObservadorFunc) Actualizar(evento Event) { o.alRecibir(evento) }
func (o *ObservadorFunc) ID() string              { return o.identificador }
