package acceso

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	vista "uniapp/internal/adapters/presentation/acceso"
	modelo "uniapp/internal/application/acceso"
)

const (
	nombreCookie = "uniapp_sesion"
	rutaCookie   = "/"
)

type controladorBase struct {
	servicio     *modelo.Servicio
	cookieSegura bool
	errores      *MapeadorDeErrores
}

func (base *controladorBase) decodificarCuerpo(writer http.ResponseWriter, request *http.Request, destino any) bool {
	request.Body = http.MaxBytesReader(writer, request.Body, 1<<20)
	decodificador := json.NewDecoder(request.Body)
	decodificador.DisallowUnknownFields()

	if err := decodificador.Decode(destino); err != nil {
		base.renderizarErrorDeLectura(writer)
		return false
	}
	return true
}

func (base *controladorBase) responderError(writer http.ResponseWriter, err error) {
	base.errores.Escribir(writer, err)
}

func (base *controladorBase) cookieDeSesion(token string) *http.Cookie {
	cookie := base.compilarCookieBase()
	cookie.Value = token
	cookie.Expires = time.Now().Add(modelo.DuracionSesion)
	return cookie
}

func (base *controladorBase) cookieExpirada() *http.Cookie {
	cookie := base.compilarCookieBase()
	cookie.MaxAge = -1
	return cookie
}

func (base *controladorBase) compilarCookieBase() *http.Cookie {
	return &http.Cookie{
		Name:     nombreCookie,
		Path:     rutaCookie,
		HttpOnly: true,
		Secure:   base.cookieSegura,
		SameSite: http.SameSiteStrictMode,
	}
}

func (base *controladorBase) renderizarErrorDeLectura(writer http.ResponseWriter) {
	vista.EscribirError(writer, vista.RespuestaError{
		Estado:  http.StatusBadRequest,
		Codigo:  "solicitud_invalida",
		Mensaje: "El cuerpo de la solicitud no es válido.",
	})
}

func ipCliente(request *http.Request) string {
	if reenviada := request.Header.Get("X-Forwarded-For"); reenviada != "" {
		if primera := strings.TrimSpace(strings.Split(reenviada, ",")[0]); primera != "" {
			return recortarIP(primera)
		}
	}
	if real := strings.TrimSpace(request.Header.Get("X-Real-IP")); real != "" {
		return recortarIP(real)
	}
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return recortarIP(host)
	}
	return recortarIP(request.RemoteAddr)
}

func recortarIP(ip string) string {
	if len(ip) > 128 {
		return ip[:128]
	}
	return ip
}
