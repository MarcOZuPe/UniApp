package acceso

type hasherDePrueba struct{}

func (hasherDePrueba) Hash(clave string) (string, error) { return "prueba:" + clave, nil }

func (hasherDePrueba) Verify(clave, codificado string) bool { return codificado == "prueba:"+clave }

type observadorDePrueba struct {
	identificador string
	alRecibir     func(Event)
}

func nuevoObservadorDePrueba(identificador string, alRecibir func(Event)) *observadorDePrueba {
	return &observadorDePrueba{identificador: identificador, alRecibir: alRecibir}
}

func (observador *observadorDePrueba) Actualizar(evento Event) { observador.alRecibir(evento) }
func (observador *observadorDePrueba) ID() string              { return observador.identificador }

type busDePrueba struct {
	observadores map[string][]Observador
}

func nuevoBusDePrueba() *busDePrueba {
	return &busDePrueba{observadores: map[string][]Observador{}}
}

func (bus *busDePrueba) Registrar(nombre string, observador Observador) {
	bus.observadores[nombre] = append(bus.observadores[nombre], observador)
}

func (bus *busDePrueba) Desregistrar(nombre string, observador Observador) {
	restantes := make([]Observador, 0, len(bus.observadores[nombre]))
	for _, actual := range bus.observadores[nombre] {
		if actual.ID() != observador.ID() {
			restantes = append(restantes, actual)
		}
	}
	bus.observadores[nombre] = restantes
}

func (bus *busDePrueba) Notificar(evento Event) {
	for _, observador := range append([]Observador(nil), bus.observadores[evento.Nombre]...) {
		observador.Actualizar(evento)
	}
}
