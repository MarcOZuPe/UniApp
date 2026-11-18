package incidente

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vistaAcceso "uniapp/internal/adapters/presentation/acceso"
	vista "uniapp/internal/adapters/presentation/incidente"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/incidente"
)

const limiteSolicitud = modelo.MaxEvidencias*modelo.EvidenciaMaximaBytes + (1 << 20)

type Controlador struct {
	servicio   *modelo.Servicio
	autenticar ctrlAcceso.Middleware
	errores    *ctrlAcceso.MapeadorDeErrores
}

func Nuevo(servicio *modelo.Servicio, autenticar ctrlAcceso.Middleware) *Controlador {
	errores := ctrlAcceso.NuevoMapeadorDeErrores()
	registrarErroresDeIncidente(errores)
	return &Controlador{servicio: servicio, autenticar: autenticar, errores: errores}
}

func (c *Controlador) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/incidentes/salud", func(w http.ResponseWriter, _ *http.Request) {
		vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]string{"flujo": "incidente", "estado": "implementado"})
	})
	mux.Handle("GET /api/incidentes/tutorias", c.proteger(c.tutorias, modeloAcceso.PermisoIncidenteAcceder))
	mux.Handle("GET /api/incidentes", c.proteger(c.listar, modeloAcceso.PermisoIncidenteAcceder))
	mux.Handle("POST /api/incidentes", c.proteger(c.registrar, modeloAcceso.PermisoIncidenteRegistrar))
	mux.Handle("GET /api/incidentes/{id}", c.proteger(c.detalle, modeloAcceso.PermisoIncidenteAcceder))
	mux.Handle("GET /api/incidentes/{id}/evidencias/{evidenciaId}", c.proteger(c.descargarEvidencia, modeloAcceso.PermisoIncidenteAcceder))
	mux.Handle("POST /api/incidentes/{id}/revision", c.proteger(c.revisar, modeloAcceso.PermisoIncidenteRevisar))
	return mux
}

func (c *Controlador) proteger(manejador http.HandlerFunc, permiso modeloAcceso.Permiso) http.Handler {
	return c.autenticar(ctrlAcceso.RequierePermiso(permiso)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Los reportes y sus metadatos no deben quedar en cachés intermedias.
		w.Header().Set("Cache-Control", "no-store")
		manejador(w, r)
	})))
}

func (c *Controlador) tutorias(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, modelo.ErrNoAutorizado)
		return
	}
	tutorias, err := c.servicio.Tutorias(r.Context(), usuario.ID, usuario.Rol)
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]any{
		"tutorias": vista.PresentarTutorias(tutorias),
	})
}

func (c *Controlador) listar(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, modelo.ErrNoAutorizado)
		return
	}
	incidentes, err := c.servicio.Listar(r.Context(), usuario.ID, usuario.Rol)
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]any{
		"incidentes": vista.PresentarIncidentes(incidentes),
	})
}

func (c *Controlador) detalle(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, modelo.ErrNoAutorizado)
		return
	}
	incidente, err := c.servicio.PorID(r.Context(), usuario.ID, usuario.Rol, r.PathValue("id"))
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, vista.PresentarIncidente(incidente))
}

func (c *Controlador) registrar(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, modelo.ErrNoAutorizado)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, limiteSolicitud)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		var demasiadoGrande *http.MaxBytesError
		if errors.As(err, &demasiadoGrande) {
			vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
				Estado: http.StatusRequestEntityTooLarge, Codigo: "solicitud_grande",
				Mensaje: "La solicitud supera el tamaño permitido.",
			})
			return
		}
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusBadRequest, Codigo: "multipart_invalido",
			Mensaje: "Adjunta las evidencias como archivos multipart.",
		})
		return
	}
	defer func() {
		if r.MultipartForm != nil {
			_ = r.MultipartForm.RemoveAll()
		}
	}()

	archivos := r.MultipartForm.File["evidencias"]
	evidencias := make([]modelo.Archivo, 0, len(archivos))
	for _, cabecera := range archivos {
		archivo, err := cabecera.Open()
		if err != nil {
			vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
				Estado: http.StatusBadRequest, Codigo: "evidencia_ilegible",
				Mensaje: "No se pudo leer una de las evidencias.",
			})
			return
		}
		contenido, err := io.ReadAll(archivo)
		_ = archivo.Close()
		if err != nil {
			vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
				Estado: http.StatusBadRequest, Codigo: "evidencia_ilegible",
				Mensaje: "No se pudo leer una de las evidencias.",
			})
			return
		}
		evidencias = append(evidencias, modelo.Archivo{Nombre: cabecera.Filename, Contenido: contenido})
	}

	incidente, err := c.servicio.Registrar(r.Context(), usuario.ID, usuario.Rol, modelo.DatosIncidente{
		EscolarID:   r.FormValue("escolar_id"),
		ReservaID:   r.FormValue("reserva_id"),
		Tipo:        modelo.Tipo(r.FormValue("tipo")),
		Descripcion: r.FormValue("descripcion"),
		Evidencias:  evidencias,
	})
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusCreated, vista.PresentarIncidente(incidente))
}

func (c *Controlador) revisar(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, modelo.ErrNoAutorizado)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var solicitud struct {
		Resultado  modelo.Resultado `json:"resultado"`
		Comentario string           `json:"comentario"`
	}
	decodificador := json.NewDecoder(r.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(&solicitud); err != nil {
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusBadRequest, Codigo: "json_invalido",
			Mensaje: "El cuerpo de la revisión no es válido.",
		})
		return
	}
	if err := asegurarFinDeJSON(decodificador); err != nil {
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusBadRequest, Codigo: "json_invalido",
			Mensaje: "El cuerpo de la revisión no es válido.",
		})
		return
	}
	incidente, err := c.servicio.Revisar(
		r.Context(), usuario.ID, usuario.Rol, r.PathValue("id"),
		modelo.DatosRevision{Resultado: solicitud.Resultado, Comentario: solicitud.Comentario},
	)
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, vista.PresentarIncidente(incidente))
}

func (c *Controlador) descargarEvidencia(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, modelo.ErrNoAutorizado)
		return
	}
	_, evidencia, contenido, err := c.servicio.AbrirEvidencia(
		r.Context(), usuario.ID, usuario.Rol,
		r.PathValue("id"), r.PathValue("evidenciaId"),
	)
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	disposicion := mime.FormatMediaType("attachment", map[string]string{"filename": evidencia.Nombre})
	w.Header().Set("Content-Type", evidencia.Tipo)
	w.Header().Set("Content-Disposition", disposicion)
	w.Header().Set("Content-Length", strconv.FormatInt(evidencia.TamanoBytes, 10))
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(contenido)
}

func asegurarFinDeJSON(decodificador *json.Decoder) error {
	var extra any
	if err := decodificador.Decode(&extra); err != io.EOF {
		return errors.New("contenido extra")
	}
	return nil
}
