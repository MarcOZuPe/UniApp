package acceso

import (
	"encoding/json"
	"net/http"
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
