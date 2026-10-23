package tutor

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	ctrlAcceso "uniapp/internal/controllers/acceso"
	modeloAcceso "uniapp/internal/models/acceso"
	modelo "uniapp/internal/models/tutor"
	vistaAcceso "uniapp/internal/views/acceso"
	vista "uniapp/internal/views/tutor"
)

const (
	maximoArchivos  = 5
	maximoSolicitud = maximoArchivos*modelo.MaximoDocumento + (1 << 20)
)

type Controlador struct {
	servicio       *modelo.Servicio
	autenticar     ctrlAcceso.Middleware
	coordinacion   ctrlAcceso.Middleware
	presentarTutor ctrlAcceso.Middleware
}

func Nuevo(servicio *modelo.Servicio, autenticar ctrlAcceso.Middleware) *Controlador {
	return &Controlador{
		servicio:       servicio,
		autenticar:     autenticar,
		coordinacion:   ctrlAcceso.RequierePermiso(modeloAcceso.PermisoCoordinacionAcceder),
		presentarTutor: ctrlAcceso.RequierePermiso(modeloAcceso.PermisoExpedientePresentar),
	}
}

func (c *Controlador) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tutor/salud", func(w http.ResponseWriter, _ *http.Request) {
		vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]string{"flujo": "tutor", "estado": "implementado"})
	})
	mux.HandleFunc("GET /api/tutor/perfiles", c.perfilesPublicos)
	mux.HandleFunc("GET /api/tutor/perfiles/{id}", c.perfilPublico)
	mux.Handle("GET /api/tutor/mi-expediente", c.proteger(http.HandlerFunc(c.miExpediente), c.presentarTutor))
	mux.Handle("POST /api/tutor/expediente", c.proteger(http.HandlerFunc(c.presentarExpediente), c.presentarTutor))
	mux.Handle("GET /api/tutor/revision", c.proteger(http.HandlerFunc(c.pendientes), c.coordinacion))
	mux.Handle("POST /api/tutor/{id}/revision", c.proteger(http.HandlerFunc(c.revisar), c.coordinacion))
	mux.Handle("GET /api/tutor/documentos/{id}", c.proteger(http.HandlerFunc(c.documento), nil))
	return mux
}

func (c *Controlador) proteger(handler http.Handler, permiso ctrlAcceso.Middleware) http.Handler {
	if permiso == nil {
		return c.autenticar(handler)
	}
	return ctrlAcceso.Encadenar(handler, c.autenticar, permiso)
}

func (c *Controlador) miExpediente(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	expediente, err := c.servicio.MiExpediente(r.Context(), usuario.ID)
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, vista.PresentarExpediente(expediente))
}

func (c *Controlador) presentarExpediente(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, maximoSolicitud)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		c.responderError(w, modelo.ErrorValidacion{Campos: map[string]string{"documentos": "La solicitud es inválida o supera el tamaño permitido."}})
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	archivos := r.MultipartForm.File["documentos"]
	if len(archivos) > maximoArchivos {
		c.responderError(w, modelo.ErrorValidacion{Campos: map[string]string{"documentos": "Adjunta como máximo 5 archivos PDF."}})
		return
	}

	datos := modelo.DatosExpediente{
		Presentacion: r.FormValue("presentacion"),
		Cursos:       r.FormValue("cursos"),
		Experiencia:  r.FormValue("experiencia"),
	}
	for _, encabezado := range archivos {
		archivo, err := encabezado.Open()
		if err != nil {
			c.responderError(w, modelo.ErrDocumentoInvalido)
			return
		}
		contenido, readErr := io.ReadAll(io.LimitReader(archivo, modelo.MaximoDocumento+1))
		closeErr := archivo.Close()
		if readErr != nil || closeErr != nil || len(contenido) > modelo.MaximoDocumento {
			c.responderError(w, modelo.ErrorValidacion{Campos: map[string]string{"documentos": "Cada PDF debe pesar como máximo 10 MB."}})
			return
		}
		datos.Documentos = append(datos.Documentos, modelo.Archivo{Nombre: encabezado.Filename, Contenido: contenido})
	}

	expediente, err := c.servicio.Presentar(r.Context(), usuario.ID, datos)
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusCreated, vista.PresentarExpediente(expediente))
}

