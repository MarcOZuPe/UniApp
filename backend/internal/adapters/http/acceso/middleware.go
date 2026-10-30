package acceso

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

type Middleware func(http.Handler) http.Handler

func Encadenar(manejador http.Handler, middleware ...Middleware) http.Handler {
	for indice := len(middleware) - 1; indice >= 0; indice-- {
		manejador = middleware[indice](manejador)
	}
	return manejador
}

func Registro(logger *log.Logger) Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorRegistro{siguiente: siguiente, logger: logger}
	}
}

func Recuperar(logger *log.Logger) Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorRecuperar{siguiente: siguiente, logger: logger}
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

func (controlador *controladorBase) Autenticar() Middleware {
	return func(siguiente http.Handler) http.Handler {
		return &decoradorAutenticar{siguiente: siguiente, controlador: controlador}
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

func (escritor *escritorEstado) WriteHeader(estado int) {
	escritor.estado = estado
	escritor.ResponseWriter.WriteHeader(estado)
}

type decoradorRegistro struct {
	siguiente http.Handler
	logger    *log.Logger
}

func (decorador *decoradorRegistro) ServeHTTP(respuesta http.ResponseWriter, peticion *http.Request) {
	inicio := time.Now()
	escritor := &escritorEstado{ResponseWriter: respuesta, estado: http.StatusOK}
	decorador.siguiente.ServeHTTP(escritor, peticion)
	decorador.logger.Printf("%s %s -> %d (%s)",
		peticion.Method, peticion.URL.Path, escritor.estado, time.Since(inicio).Round(time.Millisecond))
}

type decoradorRecuperar struct {
	siguiente http.Handler
	logger    *log.Logger
}

func (decorador *decoradorRecuperar) ServeHTTP(respuesta http.ResponseWriter, peticion *http.Request) {
	defer func() {
		if causa := recover(); causa != nil {
			decorador.logger.Printf("pánico: %v", causa)
			vista.EscribirError(respuesta, vista.RespuestaError{
				Estado:  http.StatusInternalServerError,
				Codigo:  "error_interno",
				Mensaje: "Ocurrió un error inesperado.",
			})
		}
	}()
	decorador.siguiente.ServeHTTP(respuesta, peticion)
}

type decoradorCabeceras struct{ siguiente http.Handler }

func (decorador *decoradorCabeceras) ServeHTTP(respuesta http.ResponseWriter, peticion *http.Request) {
	respuesta.Header().Set("X-Content-Type-Options", "nosniff")
	respuesta.Header().Set("X-Frame-Options", "DENY")
	respuesta.Header().Set("Referrer-Policy", "no-referrer")
	decorador.siguiente.ServeHTTP(respuesta, peticion)
}

type decoradorOrigen struct {
	siguiente  http.Handler
	permitidos []string
}

func (decorador *decoradorOrigen) ServeHTTP(respuesta http.ResponseWriter, peticion *http.Request) {
	if cambiaEstado(peticion.Method) {
		origen := peticion.Header.Get("Origin")
		if origen != "" && !decorador.permite(origen) && !mismoOrigen(origen, peticion.Host) {
			vista.EscribirError(respuesta, vista.RespuestaError{
				Estado:  http.StatusForbidden,
				Codigo:  "origen_no_permitido",
				Mensaje: "Origen no permitido.",
			})
			return
		}
	}
	decorador.siguiente.ServeHTTP(respuesta, peticion)
}

func (decorador *decoradorOrigen) permite(origen string) bool {
	for _, permitido := range decorador.permitidos {
		if origen == permitido {
			return true
		}
	}
	return false
}

func cambiaEstado(metodo string) bool {
	return metodo != http.MethodGet && metodo != http.MethodHead && metodo != http.MethodOptions
}

func mismoOrigen(origen, host string) bool {
	urlOrigen, err := url.Parse(origen)
	if err != nil || urlOrigen.Host == "" {
		return false
	}
	return strings.EqualFold(urlOrigen.Host, host)
}

type claveContexto struct{}

func UsuarioDe(contexto context.Context) *modelo.Usuario {
	usuario, _ := contexto.Value(claveContexto{}).(*modelo.Usuario)
	return usuario
}

func (controlador *controladorBase) leerCookie(peticion *http.Request) string {
	if cookie, err := peticion.Cookie(nombreCookie); err == nil {
		return cookie.Value
	}
	return ""
}

type decoradorAutenticar struct {
	siguiente   http.Handler
	controlador *controladorBase
}

func (decorador *decoradorAutenticar) ServeHTTP(respuesta http.ResponseWriter, peticion *http.Request) {
	usuario, err := decorador.controlador.servicio.Sesion.UsuarioDeSesion(
		peticion.Context(), decorador.controlador.leerCookie(peticion),
	)
	if err != nil {
		vista.EscribirError(respuesta, vista.RespuestaError{
			Estado:  http.StatusUnauthorized,
			Codigo:  "no_autenticado",
			Mensaje: "Debes iniciar sesión.",
		})
		return
	}
	contexto := context.WithValue(peticion.Context(), claveContexto{}, usuario)
	decorador.siguiente.ServeHTTP(respuesta, peticion.WithContext(contexto))
}

type decoradorPermiso struct {
	siguiente  http.Handler
	requeridos []modelo.Permiso
}

func (decorador *decoradorPermiso) ServeHTTP(respuesta http.ResponseWriter, peticion *http.Request) {
	usuario := UsuarioDe(peticion.Context())
	if usuario == nil {
		vista.EscribirError(respuesta, vista.RespuestaError{
			Estado:  http.StatusUnauthorized,
			Codigo:  "no_autenticado",
			Mensaje: "Debes iniciar sesión.",
		})
		return
	}
	for _, requerido := range decorador.requeridos {
		if !usuario.Rol.TienePermiso(requerido) {
			vista.EscribirError(respuesta, vista.RespuestaError{
				Estado:  http.StatusForbidden,
				Codigo:  "sin_permiso",
				Mensaje: "Tu rol no tiene acceso a esta sección.",
			})
			return
		}
	}
	decorador.siguiente.ServeHTTP(respuesta, peticion)
}
