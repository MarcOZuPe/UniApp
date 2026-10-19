package acceso

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	modelo "uniapp/internal/models/acceso"
	vista "uniapp/internal/views/acceso"
)

const (
	nombreCookie = "uniapp_sesion"
	rutaCookie   = "/"
)

type Controlador struct {
	servicio     *modelo.Servicio
	cookieSegura bool
}

func Nuevo(servicio *modelo.Servicio, cookieSegura bool) *Controlador {
	return &Controlador{servicio: servicio, cookieSegura: cookieSegura}
}

func (c *Controlador) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/registro", c.registro)
	mux.HandleFunc("POST /api/auth/ingreso", c.ingreso)
	mux.HandleFunc("POST /api/auth/recuperar", c.recuperar)
	mux.HandleFunc("POST /api/auth/restablecer", c.restablecer)
	mux.Handle("POST /api/auth/salida", Encadenar(http.HandlerFunc(c.salida), c.Autenticar()))
	mux.Handle("GET /api/auth/yo", Encadenar(http.HandlerFunc(c.yo), c.Autenticar()))
	mux.HandleFunc("GET /api/salud", c.salud)
	return mux
}

func (c *Controlador) cookieBase() *http.Cookie {
	return &http.Cookie{
		Name:     nombreCookie,
		Path:     rutaCookie,
		HttpOnly: true,
		Secure:   c.cookieSegura,
		SameSite: http.SameSiteStrictMode,
	}
}

func (c *Controlador) cookieDeSesion(token string) *http.Cookie {
	cookie := c.cookieBase()
	cookie.Value = token
	cookie.Expires = time.Now().Add(modelo.DuracionSesion)
	return cookie
}

func (c *Controlador) cookieExpirada() *http.Cookie {
	cookie := c.cookieBase()
	cookie.MaxAge = -1
	return cookie
}

func (c *Controlador) salud(w http.ResponseWriter, _ *http.Request) {
	vista.EscribirJSON(w, http.StatusOK, map[string]string{"estado": "ok"})
}

func (c *Controlador) decodificarCuerpo(w http.ResponseWriter, r *http.Request, destino any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decodificador := json.NewDecoder(r.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		vista.EscribirError(w, vista.RespuestaError{
			Estado:  http.StatusBadRequest,
			Codigo:  "solicitud_invalida",
			Mensaje: "El cuerpo de la solicitud no es válido.",
		})
		return false
	}
	return true
}

func (c *Controlador) responderError(w http.ResponseWriter, err error) {
	var errValidacion modelo.ErrValidacion
	switch {
	case errors.As(err, &errValidacion):
		vista.EscribirError(w, vista.RespuestaError{
			Estado: http.StatusUnprocessableEntity, Codigo: "validacion",
			Mensaje: "Revisa los datos ingresados.", Campos: errValidacion.Campos,
		})
	case errors.Is(err, modelo.ErrEmailEnUso):
		vista.EscribirError(w, vista.RespuestaError{
			Estado: http.StatusConflict, Codigo: "correo_en_uso",
			Mensaje: err.Error(), Campos: map[string]string{"email": err.Error()},
		})
	case errors.Is(err, modelo.ErrRolNoPermitido):
		vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusForbidden, Codigo: "rol_no_permitido", Mensaje: err.Error()})
	case errors.Is(err, modelo.ErrCredenciales):
		vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusUnauthorized, Codigo: "credenciales", Mensaje: err.Error()})
	case errors.Is(err, modelo.ErrDemasiadosIntentos):
		vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusTooManyRequests, Codigo: "demasiados_intentos", Mensaje: err.Error()})
	case errors.Is(err, modelo.ErrTokenInvalido):
		vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusBadRequest, Codigo: "token_invalido", Mensaje: err.Error()})
	default:
		vista.EscribirError(w, vista.RespuestaError{Estado: http.StatusInternalServerError, Codigo: "error_interno", Mensaje: "Ocurrió un error inesperado."})
	}
}
