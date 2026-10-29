package oferta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vista "uniapp/internal/adapters/presentation/oferta"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/oferta"
)

type repositorio interface {
	ObtenerDisponibilidad(ctx context.Context, tutorID string) (modelo.CupoSemanal, []modelo.SlotSemanal, error)
	AgregarSlots(ctx context.Context, tutorID string, slots []modelo.SlotSemanal) (modelo.CupoSemanal, []modelo.SlotSemanal, error)
	EliminarSlot(ctx context.Context, tutorID string, slotID int64) error
	CambiarCupo(ctx context.Context, tutorID string, minutos int) error
}

type Controlador struct {
	repositorio repositorio
}

type entradaSlot struct {
	DiaSemana int    `json:"dia_semana"`
	Inicio    string `json:"inicio"`
	Fin       string `json:"fin"`
}

type entradaSlots struct {
	Slots []entradaSlot `json:"slots"`
}

type entradaCupo struct {
	Minutos int `json:"minutos"`
}

func Rutas(repositorio repositorio, autenticar func(http.Handler) http.Handler) http.Handler {
	controlador := &Controlador{repositorio: repositorio}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/oferta/salud", controlador.salud)
	protegido := http.NewServeMux()
	protegido.HandleFunc("GET /api/oferta/slots", controlador.obtenerDisponibilidad)
	protegido.HandleFunc("POST /api/oferta/slots", controlador.agregarSlots)
	protegido.HandleFunc("DELETE /api/oferta/slots/{id}", controlador.eliminarSlot)
	protegido.HandleFunc("PUT /api/oferta/cupo", controlador.cambiarCupo)
	mux.Handle("/api/oferta/", autenticar(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		usuario := ctrlAcceso.UsuarioDe(r.Context())
		if usuario == nil || usuario.Rol != modeloAcceso.RolTutor {
			vista.EscribirError(w, http.StatusForbidden, "rol_no_permitido", "Solo un tutor puede gestionar su disponibilidad.", nil)
			return
		}
		protegido.ServeHTTP(w, r)
	})))
	return mux
}

func (c *Controlador) salud(w http.ResponseWriter, _ *http.Request) {
	vista.EscribirJSON(w, http.StatusOK, map[string]string{"flujo": "oferta", "estado": "ok"})
}

func (c *Controlador) obtenerDisponibilidad(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	cupo, slots, err := c.repositorio.ObtenerDisponibilidad(r.Context(), usuario.ID)
	if err != nil {
		responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusOK, vista.PresentarDisponibilidad(cupo, slots))
}

func (c *Controlador) agregarSlots(w http.ResponseWriter, r *http.Request) {
	var entrada entradaSlots
	if !decodificarCuerpo(w, r, &entrada) {
		return
	}
	if len(entrada.Slots) == 0 {
		vista.EscribirError(w, http.StatusBadRequest, "solicitud_invalida", "Debes enviar al menos un slot.", map[string]string{"slots": "requerido"})
		return
	}
	slots := make([]modelo.SlotSemanal, 0, len(entrada.Slots))
	for i, entrada := range entrada.Slots {
		inicio, errInicio := parsearHora(entrada.Inicio)
		fin, errFin := parsearHora(entrada.Fin)
		if entrada.DiaSemana < int(time.Sunday) || entrada.DiaSemana > int(time.Saturday) || errInicio != nil || errFin != nil {
			vista.EscribirError(w, http.StatusUnprocessableEntity, "validacion", "Revisa los slots ingresados.", map[string]string{
				fmt.Sprintf("slots.%d", i): "día u hora inválidos; usa dia_semana de 0 a 6 y horas HH:MM",
			})
			return
		}
		slots = append(slots, modelo.SlotSemanal{
			DiaSemana: time.Weekday(entrada.DiaSemana),
			Inicio:    inicio,
			Fin:       fin,
		})
	}
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	cupo, creados, err := c.repositorio.AgregarSlots(r.Context(), usuario.ID, slots)
	if err != nil {
		responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusCreated, vista.PresentarDisponibilidad(cupo, creados))
}

func (c *Controlador) eliminarSlot(w http.ResponseWriter, r *http.Request) {
	slotID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || slotID <= 0 {
		vista.EscribirError(w, http.StatusBadRequest, "solicitud_invalida", "El identificador del slot no es válido.", nil)
		return
	}
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if err := c.repositorio.EliminarSlot(r.Context(), usuario.ID, slotID); err != nil {
		responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusNoContent, nil)
}

func (c *Controlador) cambiarCupo(w http.ResponseWriter, r *http.Request) {
	var entrada entradaCupo
	if !decodificarCuerpo(w, r, &entrada) {
		return
	}
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if err := c.repositorio.CambiarCupo(r.Context(), usuario.ID, entrada.Minutos); err != nil {
		responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusOK, map[string]int{"cupo_semanal_minutos": entrada.Minutos})
}

func decodificarCuerpo(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decodificador := json.NewDecoder(r.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		vista.EscribirError(w, http.StatusBadRequest, "solicitud_invalida", "El cuerpo de la solicitud no es válido.", nil)
		return false
	}
	var adicional any
	if err := decodificador.Decode(&adicional); !errors.Is(err, io.EOF) {
		vista.EscribirError(w, http.StatusBadRequest, "solicitud_invalida", "El cuerpo debe contener un único objeto JSON.", nil)
		return false
	}
	return true
}

func parsearHora(valor string) (time.Duration, error) {
	if strings.TrimSpace(valor) != valor {
		return 0, modelo.ErrSlotInvalido
	}
	if valor == "24:00" {
		return 24 * time.Hour, nil
	}
	momento, err := time.Parse("15:04", valor)
	if err != nil {
		return 0, modelo.ErrSlotInvalido
	}
	return time.Duration(momento.Hour())*time.Hour + time.Duration(momento.Minute())*time.Minute, nil
}

func responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, modelo.ErrSlotInvalido), errors.Is(err, modelo.ErrCupoInvalido):
		vista.EscribirError(w, http.StatusUnprocessableEntity, "validacion", err.Error(), nil)
	case errors.Is(err, modelo.ErrSolapamiento):
		vista.EscribirError(w, http.StatusConflict, "solapamiento", err.Error(), nil)
	case errors.Is(err, modelo.ErrCapacidadSemanal):
		vista.EscribirError(w, http.StatusConflict, "cupo_excedido", err.Error(), nil)
	case errors.Is(err, modelo.ErrReduccionConActivos):
		vista.EscribirError(w, http.StatusConflict, "compromisos_activos", err.Error(), nil)
	case errors.Is(err, modelo.ErrReduccionNoVerificable):
		vista.EscribirError(w, http.StatusServiceUnavailable, "reduccion_no_verificable", err.Error(), nil)
	case errors.Is(err, modelo.ErrSlotNoEncontrado):
		vista.EscribirError(w, http.StatusNotFound, "slot_no_encontrado", err.Error(), nil)
	default:
		log.Printf("error en oferta: %v", err)
		vista.EscribirError(w, http.StatusInternalServerError, "error_interno", "Ocurrió un error inesperado.", nil)
	}
}
