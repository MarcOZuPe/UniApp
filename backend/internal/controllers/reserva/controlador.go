package reserva

import (
	"encoding/json"
	"net/http"

	ctrlAcceso "uniapp/internal/controllers/acceso"
	modeloAcceso "uniapp/internal/models/acceso"
	modelo "uniapp/internal/models/reserva"
	vistaAcceso "uniapp/internal/views/acceso"
	vista "uniapp/internal/views/reserva"
)

// Controlador expone la reserva base. Recibe el decorador de sesión con su tipo sin
// nombre para poder usar el mismo Autenticar del flujo A sin acoplarse a su paquete.
type Controlador struct {
	servicio   *modelo.Servicio
	autenticar func(http.Handler) http.Handler
	errores    *ctrlAcceso.MapeadorDeErrores
}

func Nuevo(servicio *modelo.Servicio, autenticar func(http.Handler) http.Handler) *Controlador {
	errores := ctrlAcceso.NuevoMapeadorDeErrores()
	registrarErroresDeReserva(errores)
	return &Controlador{servicio: servicio, autenticar: autenticar, errores: errores}
}

func (c *Controlador) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/reserva/salud", c.salud)
	mux.Handle("GET /api/reserva/slots", c.autenticar(http.HandlerFunc(c.slots)))
	mux.Handle("GET /api/reserva/{id}", c.autenticar(http.HandlerFunc(c.consultar)))
	mux.Handle("POST /api/reserva", c.autenticar(http.HandlerFunc(c.crear)))
	return mux
}

func (c *Controlador) salud(w http.ResponseWriter, _ *http.Request) {
	vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]string{"flujo": "reserva", "estado": "implementado"})
}

func (c *Controlador) slots(w http.ResponseWriter, r *http.Request) {
	if ctrlAcceso.UsuarioDe(r.Context()) == nil {
		c.responderError(w, modelo.ErrNoAutorizado)
		return
	}
	disposiciones, err := c.servicio.Slots(r.Context())
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]any{"slots": vista.PresentarSlots(disposiciones)})
}

func (c *Controlador) crear(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.responderError(w, modelo.ErrNoAutorizado)
		return
	}
	// Solo el escolar reserva; el apoderado gestiona autorizaciones (I06) y el tutor
	// publica disponibilidad (I03).
	if usuario.Rol != modeloAcceso.RolEscolar {
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusForbidden, Codigo: "solo_escolar",
			Mensaje: "Solo una cuenta de escolar puede crear reservas.",
		})
		return
	}
	var entrada struct {
		SlotID   string `json:"slot_id"`
		Tema     string `json:"tema"`
		Objetivo string `json:"objetivo"`
		Grado    string `json:"grado"`
		Adjuntos []struct {
			Nombre      string `json:"nombre"`
			Tipo        string `json:"tipo"`
			TamanoBytes int64  `json:"tamano_bytes"`
		} `json:"adjuntos"`
	}
	if !decodificarCuerpo(w, r, &entrada) {
		return
	}
	datos := modelo.DatosReserva{SlotID: entrada.SlotID, Tema: entrada.Tema, Objetivo: entrada.Objetivo, Grado: entrada.Grado}
	for _, adjunto := range entrada.Adjuntos {
		datos.Adjuntos = append(datos.Adjuntos, modelo.Adjunto{Nombre: adjunto.Nombre, Tipo: adjunto.Tipo, TamanoBytes: adjunto.TamanoBytes})
	}
	reserva, err := c.servicio.Reservar(r.Context(), usuario.ID, datos)
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusCreated, vista.PresentarReserva(reserva))
}

func (c *Controlador) consultar(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.responderError(w, modelo.ErrNoAutorizado)
		return
	}
	reserva, err := c.servicio.ReservaDe(r.Context(), r.PathValue("id"), usuario.ID)
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, vista.PresentarReserva(reserva))
}

func decodificarCuerpo(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
	decodificador := json.NewDecoder(r.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusBadRequest, Codigo: "solicitud_invalida",
			Mensaje: "El cuerpo de la solicitud no es válido.",
		})
		return false
	}
	return true
}

func (c *Controlador) responderError(w http.ResponseWriter, err error) {
	c.errores.Escribir(w, err)
}
