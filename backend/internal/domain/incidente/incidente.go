package incidente

import (
	"errors"
	"net/http"
	"path"
	"strings"
	"time"
	"unicode/utf8"
)

// I13: el incidente señala una posible falta del tutor. La evidencia es sensible y
// solo se comparte con el escolar, el apoderado que reporta o consulta y coordinación.
// El tutor denunciado nunca accede a los adjuntos.
const (
	MaxEvidencias         = 5
	EvidenciaMaximaBytes  = 10 << 20
	NombreEvidenciaMaximo = 255
	DescripcionMinima     = 20
	DescripcionMaxima     = 2000
	ComentarioMinimo      = 5
	ComentarioMaximo      = 1000
)

type Tipo string

const (
	TipoTratoInadecuado Tipo = "trato_inadecuado"
	TipoDiscriminacion  Tipo = "discriminacion"
	TipoAcoso           Tipo = "acoso"
	TipoSeguridad       Tipo = "seguridad"
	TipoOtro            Tipo = "otro"
)

var TiposValidos = map[Tipo]string{
	TipoTratoInadecuado: "Trato inadecuado",
	TipoDiscriminacion:  "Discriminación",
	TipoAcoso:           "Acoso o hostigamiento",
	TipoSeguridad:       "Riesgo de seguridad",
	TipoOtro:            "Otro",
}

type Estado string

const (
	EstadoPendiente Estado = "pendiente"
	EstadoRevisado  Estado = "revisado"
)

type Resultado string

const (
	ResultadoProcede   Resultado = "procede"
	ResultadoNoProcede Resultado = "no_procede"
)

type Archivo struct {
	Nombre    string
	Contenido []byte
}

type Evidencia struct {
	ID                  string
	Nombre              string
	Tipo                string
	TamanoBytes         int64
	ClaveAlmacenamiento string
	CreadoEn            time.Time
}

type Tutoria struct {
	ReservaID       string
	EscolarID       string
	EscolarNombre   string
	TutorID         string
	TutorNombre     string
	TutorFacultadID string
	Curso           string
	Inicio          time.Time
	Fin             time.Time
	Asistio         bool
}

type EscolarVinculado struct {
	ID     string
	Nombre string
}

type DatosIncidente struct {
	EscolarID    string
	ReportadoPor string
	ReservaID    string
	Tipo         Tipo
	Descripcion  string
	Evidencias   []Archivo
}

type Incidente struct {
	ID                 string
	EscolarID          string
	EscolarNombre      string
	ReportadoPor       string
	TutorID            string
	TutorNombre        string
	FacultadID         string
	ReservaID          string
	Curso              string
	Inicio             time.Time
	Tipo               Tipo
	Descripcion        string
	Estado             Estado
	Resultado          Resultado
	ComentarioRevision string
	RevisadoPor        string
	RevisadoEn         *time.Time
	CreadoEn           time.Time
	Evidencias         []Evidencia
}

type DatosRevision struct {
	Resultado  Resultado
	Comentario string
}

type Filtro struct {
	// Todos lo usa coordinación. En los demás casos se limita por escolar o por la
	// lista de escolares vinculados al apoderado.
	Todos       bool
	EscolarID   string
	EscolaresID []string
	FacultadID  string
}

var (
	ErrNoAutorizado          = errors.New("no autorizado para acceder a incidentes")
	ErrTutoriaNoEncontrada   = errors.New("la tutoría indicada no pertenece al escolar")
	ErrEscolarNoVinculado    = errors.New("el escolar no está vinculado al apoderado")
	ErrIncidenteNoEncontrado = errors.New("incidente no encontrado")
	ErrEvidenciaNoEncontrada = errors.New("evidencia no encontrada")
	ErrYaRevisado            = errors.New("el incidente ya fue revisado")
	ErrEvidenciaInvalida     = errors.New("evidencia inválida")
	ErrAlmacen               = errors.New("no se pudo guardar la evidencia")
	ErrIncidenteInvalido     = errors.New("incidente inválido")
)

type ErrorValidacion struct {
	Campos map[string]string
}

func (e ErrorValidacion) Error() string { return ErrIncidenteInvalido.Error() }

// PrepararEvidencia valida por contenido y extensión. No confía en el tipo MIME que
// declara el navegador. Devuelve el nombre saneado para usarlo en la descarga.
func PrepararEvidencia(archivo Archivo, ahora time.Time) (Evidencia, error) {
	nombre := nombreSeguro(archivo.Nombre)
	if nombre == "" || !utf8.ValidString(nombre) || utf8.RuneCountInString(nombre) > NombreEvidenciaMaximo {
		return Evidencia{}, ErrEvidenciaInvalida
	}
	if len(archivo.Contenido) == 0 || len(archivo.Contenido) > EvidenciaMaximaBytes {
		return Evidencia{}, ErrEvidenciaInvalida
	}
	tipo, extensiones, permitido := tipoEvidencia(archivo.Contenido)
	if !permitido || !contiene(extensiones, strings.ToLower(path.Ext(nombre))) {
		return Evidencia{}, ErrEvidenciaInvalida
	}
	if ahora.IsZero() {
		ahora = time.Now().UTC()
	}
	return Evidencia{
		Nombre:      nombre,
		Tipo:        tipo,
		TamanoBytes: int64(len(archivo.Contenido)),
		CreadoEn:    ahora.UTC(),
	}, nil
}

