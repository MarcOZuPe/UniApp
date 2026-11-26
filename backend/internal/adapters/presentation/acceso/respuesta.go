package acceso

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type ErrorJSON struct {
	Codigo  string            `json:"codigo"`
	Mensaje string            `json:"mensaje"`
	Campos  map[string]string `json:"campos,omitempty"`
}

type RespuestaError struct {
	Estado  int
	Codigo  string
	Mensaje string
	Campos  map[string]string
}

func EscribirJSON(writer http.ResponseWriter, codigoEstado int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-store")
	writer.WriteHeader(codigoEstado)
	if payload != nil {
		_ = json.NewEncoder(writer).Encode(payload)
	}
}

func EscribirError(writer http.ResponseWriter, detalle RespuestaError) {
	EscribirJSON(writer, detalle.Estado, ErrorJSON{
		Codigo:  detalle.Codigo,
		Mensaje: detalle.Mensaje,
		Campos:  detalle.Campos,
	})
}

// EscribirJSONCacheable publica representaciones que no dependen de la sesión
// y que pueden ser reutilizadas por clientes e intermediarios. El ETag permite
// respuestas condicionales 304 sin volver a transferir el cuerpo.
func EscribirJSONCacheable(
	writer http.ResponseWriter,
	request *http.Request,
	codigoEstado int,
	payload any,
	maxAge int,
) {
	cuerpo, err := json.Marshal(payload)
	if err != nil {
		EscribirError(writer, RespuestaError{
			Estado: http.StatusInternalServerError, Codigo: "serializacion",
			Mensaje: "No se pudo preparar la respuesta.",
		})
		return
	}
	suma := sha256.Sum256(cuerpo)
	etag := `"` + hex.EncodeToString(suma[:16]) + `"`
	if maxAge < 0 {
		maxAge = 0
	}
	writer.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge)+", must-revalidate")
	writer.Header().Set("ETag", etag)
	writer.Header().Set("Vary", "Accept-Encoding")
	if etiquetaCoincide(request.Header.Get("If-None-Match"), etag) {
		writer.WriteHeader(http.StatusNotModified)
		return
	}
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(codigoEstado)
	_, _ = writer.Write(cuerpo)
}

func etiquetaCoincide(cabecera, etag string) bool {
	for _, valor := range strings.Split(cabecera, ",") {
		if strings.TrimSpace(valor) == etag || strings.TrimSpace(valor) == "*" {
			return true
		}
	}
	return false
}
