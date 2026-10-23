package tutor

import (
	"time"

	modelo "uniapp/internal/models/tutor"
)

type Documento struct {
	ID       string    `json:"id"`
	Nombre   string    `json:"nombre"`
	Tamano   int64     `json:"tamano_bytes"`
	Tipo     string    `json:"tipo"`
	CreadoEn time.Time `json:"creado_en"`
}

type Expediente struct {
	TutorID      string      `json:"tutor_id"`
	Presentacion string      `json:"presentacion"`
	Cursos       string      `json:"cursos"`
	Experiencia  string      `json:"experiencia"`
	Estado       string      `json:"estado"`
	PresentadoEn time.Time   `json:"presentado_en"`
	VenceEn      *time.Time  `json:"vence_en,omitempty"`
	Motivo       string      `json:"motivo_rechazo,omitempty"`
	Documentos   []Documento `json:"documentos"`
}

type PerfilPublico struct {
	TutorID      string `json:"tutor_id"`
	Presentacion string `json:"presentacion"`
	Cursos       string `json:"cursos"`
	Experiencia  string `json:"experiencia"`
}

func PresentarPerfilPublico(expediente *modelo.Expediente) PerfilPublico {
	return PerfilPublico{
		TutorID: expediente.TutorID, Presentacion: expediente.Presentacion,
		Cursos: expediente.Cursos, Experiencia: expediente.Experiencia,
	}
}

func PresentarPerfilesPublicos(perfiles []modelo.PerfilListado) []PerfilPublico {
	resultado := make([]PerfilPublico, 0, len(perfiles))
	for _, perfil := range perfiles {
		resultado = append(resultado, PerfilPublico{
			TutorID: perfil.TutorID, Presentacion: perfil.Presentacion,
			Cursos: perfil.Cursos, Experiencia: perfil.Experiencia,
		})
	}
	return resultado
}

func PresentarExpediente(expediente *modelo.Expediente) Expediente {
	documentos := make([]Documento, 0, len(expediente.Documentos))
	for _, documento := range expediente.Documentos {
		documentos = append(documentos, Documento{
			ID: documento.ID, Nombre: documento.Nombre, Tamano: documento.Tamano,
			Tipo: documento.Tipo, CreadoEn: documento.CreadoEn,
		})
	}
	return Expediente{
		TutorID: expediente.TutorID, Presentacion: expediente.Presentacion,
		Cursos: expediente.Cursos, Experiencia: expediente.Experiencia,
		Estado: string(expediente.Estado), PresentadoEn: expediente.PresentadoEn,
		VenceEn: expediente.VenceEn, Motivo: expediente.Motivo, Documentos: documentos,
	}
}
