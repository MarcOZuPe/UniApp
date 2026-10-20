package acceso

import (
	"context"
	"log"
	"net/http"
	"time"

	modelo "uniapp/internal/models/acceso"
	vista "uniapp/internal/views/acceso"
)

type Middleware func(http.Handler) http.Handler

func Encadenar(manejador http.Handler, decoradores ...Middleware) http.Handler {
	for i := len(decoradores) - 1; i >= 0; i-- {
		manejador = decoradores[i](manejador)
	}
	return manejador
}

func Registro(registro *log.Logger) Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorRegistro{siguiente: siguiente, log: registro}
	}
}

func Recuperar(registro *log.Logger) Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorRecuperar{siguiente: siguiente, log: registro}
	}
}

func CabecerasSeguras() Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorCabeceras{siguiente: siguiente}
	}
}

func OrigenPermitido(permitidos ...string) Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorOrigen{siguiente: siguiente, permitidos: permitidos}
	}
}

func (c *Controlador) Autenticar() Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorAutenticar{siguiente: siguiente, controlador: c}
	}
}

func RequierePermiso(requeridos ...modelo.Permiso) Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorPermiso{siguiente: siguiente, requeridos: requeridos}
	}
}

type escritorEstado struct {
	http.ResponseWriter
	estado int
}

func (e *escritorEstado) WriteHeader(estado int) {
	e.estado = estado
	e.ResponseWriter.WriteHeader(estado)
}

type decoradorRegistro struct {
	siguiente http.Handler
	log       *log.Logger
}

func (d *decoradorRegistro) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inicio := time.Now()
	escritor := &escritorEstado{ResponseWriter: w, estado: http.StatusOK}
	d.siguiente.ServeHTTP(escritor, r)
	d.log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, escritor.estado, time.Since(inicio).Round(time.Millisecond))
}

type decoradorRecuperar struct {
	siguiente http.Handler
	log       *log.Logger
}

func (d *decoradorRecuperar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if causa := recover(); causa != nil {
			d.log.Printf("pánico: %v", causa)
			vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusInternalServerError, Codigo: "error_interno", Mensaje: "Ocurrió un error inesperado."})
		}
	}()
	d.siguiente.ServeHTTP(w, r)
}

type decoradorCabeceras struct{ siguiente http.Handler }

func (d *decoradorCabeceras) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Referrer-Policy", "no-referrer")
	d.siguiente.ServeHTTP(w, r)
}

type decoradorOrigen struct {
	siguiente  http.Handler
	permitidos []string
}

func (d *decoradorOrigen) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cambiaEstado(r.Method) {
		if origen := r.Header.Get("Origin"); origen != "" && !d.permite(origen) {
			vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusForbidden, Codigo: "origen_no_permitido", Mensaje: "Origen no permitido."})
			return
		}
	}
	d.siguiente.ServeHTTP(w, r)
}

func (d *decoradorOrigen) permite(origen string) bool {
	for _, permitido := range d.permitidos {
		if origen == permitido {
			return true
		}
	}
	return false
}

func cambiaEstado(metodo string) bool {
	return metodo != http.MethodGet && metodo != http.MethodHead && metodo != http.MethodOptions
}

type claveContexto struct{}

func UsuarioDe(ctx context.Context) *modelo.Usuario {
	usuario, _ := ctx.Value(claveContexto{}).(*modelo.Usuario)
	return usuario
}

func (c *Controlador) leerCookie(r *http.Request) string {
	if cookie, err := r.Cookie(nombreCookie); err == nil {
		return cookie.Value
	}
	return ""
}

type decoradorAutenticar struct {
	siguiente   http.Handler
	controlador *Controlador
}

func (d *decoradorAutenticar) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	usuario, err := d.controlador.servicio.UsuarioDeSesion(r.Context(), d.controlador.leerCookie(r))
	if err != nil {
		vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusUnauthorized, Codigo: "no_autenticado", Mensaje: "Debes iniciar sesión."})
		return
	}
	d.siguiente.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), claveContexto{}, usuario)))
}

type decoradorPermiso struct {
	siguiente  http.Handler
	requeridos []modelo.Permiso
}

func (d *decoradorPermiso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	usuario := UsuarioDe(r.Context())
	if usuario == nil {
		vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusUnauthorized, Codigo: "no_autenticado", Mensaje: "Debes iniciar sesión."})
		return
	}
	for _, requerido := range d.requeridos {
		if !usuario.Rol.TienePermiso(requerido) {
			vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusForbidden, Codigo: "sin_permiso", Mensaje: "Tu rol no tiene acceso a esta sección."})
			return
		}
	}
	d.siguiente.ServeHTTP(w, r)
}
