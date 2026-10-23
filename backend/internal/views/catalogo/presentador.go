package catalogo

import (
	"encoding/json"
	"net/http"
	"time"

	modelo "uniapp/internal/models/catalogo"
)

type Curso struct {
	ID     int64  `json:"id"`
	Nombre string `json:"nombre"`
	Grado  string `json:"grado"`
	Estado string `json:"estado"`
	Motivo string `json:"motivo_rechazo,omitempty"`
}

type Colegio struct {
	Nombre   string `json:"nombre"`
	Distrito string `json:"distrito"`
}

type Importacion struct {
	ID          int64     `json:"id"`
	Fuente      string    `json:"fuente"`
	ImportadoEn time.Time `json:"importado_en"`
	Cursos      []Curso   `json:"cursos"`
	Colegios    []Colegio `json:"colegios"`
}

type SolicitudColegio struct {
	ID           int64     `json:"id"`
	Nombre       string    `json:"nombre"`
	Distrito     string    `json:"distrito"`
	SolicitadoEn time.Time `json:"solicitado_en"`
}

type ErrorJSON struct {
	Codigo  string            `json:"codigo"`
	Mensaje string            `json:"mensaje"`
	Campos  map[string]string `json:"campos,omitempty"`
}

func PresentarImportacion(importacion *modelo.Importacion) Importacion {
	resultado := Importacion{
		ID: importacion.ID, Fuente: importacion.Fuente, ImportadoEn: importacion.ImportadoEn,
		Cursos:   make([]Curso, 0, len(importacion.Cursos)),
		Colegios: make([]Colegio, 0, len(importacion.Colegios)),
	}
	for _, curso := range importacion.Cursos {
		resultado.Cursos = append(resultado.Cursos, Curso{
			ID: curso.ID, Nombre: curso.Nombre, Grado: curso.Grado,
			Estado: string(curso.Estado), Motivo: curso.Motivo,
		})
	}
	for _, colegio := range importacion.Colegios {
		resultado.Colegios = append(resultado.Colegios, Colegio{
			Nombre: colegio.Nombre, Distrito: colegio.Distrito,
		})
	}
	return resultado
}

func PresentarSolicitudes(solicitudes []modelo.SolicitudColegio) []SolicitudColegio {
	resultado := make([]SolicitudColegio, 0, len(solicitudes))
	for _, solicitud := range solicitudes {
		resultado = append(resultado, SolicitudColegio{
			ID: solicitud.ID, Nombre: solicitud.Nombre, Distrito: solicitud.Distrito,
			SolicitadoEn: solicitud.SolicitadoEn,
		})
	}
	return resultado
}

func EscribirJSON(w http.ResponseWriter, estado int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	if valor != nil {
		_ = json.NewEncoder(w).Encode(valor)
	}
}

func EscribirError(w http.ResponseWriter, estado int, codigo, mensaje string, campos map[string]string) {
	EscribirJSON(w, estado, ErrorJSON{Codigo: codigo, Mensaje: mensaje, Campos: campos})
}
