package oferta

import (
	"errors"
	"fmt"
	"time"
	_ "time/tzdata"
)

const (
	ZonaHoraria       = "America/Lima"
	DuracionSlot      = time.Hour
	CupoSemanalMinimo = 180
	CupoSemanalMaximo = 360
	MinutosEnUnaHora  = 60
)

var (
	ErrSlotInvalido           = errors.New("slot de disponibilidad inválido")
	ErrSolapamiento           = errors.New("los slots de un tutor no pueden solaparse")
	ErrCupoInvalido           = errors.New("el cupo semanal debe estar entre 180 y 360 minutos, en múltiplos de 60")
	ErrReduccionConActivos    = errors.New("no se puede reducir el cupo semanal con compromisos activos")
	ErrReduccionNoVerificable = errors.New("no se puede verificar si hay compromisos activos para reducir el cupo")
	ErrCapacidadSemanal       = errors.New("la disponibilidad excede el cupo semanal")
	ErrSlotNoEncontrado       = errors.New("slot de disponibilidad no encontrado")
)

type SlotSemanal struct {
	ID        int64
	TutorID   string
	DiaSemana time.Weekday
	Inicio    time.Duration
	Fin       time.Duration
}

type CupoSemanal struct {
	TutorID string
	Minutos int
}

func (s SlotSemanal) Validar() error {
	if s.TutorID == "" || s.DiaSemana < time.Sunday || s.DiaSemana > time.Saturday ||
		s.Inicio < 0 || s.Inicio >= 24*time.Hour ||
		s.Fin <= 0 || s.Fin > 24*time.Hour ||
		s.Inicio%time.Minute != 0 || s.Fin%time.Minute != 0 ||
		s.Fin-s.Inicio != DuracionSlot {
		return ErrSlotInvalido
	}
	return nil
}

func ValidarSlots(slots []SlotSemanal) error {
	for i, slot := range slots {
		if err := slot.Validar(); err != nil {
			return fmt.Errorf("slot %d: %w", i, err)
		}
		for j := 0; j < i; j++ {
			anterior := slots[j]
			if slot.TutorID == anterior.TutorID &&
				slot.DiaSemana == anterior.DiaSemana &&
				slot.Inicio < anterior.Fin && anterior.Inicio < slot.Fin {
				return ErrSolapamiento
			}
		}
	}
	return nil
}

func (c CupoSemanal) Validar() error {
	if c.TutorID == "" || c.Minutos < CupoSemanalMinimo ||
		c.Minutos > CupoSemanalMaximo || c.Minutos%MinutosEnUnaHora != 0 {
		return ErrCupoInvalido
	}
	return nil
}

func ValidarCambioCupoSemanal(actual, nuevo CupoSemanal, compromisosActivos int) error {
	if err := actual.Validar(); err != nil {
		return err
	}
	if err := nuevo.Validar(); err != nil {
		return err
	}
	if actual.TutorID != nuevo.TutorID || compromisosActivos < 0 {
		return ErrCupoInvalido
	}
	if nuevo.Minutos < actual.Minutos && compromisosActivos > 0 {
		return ErrReduccionConActivos
	}
	return nil
}

func EnZonaHorariaLima(momento time.Time) (time.Time, error) {
	zona, err := time.LoadLocation(ZonaHoraria)
	if err != nil {
		return time.Time{}, fmt.Errorf("cargar zona horaria %s: %w", ZonaHoraria, err)
	}
	return momento.In(zona), nil
}

func AhoraEnLima() (time.Time, error) {
	return EnZonaHorariaLima(time.Now())
}

func MinutosDelDia(hora, minuto int) (time.Duration, error) {
	if hora < 0 || hora >= 24 || minuto < 0 || minuto >= MinutosEnUnaHora {
		return 0, ErrSlotInvalido
	}
	minutos := hora*MinutosEnUnaHora + minuto
	return time.Duration(minutos) * time.Minute, nil
}

func HorarioDesdeMinutos(minutos int) time.Duration {
	return time.Duration(minutos) * time.Minute
}
