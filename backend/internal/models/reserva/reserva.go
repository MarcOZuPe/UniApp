package reserva

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// Reglas de la reserva base (I04). Viven en el modelo, no en el controlador ni en la
// vista: RN-03 (48 horas), RN-04 (UTC y America/Lima), RN-07 (datos de preparación) y
// D06 (sesiones de 60 minutos, horizonte de 28 días).
const (
	DuracionSesion     = 60 * time.Minute
	AnticipacionMinima = 48 * time.Hour
	HorizonteMaximo    = 28 * 24 * time.Hour

	TemaMinimo, TemaMaximo         = 10, 200
	ObjetivoMinimo, ObjetivoMaximo = 20, 1000
	GradoMaximo                    = 80
	MaxAdjuntos                    = 5
	AdjuntoMaximoBytes             = 10 << 20
	NombreAdjuntoMaximo            = 255
)

// ZonaLima es una zona fija de -05:00: el Perú no aplica horario de verano (RN-04).
// Se evita depender de la base de zonas horarias del sistema operativo.
var ZonaLima = time.FixedZone("America/Lima", -5*60*60)

type Estado string

const EstadoConfirmada Estado = "confirmada"

// Slot es un horario publicado por el flujo C (oferta). En I04 se consume desde una
// oferta sintética; cuando exista la API de Mateo se reemplaza solo el puerto.
type Slot struct {
	ID          string
	TutorID     string
	TutorNombre string
	Curso       string
	Inicio      time.Time
	Fin         time.Time
}

// Disposicion indica si un slot puede reservarse y, si no, por qué. La pantalla la usa
// para deshabilitar opciones sin duplicar reglas en la vista.
type Disposicion struct {
	Slot   Slot
	Motivo string
}

func (d Disposicion) Disponible() bool { return d.Motivo == "" }

type Adjunto struct {
	Nombre      string
	Tipo        string
	TamanoBytes int64
}

type Reserva struct {
	ID          string
	EscolarID   string
	TutorID     string
	TutorNombre string
	SlotID      string
	Curso       string
	Grado       string
	Inicio      time.Time
	Fin         time.Time
	Tema        string
	Objetivo    string
	Estado      Estado
	CreadaEn    time.Time
	Adjuntos    []Adjunto
}

// DatosReserva es lo que envía el formulario. El escolar nunca se recibe del cliente:
// sale de la sesión autenticada.
type DatosReserva struct {
	SlotID   string
	Tema     string
	Objetivo string
	Grado    string
	Adjuntos []Adjunto
}

var (
	ErrSlotInexistente     = errors.New("el horario elegido no existe")
	ErrSlotVencido         = errors.New("el horario elegido ya pasó")
	ErrSlotOcupado         = errors.New("el horario elegido ya fue reservado")
	ErrReservaNoEncontrada = errors.New("reserva no encontrada")
	ErrNoAutorizado        = errors.New("no autorizado")
	ErrSinAutorizacion     = errors.New("el escolar no tiene autorización vigente para reservar")
	ErrAnticipacion        = errors.New("la sesión debe empezar al menos 48 horas después de este momento")
	ErrHorizonte           = errors.New("el horario está fuera del horizonte de 28 días")
	ErrReservaInvalida     = errors.New("reserva inválida")
)

// ErrValidacion agrupa los errores por campo para que la vista los muestre sin repetir
// las reglas. Es propio del flujo D para no acoplar la reserva al modelo de acceso.
type ErrValidacion struct{ Campos map[string]string }

func (e ErrValidacion) Error() string { return "datos inválidos" }

// MotivoPorTiempo devuelve "" si el slot cumple la ventana de reserva, o el motivo por
// el que no puede reservarse todavía. Es la única fuente de RN-03 en el proyecto.
func MotivoPorTiempo(slot Slot, ahora time.Time) string {
	if err := validarVentana(slot, ahora); err != nil {
		return err.Error()
	}
	return ""
}

func validarVentana(slot Slot, ahora time.Time) error {
	ahora = ahora.UTC()
	switch {
	case !slot.Fin.After(ahora):
		return ErrSlotVencido
	case slot.Inicio.Sub(ahora) < AnticipacionMinima:
		return ErrAnticipacion
	case slot.Inicio.Sub(ahora) > HorizonteMaximo:
		return ErrHorizonte
	}
	return nil
}

