package sesion

import (
	"errors"
	"time"
)

// Politica temporal de I11. Las declaraciones y el resultado solo se aceptan
// cerca de la sesion real, para que reflejen lo ocurrido y no una intencion.
const (
	// AnticipacionDeclaracion permite declarar poco antes del inicio
	// (por ejemplo, al entrar a la sala).
	AnticipacionDeclaracion = 15 * time.Minute
	// PlazoDeclaracion es el margen posterior al fin para declarar
	// asistencia o reportar ausencia.
	PlazoDeclaracion = 24 * time.Hour
	// PlazoResultado es el margen posterior al fin para registrar el resultado.
	PlazoResultado = 7 * 24 * time.Hour
)

var ErrFueraDePlazo = errors.New("la accion esta fuera del plazo permitido para esta sesion")

// Intervalo es un value object inmutable [desde, hasta], ambos inclusive.
type Intervalo struct {
	desde time.Time
	hasta time.Time
}

func (i Intervalo) Desde() time.Time { return i.desde }
func (i Intervalo) Hasta() time.Time { return i.hasta }

func (i Intervalo) Contiene(momento time.Time) bool {
	return !momento.Before(i.desde) && !momento.After(i.hasta)
}

// VentanaDeclaracion delimita cuando un participante puede declarar su
// asistencia o reportar su ausencia.
func (s *Sesion) VentanaDeclaracion() Intervalo {
	return Intervalo{
		desde: s.inicio.Add(-AnticipacionDeclaracion),
		hasta: s.fin.Add(PlazoDeclaracion),
	}
}

// VentanaResultado delimita cuando el tutor puede registrar el resultado:
// nunca antes de que la sesion haya comenzado.
func (s *Sesion) VentanaResultado() Intervalo {
	return Intervalo{
		desde: s.inicio,
		hasta: s.fin.Add(PlazoResultado),
	}
}

func exigirDentroDe(ventana Intervalo, momento time.Time) error {
	if !ventana.Contiene(momento) {
		return ErrFueraDePlazo
	}
	return nil
}
