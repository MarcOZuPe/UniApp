package catalogo

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	ctrlAcceso "uniapp/internal/controllers/acceso"
	modeloAcceso "uniapp/internal/models/acceso"
	modelo "uniapp/internal/models/catalogo"
	vista "uniapp/internal/views/catalogo"
)

const maximoSolicitud = 4 << 20

type Controlador struct {
	servicio     *modelo.Servicio
	autenticar   ctrlAcceso.Middleware
	coordinacion ctrlAcceso.Middleware
	solicitar    ctrlAcceso.Middleware
}

type entradaImportacion struct {
	Fuente   string           `json:"fuente"`
	Cursos   []entradaCurso   `json:"cursos"`
	Colegios []entradaColegio `json:"colegios"`
}

type entradaCurso struct {
	Nombre string `json:"nombre"`
	Grado  string `json:"grado"`
}

type entradaColegio struct {
	Nombre   string `json:"nombre"`
	Distrito string `json:"distrito"`
}

type entradaRevision struct {
	Estado modelo.EstadoCurso `json:"estado"`
	Motivo string             `json:"motivo_rechazo"`
}

type entradaSolicitud struct {
	Nombre   string `json:"nombre"`
	Distrito string `json:"distrito"`
}

func Nuevo(servicio *modelo.Servicio, autenticar ctrlAcceso.Middleware) *Controlador {
	return &Controlador{
		servicio:     servicio,
		autenticar:   autenticar,
		coordinacion: ctrlAcceso.RequierePermiso(modeloAcceso.PermisoCatalogoGestionar),
		solicitar:    ctrlAcceso.RequierePermiso(modeloAcceso.PermisoCatalogoSolicitarColegio),
	}
}

func (c *Controlador) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/catalogo", c.catalogoPublico)
	mux.Handle("POST /api/catalogo/importaciones", c.proteger(http.HandlerFunc(c.importar), c.coordinacion))
	mux.Handle("GET /api/catalogo/cursos/pendientes", c.proteger(http.HandlerFunc(c.cursosPendientes), c.coordinacion))
	mux.Handle("POST /api/catalogo/cursos/{id}/revision", c.proteger(http.HandlerFunc(c.revisarCurso), c.coordinacion))
	mux.Handle("POST /api/catalogo/colegios/solicitudes", c.proteger(http.HandlerFunc(c.solicitarColegio), c.solicitar))
	mux.Handle("GET /api/catalogo/colegios/solicitudes", c.proteger(http.HandlerFunc(c.solicitudesPendientes), c.coordinacion))
	return mux
}

func (c *Controlador) proteger(handler http.Handler, permiso ctrlAcceso.Middleware) http.Handler {
	return ctrlAcceso.Encadenar(handler, c.autenticar, permiso)
}

func (c *Controlador) catalogoPublico(w http.ResponseWriter, r *http.Request) {
	importacion, err := c.servicio.CatalogoPublico(r.Context())
	if err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusOK, vista.PresentarImportacion(importacion))
}

func (c *Controlador) importar(w http.ResponseWriter, r *http.Request) {
	var entrada entradaImportacion
	if !decodificarJSON(w, r, &entrada) {
		return
	}
	cursos := make([]modelo.CursoEntrada, 0, len(entrada.Cursos))
	for _, curso := range entrada.Cursos {
		cursos = append(cursos, modelo.CursoEntrada{Nombre: curso.Nombre, Grado: curso.Grado})
	}
	colegios := make([]modelo.ColegioEntrada, 0, len(entrada.Colegios))
	for _, colegio := range entrada.Colegios {
		colegios = append(colegios, modelo.ColegioEntrada{Nombre: colegio.Nombre, Distrito: colegio.Distrito})
	}
	importacion, err := c.servicio.Importar(r.Context(), modelo.DatosImportacion{
		Fuente: entrada.Fuente, Cursos: cursos, Colegios: colegios,
	})
	if err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusCreated, vista.PresentarImportacion(importacion))
}