func normalizarDatos(datos DatosReserva) DatosReserva {
	datos.SlotID = strings.TrimSpace(datos.SlotID)
	datos.Tema = strings.TrimSpace(datos.Tema)
	datos.Objetivo = strings.TrimSpace(datos.Objetivo)
	datos.Grado = strings.TrimSpace(datos.Grado)
	adjuntos := make([]Adjunto, len(datos.Adjuntos))
	for i, adjunto := range datos.Adjuntos {
		adjunto.Nombre = strings.TrimSpace(adjunto.Nombre)
		adjunto.Tipo = strings.TrimSpace(adjunto.Tipo)
		if adjunto.Tipo == "" {
			adjunto.Tipo = "application/octet-stream"
		}
		adjuntos[i] = adjunto
	}
	if len(adjuntos) == 0 {
		adjuntos = nil
	}
	datos.Adjuntos = adjuntos
	return datos
}

func validarDatos(datos DatosReserva) map[string]string {
	campos := map[string]string{}
	if datos.SlotID == "" {
		campos["slot_id"] = "elige un horario disponible"
	}
	if longitud := utf8.RuneCountInString(datos.Tema); longitud < TemaMinimo || longitud > TemaMaximo {
		campos["tema"] = "el tema debe tener entre 10 y 200 caracteres"
	}
	if longitud := utf8.RuneCountInString(datos.Objetivo); longitud < ObjetivoMinimo || longitud > ObjetivoMaximo {
		campos["objetivo"] = "el objetivo debe tener entre 20 y 1000 caracteres"
	}
	if longitud := utf8.RuneCountInString(datos.Grado); longitud < 1 || longitud > GradoMaximo {
		campos["grado"] = "indica tu grado (por ejemplo, 5to de secundaria)"
	}
	if len(datos.Adjuntos) > MaxAdjuntos {
		campos["adjuntos"] = "puedes adjuntar como máximo 5 archivos"
	}
	for _, adjunto := range datos.Adjuntos {
		if mensaje := validarAdjunto(adjunto); mensaje != "" {
			campos["adjuntos"] = mensaje
			break
		}
	}
	return campos
}

func validarAdjunto(adjunto Adjunto) string {
	nombre := strings.TrimSpace(adjunto.Nombre)
	if nombre == "" {
		return "cada adjunto necesita un nombre"
	}
	if utf8.RuneCountInString(nombre) > NombreAdjuntoMaximo {
		return "el nombre del adjunto es demasiado largo"
	}
	if strings.ContainsAny(nombre, "/\\") {
		return "el nombre del adjunto no puede contener rutas"
	}
	if adjunto.TamanoBytes <= 0 {
		return "el adjunto está vacío"
	}
	if adjunto.TamanoBytes > AdjuntoMaximoBytes {
		return "cada adjunto debe pesar como máximo 10 MB"
	}
	return ""
}

// NuevaReserva aplica RN-03, RN-07 y D05 (la reserva válida se confirma de inmediato).
func NuevaReserva(escolarID string, datos DatosReserva, slot Slot, ahora time.Time) (*Reserva, error) {
	if strings.TrimSpace(escolarID) == "" {
		return nil, ErrNoAutorizado
	}
	datos = normalizarDatos(datos)
	if campos := validarDatos(datos); len(campos) > 0 {
		return nil, ErrValidacion{Campos: campos}
	}
	if slot.ID == "" || slot.TutorID == "" || !slot.Fin.After(slot.Inicio) {
		return nil, ErrSlotInexistente
	}
	if err := validarVentana(slot, ahora); err != nil {
		return nil, err
	}
	return &Reserva{
		EscolarID:   escolarID,
		TutorID:     slot.TutorID,
		TutorNombre: slot.TutorNombre,
		SlotID:      slot.ID,
		Curso:       slot.Curso,
		Grado:       datos.Grado,
		Inicio:      slot.Inicio.UTC(),
		Fin:         slot.Fin.UTC(),
		Tema:        datos.Tema,
		Objetivo:    datos.Objetivo,
		Estado:      EstadoConfirmada,
		CreadaEn:    ahora.UTC(),
		Adjuntos:    datos.Adjuntos,
	}, nil
}