func (c *Controlador) pendientes(w http.ResponseWriter, r *http.Request) {
	expedientes, err := c.servicio.Pendientes(r.Context())
	if err != nil {
		c.responderError(w, err)
		return
	}
	resultado := make([]vista.Expediente, 0, len(expedientes))
	for indice := range expedientes {
		resultado = append(resultado, vista.PresentarExpediente(&expedientes[indice]))
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]any{"expedientes": resultado})
}

func (c *Controlador) revisar(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	var entrada struct {
		Estado  modelo.Estado `json:"estado"`
		Motivo  string        `json:"motivo_rechazo"`
		VenceEn time.Time     `json:"vence_en"`
	}
	if !decodificarJSON(w, r, &entrada) {
		return
	}
	expediente, err := c.servicio.Revisar(r.Context(), r.PathValue("id"), usuario.ID, entrada.Estado, entrada.Motivo, entrada.VenceEn)
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, vista.PresentarExpediente(expediente))
}

func (c *Controlador) perfilPublico(w http.ResponseWriter, r *http.Request) {
	expediente, err := c.servicio.PerfilPublico(r.Context(), r.PathValue("id"))
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, vista.PresentarPerfilPublico(expediente))
}

func (c *Controlador) perfilesPublicos(w http.ResponseWriter, r *http.Request) {
	perfiles, err := c.servicio.PerfilesPublicos(r.Context())
	if err != nil {
		c.responderError(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]any{
		"perfiles": vista.PresentarPerfilesPublicos(perfiles),
	})
}

func (c *Controlador) documento(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	contenido, documento, err := c.servicio.DescargarDocumento(
		r.Context(), r.PathValue("id"), usuario.ID,
		usuario.Rol.TienePermiso(modeloAcceso.PermisoCoordinacionAcceder),
	)
	if err != nil {
		c.responderError(w, err)
		return
	}
	disposicion := mime.FormatMediaType("attachment", map[string]string{"filename": nombreDescarga(documento.Nombre)})
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", disposicion)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(contenido)
}

func nombreDescarga(nombre string) string {
	nombre = strings.ReplaceAll(nombre, `\`, "/")
	partes := strings.Split(nombre, "/")
	return partes[len(partes)-1]
}

func decodificarJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
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
	var validacion modelo.ErrorValidacion
	switch {
	case errors.As(err, &validacion):
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusUnprocessableEntity, Codigo: "validacion",
			Mensaje: "Revisa los datos del expediente.", Campos: validacion.Campos,
		})
	case errors.Is(err, modelo.ErrNoEncontrado), errors.Is(err, modelo.ErrNoPublicable):
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusNotFound, Codigo: "no_encontrado", Mensaje: "No se encontró el recurso solicitado.",
		})
	case errors.Is(err, modelo.ErrSinPermiso):
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusForbidden, Codigo: "sin_permiso", Mensaje: "No tienes acceso a este documento.",
		})
	case errors.Is(err, modelo.ErrEstadoInvalido):
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusConflict, Codigo: "estado_invalido", Mensaje: "El expediente ya fue revisado.",
		})
	case errors.Is(err, modelo.ErrExpedienteInvalido), errors.Is(err, modelo.ErrDocumentoInvalido):
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusUnprocessableEntity, Codigo: "datos_invalidos", Mensaje: err.Error(),
		})
	default:
		log.Printf("error del flujo de tutor: %v", err)
		vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
			Estado: http.StatusInternalServerError, Codigo: "error_interno", Mensaje: "Ocurrió un error inesperado.",
		})
	}
}
