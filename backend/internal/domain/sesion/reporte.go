package sesion

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
)

// I18: reporte de minutos con trazabilidad y escolares seudonimizados.
//
// El reporte se arma solo con datos del flujo E (sesion, asistencia,
// resultado, minutos y auditoria de I17). El escolar nunca aparece con su ID
// ni su nombre: se muestra un seudonimo estable derivado con HMAC. Cada
// descarga queda registrada con la huella SHA-256 del archivo entregado. El
// reporte no tiene validez oficial: no existe conexion con un registro
// externo (RF-18 sin conexion ficticia, E01 pendiente).

// MaxDiasReporte acota el periodo de un reporte.
const MaxDiasReporte = 31

// MinBytesClaveSeudonimo es el largo minimo de la clave del HMAC.
const MinBytesClaveSeudonimo = 16

// PrefijoSeudonimo antecede al seudonimo del escolar.
const PrefijoSeudonimo = "ESC-"

// largoSeudonimo es la cantidad de caracteres hexadecimales del seudonimo.
const largoSeudonimo = 10

// AvisoSinValidezOficial acompana siempre al reporte (RF-18).
const AvisoSinValidezOficial = "Sin validez oficial: no existe convenio ni registro externo verificado con la UNI (E01 pendiente)."

// EstadoSinRegistro marca en el reporte una sesion sin minutos propuestos.
const EstadoSinRegistro EstadoMinutos = "sin_registro"

var (
	ErrFiltroReporteInvalido  = errors.New("filtro de reporte invalido")
	ErrFormatoReporteInvalido = errors.New("formato de reporte invalido")
	ErrClaveSeudonimoInvalida = errors.New("clave de seudonimo invalida")
	ErrReporteInvalido        = errors.New("reporte invalido")
)

// FormatoReporte es el tipo de archivo entregado.
type FormatoReporte string

const (
	FormatoCSV FormatoReporte = "csv"
	FormatoPDF FormatoReporte = "pdf"
)

// ValidarFormatoReporte acepta solo csv o pdf.
func ValidarFormatoReporte(valor string) (FormatoReporte, error) {
	switch formato := FormatoReporte(strings.ToLower(strings.TrimSpace(valor))); formato {
	case FormatoCSV, FormatoPDF:
		return formato, nil
	}
	return "", ErrFormatoReporteInvalido
}

// FiltroReporte delimita las sesiones: Desde <= inicio < Hasta.
type FiltroReporte struct {
	desde   time.Time
	hasta   time.Time
	tutorID string
}

// NuevoFiltroReporte exige un periodo valido de hasta MaxDiasReporte dias.
func NuevoFiltroReporte(desde, hasta time.Time, tutorID string) (FiltroReporte, error) {
	if desde.IsZero() || hasta.IsZero() || !desde.Before(hasta) {
		return FiltroReporte{}, ErrFiltroReporteInvalido
	}
	if hasta.Sub(desde) > MaxDiasReporte*24*time.Hour {
		return FiltroReporte{}, ErrFiltroReporteInvalido
	}
	return FiltroReporte{desde: desde.UTC(), hasta: hasta.UTC(), tutorID: strings.TrimSpace(tutorID)}, nil
}

func (f FiltroReporte) Desde() time.Time { return f.desde }
func (f FiltroReporte) Hasta() time.Time { return f.hasta }
func (f FiltroReporte) TutorID() string  { return f.tutorID }

// Incluye indica si una sesion entra en el filtro.
func (f FiltroReporte) Incluye(s *Sesion) bool {
	if s == nil || s.Inicio().Before(f.desde) || !s.Inicio().Before(f.hasta) {
		return false
	}
	return f.tutorID == "" || s.TutorID() == f.tutorID
}

// Seudonimizador deriva un seudonimo estable e irreversible sin la clave.
type Seudonimizador struct {
	clave []byte
}

func NuevoSeudonimizador(clave []byte) (Seudonimizador, error) {
	if len(clave) < MinBytesClaveSeudonimo {
		return Seudonimizador{}, ErrClaveSeudonimoInvalida
	}
	return Seudonimizador{clave: append([]byte(nil), clave...)}, nil
}

// Seudonimo devuelve ESC- y 10 caracteres hexadecimales de HMAC-SHA256.
func (s Seudonimizador) Seudonimo(usuarioID string) string {
	mac := hmac.New(sha256.New, s.clave)
	mac.Write([]byte(usuarioID))
	return PrefijoSeudonimo + hex.EncodeToString(mac.Sum(nil))[:largoSeudonimo]
}

