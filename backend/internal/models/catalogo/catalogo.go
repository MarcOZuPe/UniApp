package catalogo

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"
)

type EstadoCurso string
type EstadoSolicitud string

const (
	EstadoPendiente EstadoCurso = "pendiente"
	EstadoAprobado  EstadoCurso = "aprobado"
	EstadoRechazado EstadoCurso = "rechazado"

	EstadoSolicitudPendiente   EstadoSolicitud = "pendiente"
	EstadoSolicitudIncorporada EstadoSolicitud = "incorporada"
)

var (
	ErrImportacionInvalida = errors.New("importación de catálogo inválida")
	ErrNoEncontrado        = errors.New("recurso de catálogo no encontrado")
	ErrEstadoInvalido      = errors.New("estado de aprobación inválido")
	ErrCursoInvalido       = errors.New("curso inválido")
	ErrSolicitudInvalida   = errors.New("solicitud de colegio inválida")
	ErrColegioYaExiste     = errors.New("el colegio ya está en el padrón")
	ErrSolicitudDuplicada  = errors.New("ya existe una solicitud pendiente para ese colegio")
)

type ErrorValidacion struct {
	Campos map[string]string
}

func (e ErrorValidacion) Error() string { return ErrImportacionInvalida.Error() }

type CursoEntrada struct {
	Nombre string
	Grado  string
}

type ColegioEntrada struct {
	Nombre   string
	Distrito string
}

type DatosImportacion struct {
	Fuente   string
	Cursos   []CursoEntrada
	Colegios []ColegioEntrada
}

type Curso struct {
	ID          int64
	Nombre      string
	Grado       string
	Estado      EstadoCurso
	Motivo      string
	RevisadoPor string
	RevisadoEn  *time.Time
}

type Colegio struct {
	Nombre   string
	Distrito string
}

type Importacion struct {
	ID          int64
	Fuente      string
	ImportadoEn time.Time
	Cursos      []Curso
	Colegios    []Colegio
}

type SolicitudColegio struct {
	ID           int64
	Nombre       string
	Distrito     string
	Estado       EstadoSolicitud
	SolicitadoEn time.Time
}

type SolicitudColegioEntrada struct {
	Nombre   string
	Distrito string
}

type Repositorio interface {
	Importar(context.Context, *Importacion) error
	ImportacionActual(context.Context) (*Importacion, error)
	Curso(context.Context, int64) (Curso, error)
	RevisarCurso(context.Context, Curso, string) error
	SolicitarColegio(context.Context, SolicitudColegio, string) error
	SolicitudesPendientes(context.Context) ([]SolicitudColegio, error)
}

