package tutor

import (
	"errors"
	"path"
	"strings"
	"time"
)

const (
	EstadoPendiente Estado = "pendiente"
	EstadoAprobado  Estado = "aprobado"
	EstadoRechazado Estado = "rechazado"
	MaximoDocumento        = 10 << 20
)

type Estado string

type Documento struct {
	ID       string
	Nombre   string
	Tamano   int64
	Tipo     string
	CreadoEn time.Time
}

type Expediente struct {
	TutorID      string
	Presentacion string
	Cursos       string
	Experiencia  string
	Estado       Estado
	PresentadoEn time.Time
	VenceEn      *time.Time
	RevisadoPor  string
	RevisadoEn   *time.Time
	Motivo       string
	Documentos   []Documento
}

type PerfilListado struct {
	TutorID      string
	Presentacion string
	Cursos       string
	Experiencia  string
}

type DatosExpediente struct {
	Presentacion string
	Cursos       string
	Experiencia  string
	Documentos   []Archivo
}

type Archivo struct {
	Nombre    string
	Contenido []byte
}

var (
	ErrNoEncontrado       = errors.New("expediente no encontrado")
	ErrSinPermiso         = errors.New("sin permiso para esta operación")
	ErrExpedienteInvalido = errors.New("expediente inválido")
	ErrEstadoInvalido     = errors.New("estado de revisión inválido")
	ErrDocumentoInvalido  = errors.New("documento inválido")
	ErrNoPublicable       = errors.New("perfil de tutor no disponible")
)

type ErrorValidacion struct {
	Campos map[string]string
}

func (e ErrorValidacion) Error() string { return ErrExpedienteInvalido.Error() }

func NuevoExpediente(tutorID string, datos DatosExpediente, ahora time.Time) (*Expediente, error) {
	datos.Presentacion = strings.TrimSpace(datos.Presentacion)
	datos.Cursos = strings.TrimSpace(datos.Cursos)
	datos.Experiencia = strings.TrimSpace(datos.Experiencia)

	campos := map[string]string{}
	if tutorID == "" {
		campos["tutor_id"] = "No se pudo identificar la cuenta."
	}
	if datos.Presentacion == "" || len([]rune(datos.Presentacion)) > 500 {
		campos["presentacion"] = "Es obligatoria y debe tener hasta 500 caracteres."
	}
	if datos.Cursos == "" || len([]rune(datos.Cursos)) > 500 {
		campos["cursos"] = "Indica los cursos que puedes enseñar (hasta 500 caracteres)."
	}
	if datos.Experiencia == "" || len([]rune(datos.Experiencia)) > 2000 {
		campos["experiencia"] = "Es obligatoria y debe tener hasta 2000 caracteres."
	}
	if len(datos.Documentos) == 0 || len(datos.Documentos) > 5 {
		campos["documentos"] = "Adjunta entre 1 y 5 documentos PDF."
	}
	for _, archivo := range datos.Documentos {
		if !esPDF(archivo) {
			campos["documentos"] = "Cada documento debe ser un PDF válido de hasta 10 MB."
			break
		}
	}
	if len(campos) > 0 {
		return nil, ErrorValidacion{Campos: campos}
	}
	if ahora.IsZero() {
		return nil, ErrExpedienteInvalido
	}

	documentos := make([]Documento, 0, len(datos.Documentos))
	for _, archivo := range datos.Documentos {
		documentos = append(documentos, Documento{
			Nombre: nombreSeguro(archivo.Nombre),
			Tamano: int64(len(archivo.Contenido)),
			Tipo:   "application/pdf",
		})
	}

	return &Expediente{
		TutorID:      tutorID,
		Presentacion: datos.Presentacion,
		Cursos:       datos.Cursos,
		Experiencia:  datos.Experiencia,
		Estado:       EstadoPendiente,
		PresentadoEn: ahora.UTC(),
		Documentos:   documentos,
	}, nil
}

func (e *Expediente) Revisar(estado Estado, revisorID, motivo string, venceEn, ahora time.Time) error {
	if e == nil || e.TutorID == "" || revisorID == "" || ahora.IsZero() {
		return ErrExpedienteInvalido
	}
	if e.Estado != EstadoPendiente {
		return ErrEstadoInvalido
	}

	switch estado {
	case EstadoAprobado:
		if venceEn.IsZero() || !venceEn.After(ahora) {
			return ErrExpedienteInvalido
		}
		venceEn = venceEn.UTC()
		e.VenceEn = &venceEn
		e.Motivo = ""
	case EstadoRechazado:
		if strings.TrimSpace(motivo) == "" {
			return ErrExpedienteInvalido
		}
		e.VenceEn = nil
		e.Motivo = strings.TrimSpace(motivo)
	default:
		return ErrEstadoInvalido
	}

	revisadoEn := ahora.UTC()
	e.Estado = estado
	e.RevisadoPor = revisorID
	e.RevisadoEn = &revisadoEn
	return nil
}

func (e Expediente) Publicable(ahora time.Time) bool {
	return e.Estado == EstadoAprobado && e.VenceEn != nil && ahora.Before(*e.VenceEn)
}

func esPDF(archivo Archivo) bool {
	return len(archivo.Contenido) > 5 &&
		len(archivo.Contenido) <= MaximoDocumento &&
		strings.EqualFold(path.Ext(nombreSeguro(archivo.Nombre)), ".pdf") &&
		string(archivo.Contenido[:5]) == "%PDF-"
}

func nombreSeguro(nombre string) string {
	nombre = strings.ReplaceAll(nombre, `\`, "/")
	partes := strings.Split(nombre, "/")
	nombre = strings.TrimSpace(partes[len(partes)-1])
	if nombre == "" || nombre == "." {
		return "documento.pdf"
	}
	return nombre
}

func Copiar(expediente *Expediente) *Expediente {
	if expediente == nil {
		return nil
	}
	copia := *expediente
	copia.Documentos = append([]Documento(nil), expediente.Documentos...)
	if expediente.VenceEn != nil {
		vencimiento := *expediente.VenceEn
		copia.VenceEn = &vencimiento
	}
	if expediente.RevisadoEn != nil {
		revision := *expediente.RevisadoEn
		copia.RevisadoEn = &revision
	}
	return &copia
}