func NuevoIncidente(datos DatosIncidente, tutoria *Tutoria, ahora time.Time) (*Incidente, error) {
	datos.EscolarID = strings.TrimSpace(datos.EscolarID)
	datos.ReportadoPor = strings.TrimSpace(datos.ReportadoPor)
	datos.ReservaID = strings.TrimSpace(datos.ReservaID)
	datos.Descripcion = strings.TrimSpace(datos.Descripcion)

	campos := map[string]string{}
	if datos.EscolarID == "" {
		campos["escolar_id"] = "No se pudo identificar al escolar."
	}
	if datos.ReportadoPor == "" {
		campos["reportado_por"] = "No se pudo identificar a quien reporta."
	}
	if datos.ReservaID == "" {
		campos["reserva_id"] = "Elige la tutoría sobre la que reportas el incidente."
	}
	if _, valido := TiposValidos[datos.Tipo]; !valido {
		campos["tipo"] = "Elige un tipo de incidente."
	}
	if longitud := utf8.RuneCountInString(datos.Descripcion); longitud < DescripcionMinima || longitud > DescripcionMaxima {
		campos["descripcion"] = "Describe lo ocurrido entre 20 y 2000 caracteres."
	}
	if len(datos.Evidencias) == 0 || len(datos.Evidencias) > MaxEvidencias {
		campos["evidencias"] = "Adjunta entre 1 y 5 evidencias PDF, JPG o PNG."
	}
	if tutoria == nil || tutoria.ReservaID != datos.ReservaID || tutoria.EscolarID != datos.EscolarID ||
		tutoria.Inicio.IsZero() || tutoria.Fin.IsZero() || tutoria.Fin.After(ahora) {
		campos["reserva_id"] = "La tutoría debe haber finalizado y pertenecer al escolar."
	}
	if len(campos) > 0 {
		return nil, ErrorValidacion{Campos: campos}
	}
	if ahora.IsZero() {
		return nil, ErrIncidenteInvalido
	}

	evidencias := make([]Evidencia, 0, len(datos.Evidencias))
	for _, archivo := range datos.Evidencias {
		evidencia, err := PrepararEvidencia(archivo, ahora)
		if err != nil {
			campos["evidencias"] = "Cada evidencia debe ser PDF, JPG o PNG de hasta 10 MB."
			return nil, ErrorValidacion{Campos: campos}
		}
		evidencias = append(evidencias, evidencia)
	}

	return &Incidente{
		EscolarID:     datos.EscolarID,
		EscolarNombre: tutoria.EscolarNombre,
		ReportadoPor:  datos.ReportadoPor,
		TutorID:       tutoria.TutorID,
		TutorNombre:   tutoria.TutorNombre,
		FacultadID:    tutoria.TutorFacultadID,
		ReservaID:     tutoria.ReservaID,
		Curso:         tutoria.Curso,
		Inicio:        tutoria.Inicio.UTC(),
		Tipo:          datos.Tipo,
		Descripcion:   datos.Descripcion,
		Estado:        EstadoPendiente,
		CreadoEn:      ahora.UTC(),
		Evidencias:    evidencias,
	}, nil
}

func (i *Incidente) Revisar(revisorID string, datos DatosRevision, ahora time.Time) error {
	if i == nil || strings.TrimSpace(revisorID) == "" || ahora.IsZero() {
		return ErrIncidenteInvalido
	}
	if i.Estado != EstadoPendiente {
		return ErrYaRevisado
	}
	datos.Comentario = strings.TrimSpace(datos.Comentario)
	campos := map[string]string{}
	if datos.Resultado != ResultadoProcede && datos.Resultado != ResultadoNoProcede {
		campos["resultado"] = "Indica si el incidente procede o no procede."
	}
	if longitud := utf8.RuneCountInString(datos.Comentario); longitud < ComentarioMinimo || longitud > ComentarioMaximo {
		campos["comentario"] = "Explica la revisión entre 5 y 1000 caracteres."
	}
	if len(campos) > 0 {
		return ErrorValidacion{Campos: campos}
	}
	revisadoEn := ahora.UTC()
	i.Estado = EstadoRevisado
	i.Resultado = datos.Resultado
	i.ComentarioRevision = datos.Comentario
	i.RevisadoPor = revisorID
	i.RevisadoEn = &revisadoEn
	return nil
}

func tipoEvidencia(contenido []byte) (string, []string, bool) {
	tipo := http.DetectContentType(contenido)
	if indice := strings.IndexByte(tipo, ';'); indice >= 0 {
		tipo = strings.TrimSpace(tipo[:indice])
	}
	switch tipo {
	case "application/pdf":
		if !strings.HasPrefix(string(contenido), "%PDF-") {
			return "", nil, false
		}
		return tipo, []string{".pdf"}, true
	case "image/png":
		return tipo, []string{".png"}, true
	case "image/jpeg":
		return tipo, []string{".jpg", ".jpeg"}, true
	default:
		return "", nil, false
	}
}

func nombreSeguro(nombre string) string {
	nombre = strings.ReplaceAll(nombre, `\`, "/")
	partes := strings.Split(nombre, "/")
	nombre = strings.TrimSpace(partes[len(partes)-1])
	if nombre == "" || nombre == "." || nombre == ".." || strings.HasPrefix(nombre, ".") {
		return ""
	}
	if strings.IndexFunc(nombre, func(caracter rune) bool {
		return caracter < 0x20 || caracter == 0x7f
	}) >= 0 {
		return ""
	}
	return nombre
}

func contiene(valores []string, buscado string) bool {
	for _, valor := range valores {
		if valor == buscado {
			return true
		}
	}
	return false
}