func NuevaImportacion(datos DatosImportacion, ahora time.Time) (*Importacion, error) {
	datos.Fuente = strings.TrimSpace(datos.Fuente)
	campos := make(map[string]string)
	if !textoValido(datos.Fuente, 120) {
		campos["fuente"] = "Indica un origen de hasta 120 caracteres."
	}
	if len(datos.Cursos) == 0 || len(datos.Cursos) > 500 {
		campos["cursos"] = "La importación debe incluir entre 1 y 500 cursos."
	}
	if len(datos.Colegios) == 0 || len(datos.Colegios) > 2000 {
		campos["colegios"] = "La importación debe incluir entre 1 y 2000 colegios."
	}
	cursos := make([]Curso, 0, len(datos.Cursos))
	cursosVistos := make(map[string]bool, len(datos.Cursos))
	for _, curso := range datos.Cursos {
		curso.Nombre = strings.TrimSpace(curso.Nombre)
		curso.Grado = strings.TrimSpace(curso.Grado)
		if !textoValido(curso.Nombre, 120) || !textoValido(curso.Grado, 80) {
			campos["cursos"] = "Cada curso debe incluir nombre y grado válidos."
			break
		}
		curso.Nombre = compactarEspacios(curso.Nombre)
		curso.Grado = compactarEspacios(curso.Grado)
		clave := claveNormalizada(curso.Nombre) + "\x00" + claveNormalizada(curso.Grado)
		if cursosVistos[clave] {
			campos["cursos"] = "No repitas un curso para el mismo grado."
			break
		}
		cursosVistos[clave] = true
		cursos = append(cursos, Curso{Nombre: curso.Nombre, Grado: curso.Grado, Estado: EstadoPendiente})
	}

	colegios := make([]Colegio, 0, len(datos.Colegios))
	colegiosVistos := make(map[string]bool, len(datos.Colegios))
	for _, colegio := range datos.Colegios {
		colegio.Nombre = strings.TrimSpace(colegio.Nombre)
		colegio.Distrito = strings.TrimSpace(colegio.Distrito)
		if !textoValido(colegio.Nombre, 160) || !textoValido(colegio.Distrito, 100) {
			campos["colegios"] = "Cada colegio debe incluir nombre y distrito válidos."
			break
		}
		colegio.Nombre = compactarEspacios(colegio.Nombre)
		colegio.Distrito = compactarEspacios(colegio.Distrito)
		clave := claveNormalizada(colegio.Nombre) + "\x00" + claveNormalizada(colegio.Distrito)
		if colegiosVistos[clave] {
			campos["colegios"] = "No repitas colegios para el mismo distrito."
			break
		}
		colegiosVistos[clave] = true
		colegios = append(colegios, Colegio{Nombre: colegio.Nombre, Distrito: colegio.Distrito})
	}
	if ahora.IsZero() {
		campos["importado_en"] = "No se pudo registrar la fecha de importación."
	}
	if len(campos) > 0 {
		return nil, ErrorValidacion{Campos: campos}
	}
	return &Importacion{
		Fuente: datos.Fuente, ImportadoEn: ahora.UTC(), Cursos: cursos, Colegios: colegios,
	}, nil
}

func NuevaSolicitudColegio(datos SolicitudColegioEntrada, ahora time.Time) (SolicitudColegio, error) {
	datos.Nombre = strings.TrimSpace(datos.Nombre)
	datos.Distrito = strings.TrimSpace(datos.Distrito)
	campos := make(map[string]string)
	if !textoValido(datos.Nombre, 160) {
		campos["nombre"] = "Indica el nombre del colegio (hasta 160 caracteres)."
	}
	if !textoValido(datos.Distrito, 100) {
		campos["distrito"] = "Indica el distrito (hasta 100 caracteres)."
	}
	if ahora.IsZero() {
		campos["solicitado_en"] = "No se pudo registrar la fecha de solicitud."
	}
	if len(campos) > 0 {
		return SolicitudColegio{}, ErrorValidacion{Campos: campos}
	}
	return SolicitudColegio{
		Nombre: compactarEspacios(datos.Nombre), Distrito: compactarEspacios(datos.Distrito),
		Estado: EstadoSolicitudPendiente, SolicitadoEn: ahora.UTC(),
	}, nil
}

func (c *Curso) Revisar(estado EstadoCurso, revisorID, motivo string, ahora time.Time) error {
	if c == nil || c.ID <= 0 || strings.TrimSpace(revisorID) == "" || ahora.IsZero() {
		return ErrCursoInvalido
	}
	if c.Estado != EstadoPendiente {
		return ErrEstadoInvalido
	}
	switch estado {
	case EstadoAprobado:
		c.Motivo = ""
	case EstadoRechazado:
		c.Motivo = strings.TrimSpace(motivo)
		if c.Motivo == "" || len([]rune(c.Motivo)) > 500 {
			return ErrorValidacion{Campos: map[string]string{"motivo": "Indica un motivo de hasta 500 caracteres."}}
		}
	default:
		return ErrEstadoInvalido
	}
	revisadoEn := ahora.UTC()
	c.Estado = estado
	c.RevisadoPor = strings.TrimSpace(revisorID)
	c.RevisadoEn = &revisadoEn
	return nil
}

func textoValido(texto string, maximo int) bool {
	if texto == "" || len([]rune(texto)) > maximo {
		return false
	}
	for _, caracter := range texto {
		if unicode.IsControl(caracter) {
			return false
		}
	}
	return true
}

func claveNormalizada(texto string) string {
	return strings.ToLower(compactarEspacios(texto))
}

func compactarEspacios(texto string) string {
	return strings.Join(strings.Fields(texto), " ")
}
