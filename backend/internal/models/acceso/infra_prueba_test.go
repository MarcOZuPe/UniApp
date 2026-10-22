package acceso

// Implementaciones mínimas para las pruebas del dominio. Los adaptadores reales
// (PBKDF2, bus, correo) viven en internal/infra.

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

func (o *observadorDePrueba) Actualizar(evento Event) { o.alRecibir(evento) }
func (o *observadorDePrueba) ID() string              { return o.identificador }

type busDePrueba struct {
	observadores map[string][]Observador
}

func nuevoBusDePrueba() *busDePrueba {
	return &busDePrueba{observadores: map[string][]Observador{}}
}

func (b *busDePrueba) Registrar(nombre string, observador Observador) {
	b.observadores[nombre] = append(b.observadores[nombre], observador)
}

func (b *busDePrueba) Desregistrar(nombre string, observador Observador) {
	restantes := make([]Observador, 0, len(b.observadores[nombre]))
	for _, actual := range b.observadores[nombre] {
		if actual.ID() != observador.ID() {
			restantes = append(restantes, actual)
		}
	}
	b.observadores[nombre] = restantes
}

func (b *busDePrueba) Notificar(evento Event) {
	for _, observador := range append([]Observador(nil), b.observadores[evento.Nombre]...) {
		observador.Actualizar(evento)
	}
}