// AsistenciaReporte resume lo declarado en I05/I11.
type AsistenciaReporte string

const (
	AsistenciaCompleta    AsistenciaReporte = "completa"
	AsistenciaSoloTutor   AsistenciaReporte = "solo_tutor"
	AsistenciaSoloEscolar AsistenciaReporte = "solo_escolar"
	AsistenciaSinDeclarar AsistenciaReporte = "sin_declarar"
	AsistenciaConAusencia AsistenciaReporte = "ausencia_reportada"
)

func resumirAsistencia(s *Sesion) AsistenciaReporte {
	if len(s.Ausencias()) > 0 {
		return AsistenciaConAusencia
	}
	tutor, escolar := false, false
	for _, a := range s.Asistencias() {
		switch a.Rol() {
		case RolTutor:
			tutor = true
		case RolEscolar:
			escolar = true
		}
	}
	switch {
	case tutor && escolar:
		return AsistenciaCompleta
	case tutor:
		return AsistenciaSoloTutor
	case escolar:
		return AsistenciaSoloEscolar
	}
	return AsistenciaSinDeclarar
}

// AccionReporte es la ultima entrada de auditoria de I17 de una sesion.
type AccionReporte struct {
	Accion AccionMinutos
	Actor  string
	En     time.Time
}

// FilaReporte es una sesion del periodo.
type FilaReporte struct {
	SesionID            string
	ReservaID           string
	Inicio              time.Time
	TutorID             string
	TutorNombre         string
	EscolarSeudonimo    string
	Asistencia          AsistenciaReporte
	ResultadoRegistrado bool
	EstadoMinutos       EstadoMinutos
	MinutosPropuestos   int
	MinutosComputables  int
	Ajustes             int
	UltimaAccion        *AccionReporte
}

// TotalTutorReporte resume un tutor: solo los verificados suman.
type TotalTutorReporte struct {
	TutorID            string
	TutorNombre        string
	Sesiones           int
	MinutosVerificados int
	PorEstado          map[EstadoMinutos]int
}

// Reporte es el contenido completo, independiente del formato.
type Reporte struct {
	ID          string
	Filtro      FiltroReporte
	GeneradoPor string
	// GeneradoPorNombre es el nombre visible del coordinador ("" si no se conoce).
	GeneradoPorNombre string
	GeneradoEn        time.Time
	Aviso             string
	Filas             []FilaReporte
	Totales           []TotalTutorReporte
}

// EntradaReporte agrupa lo que el reporte necesita de una sesion.
type EntradaReporte struct {
	Sesion  *Sesion
	Minutos *RegistroMinutos // nil si el tutor aun no propuso minutos
}

// ConstruirReporte arma filas y totales. Descarta sesiones fuera del filtro y
// ordena por inicio. nombres mapea tutorID a nombre visible.
func ConstruirReporte(
	id string,
	filtro FiltroReporte,
	generadoPor string,
	ahora time.Time,
	entradas []EntradaReporte,
	nombres map[string]string,
	seudonimos Seudonimizador,
) (*Reporte, error) {
	id, generadoPor = strings.TrimSpace(id), strings.TrimSpace(generadoPor)
	if id == "" || generadoPor == "" || ahora.IsZero() || filtro.desde.IsZero() || len(seudonimos.clave) == 0 {
		return nil, ErrReporteInvalido
	}
	reporte := &Reporte{
		ID:          id,
		Filtro:      filtro,
		GeneradoPor: generadoPor,
		GeneradoEn:  ahora.UTC(),
		Aviso:       AvisoSinValidezOficial,
		Filas:       []FilaReporte{},
		Totales:     []TotalTutorReporte{},
	}
	totales := map[string]*TotalTutorReporte{}
	for _, entrada := range entradas {
		s := entrada.Sesion
		if !filtro.Incluye(s) {
			continue
		}
		if entrada.Minutos != nil && entrada.Minutos.SesionID() != s.ID() {
			return nil, ErrReporteInvalido
		}
		fila := construirFila(s, entrada.Minutos, nombres[s.TutorID()], seudonimos)
		reporte.Filas = append(reporte.Filas, fila)

		total, existe := totales[fila.TutorID]
		if !existe {
			total = &TotalTutorReporte{TutorID: fila.TutorID, TutorNombre: fila.TutorNombre, PorEstado: map[EstadoMinutos]int{}}
			totales[fila.TutorID] = total
		}
		total.Sesiones++
		total.MinutosVerificados += fila.MinutosComputables
		total.PorEstado[fila.EstadoMinutos]++
	}
	sort.SliceStable(reporte.Filas, func(i, j int) bool {
		a, b := reporte.Filas[i], reporte.Filas[j]
		if !a.Inicio.Equal(b.Inicio) {
			return a.Inicio.Before(b.Inicio)
		}
		return a.SesionID < b.SesionID
	})
	for _, total := range totales {
		reporte.Totales = append(reporte.Totales, *total)
	}
	sort.Slice(reporte.Totales, func(i, j int) bool {
		a, b := reporte.Totales[i], reporte.Totales[j]
		if a.TutorNombre != b.TutorNombre {
			return a.TutorNombre < b.TutorNombre
		}
		return a.TutorID < b.TutorID
	})
	return reporte, nil
}