func (c *Controlador) cursosPendientes(w http.ResponseWriter, r *http.Request) {
	cursos, err := c.servicio.CursosPendientes(r.Context())
	if err != nil {
		c.responderError(w, err)
		return
	}
	resultado := make([]vista.Curso, 0, len(cursos))
	for _, curso := range cursos {
		resultado = append(resultado, vista.Curso{
			ID: curso.ID, Nombre: curso.Nombre, Grado: curso.Grado, Estado: string(curso.Estado),
		})
	}
	vista.EscribirJSON(w, http.StatusOK, map[string]any{"cursos": resultado})
}

func (c *Controlador) revisarCurso(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		vista.EscribirError(w, http.StatusBadRequest, "solicitud_invalida", "El identificador del curso no es válido.", nil)
		return
	}
	var entrada entradaRevision
	if !decodificarJSON(w, r, &entrada) {
		return
	}
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	curso, err := c.servicio.RevisarCurso(r.Context(), id, usuario.ID, entrada.Estado, entrada.Motivo)
	if err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusOK, vista.Curso{
		ID: curso.ID, Nombre: curso.Nombre, Grado: curso.Grado, Estado: string(curso.Estado),
		Motivo: curso.Motivo,
	})
}

func (c *Controlador) solicitarColegio(w http.ResponseWriter, r *http.Request) {
	var entrada entradaSolicitud
	if !decodificarJSON(w, r, &entrada) {
		return
	}
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	solicitud, err := c.servicio.SolicitarColegio(r.Context(), usuario.ID, modelo.SolicitudColegioEntrada{
		Nombre: entrada.Nombre, Distrito: entrada.Distrito,
	})
	if err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusCreated, vista.SolicitudColegio{
		ID: solicitud.ID, Nombre: solicitud.Nombre, Distrito: solicitud.Distrito,
		SolicitadoEn: solicitud.SolicitadoEn,
	})
}

func (c *Controlador) solicitudesPendientes(w http.ResponseWriter, r *http.Request) {
	solicitudes, err := c.servicio.SolicitudesPendientes(r.Context())
	if err != nil {
		c.responderError(w, err)
		return
	}
	vista.EscribirJSON(w, http.StatusOK, map[string]any{
		"solicitudes": vista.PresentarSolicitudes(solicitudes),
	})
}

func decodificarJSON(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maximoSolicitud)
	decodificador := json.NewDecoder(r.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		var limite *http.MaxBytesError
		if errors.As(err, &limite) {
			vista.EscribirError(w, http.StatusRequestEntityTooLarge, "solicitud_grande", "La solicitud supera el tamaño permitido.", nil)
		} else {
			vista.EscribirError(w, http.StatusBadRequest, "solicitud_invalida", "El cuerpo de la solicitud no es válido.", nil)
		}
		return false
	}
	var adicional any
	if err := decodificador.Decode(&adicional); !errors.Is(err, io.EOF) {
		vista.EscribirError(w, http.StatusBadRequest, "solicitud_invalida", "El cuerpo debe contener un único objeto JSON.", nil)
		return false
	}
	return true
}

func (c *Controlador) responderError(w http.ResponseWriter, err error) {
	var validacion modelo.ErrorValidacion
	switch {
	case errors.As(err, &validacion):
		vista.EscribirError(w, http.StatusUnprocessableEntity, "validacion", "Revisa los datos enviados.", validacion.Campos)
	case errors.Is(err, modelo.ErrNoEncontrado):
		vista.EscribirError(w, http.StatusNotFound, "no_encontrado", "No se encontró el recurso solicitado.", nil)
	case errors.Is(err, modelo.ErrEstadoInvalido), errors.Is(err, modelo.ErrSolicitudDuplicada),
		errors.Is(err, modelo.ErrColegioYaExiste):
		vista.EscribirError(w, http.StatusConflict, "conflicto", err.Error(), nil)
	case errors.Is(err, modelo.ErrImportacionInvalida), errors.Is(err, modelo.ErrCursoInvalido),
		errors.Is(err, modelo.ErrSolicitudInvalida):
		vista.EscribirError(w, http.StatusUnprocessableEntity, "datos_invalidos", err.Error(), nil)
	default:
		log.Printf("error del catálogo I07: %v", err)
		vista.EscribirError(w, http.StatusInternalServerError, "error_interno", "Ocurrió un error inesperado.", nil)
	}
}
