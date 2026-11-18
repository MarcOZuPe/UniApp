package incidente

import (
	"fmt"
	"time"

	modelo "uniapp/internal/application/incidente"
)

const formatoLima = "2006-01-02 15:04"

type TutoriaJSON struct {
	ReservaID     string    `json:"reserva_id"`
	EscolarID     string    `json:"escolar_id"`
	EscolarNombre string    `json:"escolar_nombre"`
	TutorID       string    `json:"tutor_id"`
	TutorNombre   string    `json:"tutor_nombre"`
	Curso         string    `json:"curso"`
	Inicio        time.Time `json:"inicio"`
	Fin           time.Time `json:"fin"`
	InicioLima    string    `json:"inicio_lima"`
	FinLima       string    `json:"fin_lima"`
	Asistio       bool      `json:"asistio"`
}

type EvidenciaJSON struct {
	ID          string    `json:"id"`
	Nombre      string    `json:"nombre"`
	Tipo        string    `json:"tipo"`
	TamanoBytes int64     `json:"tamano_bytes"`
	CreadoEn    time.Time `json:"creado_en"`
	DescargarEn string    `json:"descargar_en"`
}

type IncidenteJSON struct {
	ID                 string          `json:"id"`
	EscolarID          string          `json:"escolar_id"`
	EscolarNombre      string          `json:"escolar_nombre"`
	ReportadoPor       string          `json:"reportado_por"`
	TutorID            string          `json:"tutor_id"`
	TutorNombre        string          `json:"tutor_nombre"`
	ReservaID          string          `json:"reserva_id"`
	Curso              string          `json:"curso"`
	Inicio             time.Time       `json:"inicio"`
	InicioLima         string          `json:"inicio_lima"`
	Tipo               string          `json:"tipo"`
	TipoEtiqueta       string          `json:"tipo_etiqueta"`
	Descripcion        string          `json:"descripcion"`
	Estado             string          `json:"estado"`
	Resultado          string          `json:"resultado,omitempty"`
	ComentarioRevision string          `json:"comentario_revision,omitempty"`
	RevisadoPor        string          `json:"revisado_por,omitempty"`
	RevisadoEn         *time.Time      `json:"revisado_en,omitempty"`
	CreadoEn           time.Time       `json:"creado_en"`
	Evidencias         []EvidenciaJSON `json:"evidencias"`
}

func PresentarTutorias(tutorias []modelo.Tutoria) []TutoriaJSON {
	presentadas := make([]TutoriaJSON, 0, len(tutorias))
	for _, tutoria := range tutorias {
		presentadas = append(presentadas, TutoriaJSON{
			ReservaID:     tutoria.ReservaID,
			EscolarID:     tutoria.EscolarID,
			EscolarNombre: tutoria.EscolarNombre,
			TutorID:       tutoria.TutorID,
			TutorNombre:   tutoria.TutorNombre,
			Curso:         tutoria.Curso,
			Inicio:        tutoria.Inicio,
			Fin:           tutoria.Fin,
			InicioLima:    tutoria.Inicio.In(modeloZonaLima()).Format(formatoLima),
			FinLima:       tutoria.Fin.In(modeloZonaLima()).Format(formatoLima),
			Asistio:       tutoria.Asistio,
		})
	}
	return presentadas
}

func PresentarIncidente(incidente *modelo.Incidente) IncidenteJSON {
	if incidente == nil {
		return IncidenteJSON{Evidencias: []EvidenciaJSON{}}
	}
	evidencias := make([]EvidenciaJSON, 0, len(incidente.Evidencias))
	for _, evidencia := range incidente.Evidencias {
		evidencias = append(evidencias, EvidenciaJSON{
			ID:          evidencia.ID,
			Nombre:      evidencia.Nombre,
			Tipo:        evidencia.Tipo,
			TamanoBytes: evidencia.TamanoBytes,
			CreadoEn:    evidencia.CreadoEn,
			DescargarEn: fmt.Sprintf("/api/incidentes/%s/evidencias/%s", incidente.ID, evidencia.ID),
		})
	}
	return IncidenteJSON{
		ID:                 incidente.ID,
		EscolarID:          incidente.EscolarID,
		EscolarNombre:      incidente.EscolarNombre,
		ReportadoPor:       incidente.ReportadoPor,
		TutorID:            incidente.TutorID,
		TutorNombre:        incidente.TutorNombre,
		ReservaID:          incidente.ReservaID,
		Curso:              incidente.Curso,
		Inicio:             incidente.Inicio,
		InicioLima:         incidente.Inicio.In(modeloZonaLima()).Format(formatoLima),
		Tipo:               string(incidente.Tipo),
		TipoEtiqueta:       modelo.TiposValidos[incidente.Tipo],
		Descripcion:        incidente.Descripcion,
		Estado:             string(incidente.Estado),
		Resultado:          string(incidente.Resultado),
		ComentarioRevision: incidente.ComentarioRevision,
		RevisadoPor:        incidente.RevisadoPor,
		RevisadoEn:         incidente.RevisadoEn,
		CreadoEn:           incidente.CreadoEn,
		Evidencias:         evidencias,
	}
}

func PresentarIncidentes(incidentes []modelo.Incidente) []IncidenteJSON {
	presentados := make([]IncidenteJSON, 0, len(incidentes))
	for indice := range incidentes {
		presentados = append(presentados, PresentarIncidente(&incidentes[indice]))
	}
	return presentados
}

func modeloZonaLima() *time.Location {
	return time.FixedZone("America/Lima", -5*60*60)
}
