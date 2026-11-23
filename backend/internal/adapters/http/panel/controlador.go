package panel

import (
	"errors"
	"net/http"
	"strconv"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	vistaAcceso "uniapp/internal/adapters/presentation/acceso"
	vista "uniapp/internal/adapters/presentation/panel"
	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/application/panel"
	dominio "uniapp/internal/domain/panel"
)

type Controlador struct {
	nucleo     *modelo.Servicio
	autenticar ctrlAcceso.Middleware
	permiso    ctrlAcceso.Middleware
	errores    *ctrlAcceso.MapeadorDeErrores
}

func Nuevo(nucleo *modelo.Servicio, autenticar ctrlAcceso.Middleware) *Controlador {
	errores := ctrlAcceso.NuevoMapeadorDeErrores()
	errores.Registrar(
		func(err error) bool { return errors.Is(err, dominio.ErrNoAutorizado) },
		func(error) vistaAcceso.RespuestaError {
			return vistaAcceso.RespuestaError{
				Estado: http.StatusForbidden, Codigo: "sin_permiso",
				Mensaje: "Tu rol no tiene acceso al panel de coordinación.",
			}
		},
	)
	return &Controlador{
		nucleo:     nucleo,
		autenticar: autenticar,
		permiso:    ctrlAcceso.RequierePermiso(modeloAcceso.PermisoCoordinacionAcceder),
		errores:    errores,
	}
}

func (c *Controlador) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /api/coordinacion/metricas", c.proteger(http.HandlerFunc(c.metricas)))
	mux.Handle("GET /api/coordinacion/auditoria", c.proteger(http.HandlerFunc(c.auditoria)))
	return mux
}

func (c *Controlador) proteger(manejador http.Handler) http.Handler {
	return ctrlAcceso.Encadenar(manejador, c.autenticar, c.permiso)
}

func (c *Controlador) metricas(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, dominio.ErrNoAutorizado)
		return
	}
	metricas, err := c.nucleo.Metricas(r.Context(), usuario.Rol)
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, vista.PresentarMetricas(metricas))
}

func (c *Controlador) auditoria(w http.ResponseWriter, r *http.Request) {
	usuario := ctrlAcceso.UsuarioDe(r.Context())
	if usuario == nil {
		c.errores.Escribir(w, dominio.ErrNoAutorizado)
		return
	}
	limite := 0
	if valor := r.URL.Query().Get("limite"); valor != "" {
		convertido, err := strconv.Atoi(valor)
		if err != nil || convertido <= 0 {
			vistaAcceso.EscribirError(w, vistaAcceso.RespuestaError{
				Estado: http.StatusBadRequest, Codigo: "limite_invalido",
				Mensaje: "El límite debe ser un número mayor que cero.",
			})
			return
		}
		limite = convertido
	}
	eventos, err := c.nucleo.Auditoria(r.Context(), usuario.Rol, dominio.FiltroAuditoria{
		Limite: limite,
	})
	if err != nil {
		c.errores.Escribir(w, err)
		return
	}
	vistaAcceso.EscribirJSON(w, http.StatusOK, map[string]any{
		"eventos": vista.PresentarAuditoria(eventos),
	})
}