func construirFila(s *Sesion, minutos *RegistroMinutos, nombre string, seudonimos Seudonimizador) FilaReporte {
	if strings.TrimSpace(nombre) == "" {
		nombre = s.TutorID()
	}
	fila := FilaReporte{
		SesionID:            s.ID(),
		ReservaID:           s.ReservaID(),
		Inicio:              s.Inicio().UTC(),
		TutorID:             s.TutorID(),
		TutorNombre:         nombre,
		EscolarSeudonimo:    seudonimos.Seudonimo(s.EscolarID()),
		Asistencia:          resumirAsistencia(s),
		ResultadoRegistrado: s.Resultado() != nil,
		EstadoMinutos:       EstadoSinRegistro,
	}
	if minutos == nil {
		return fila
	}
	fila.EstadoMinutos = minutos.Estado()
	fila.MinutosPropuestos = minutos.MinutosPropuestos()
	fila.MinutosComputables = minutos.MinutosComputables()
	auditoria := minutos.Auditoria()
	for _, entrada := range auditoria {
		if entrada.Accion() == AccionAjustar || entrada.Accion() == AccionResolver {
			fila.Ajustes++
		}
	}
	if n := len(auditoria); n > 0 {
		ultima := auditoria[n-1]
		fila.UltimaAccion = &AccionReporte{Accion: ultima.Accion(), Actor: ultima.ActorID(), En: ultima.RegistradaEn().UTC()}
	}
	return fila
}

// RegistroReporte deja constancia de cada descarga (trazabilidad, RF-17).
type RegistroReporte struct {
	ID          string
	GeneradoPor string
	GeneradoEn  time.Time
	Desde       time.Time
	Hasta       time.Time
	TutorID     string
	Formato     FormatoReporte
	Filas       int
	SHA256      string
}

// NuevoRegistroReporte calcula la huella del archivo exacto que se entrega.
func NuevoRegistroReporte(reporte *Reporte, formato FormatoReporte, contenido []byte) (RegistroReporte, error) {
	if reporte == nil || len(contenido) == 0 {
		return RegistroReporte{}, ErrReporteInvalido
	}
	if _, err := ValidarFormatoReporte(string(formato)); err != nil {
		return RegistroReporte{}, err
	}
	huella := sha256.Sum256(contenido)
	return RegistroReporte{
		ID:          reporte.ID,
		GeneradoPor: reporte.GeneradoPor,
		GeneradoEn:  reporte.GeneradoEn,
		Desde:       reporte.Filtro.desde,
		Hasta:       reporte.Filtro.hasta,
		TutorID:     reporte.Filtro.tutorID,
		Formato:     formato,
		Filas:       len(reporte.Filas),
		SHA256:      hex.EncodeToString(huella[:]),
	}, nil
}

// RepositorioReportes es el puerto de lectura y registro de I18.
type RepositorioReportes interface {
	// SesionesEntre devuelve las sesiones cuyo inicio cae en el filtro.
	SesionesEntre(ctx context.Context, filtro FiltroReporte) ([]*Sesion, error)
	// RegistrarReporte inserta la constancia; nunca se modifica ni borra.
	RegistrarReporte(ctx context.Context, registro RegistroReporte) error
	// ReportesRecientes lista las ultimas constancias, la mas nueva primero.
	ReportesRecientes(ctx context.Context, limite int) ([]RegistroReporte, error)
}

// DirectorioNombres traduce un usuario a su nombre visible. Lo implementa un
// adaptador sobre el flujo A (acceso); "" si no se conoce.
type DirectorioNombres interface {
	NombreDe(ctx context.Context, usuarioID string) (string, error)
}
