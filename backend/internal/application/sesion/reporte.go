package sesion

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	dominio "uniapp/internal/domain/sesion"
)

// LimiteRegistroReportes acota la lista de constancias que se devuelve.
const LimiteRegistroReportes = 50

// ServicioReporte orquesta I18. Las reglas (filtro, seudonimo, totales y
// huella) viven en el dominio; aqui se juntan las sesiones, los minutos de
// I17 (sincronizados con las cancelaciones) y los nombres de los tutores.
type ServicioReporte struct {
	Reportes   RepositorioReportes
	Minutos    ServicioMinutos
	Nombres    DirectorioNombres
	Seudonimos Seudonimizador
	// NuevoID genera el identificador del reporte; si es nil se usa uno aleatorio.
	NuevoID func() (string, error)
}

// Generar arma el reporte del periodo. No lo registra: la constancia se guarda
// con RegistrarDescarga, una vez generado el archivo exacto que se entrega.
func (s ServicioReporte) Generar(ctx context.Context, filtro FiltroReporte, coordinadorID string, ahora time.Time) (*Reporte, error) {
	sesiones, err := s.Reportes.SesionesEntre(ctx, filtro)
	if err != nil {
		return nil, err
	}
	entradas := make([]EntradaReporte, 0, len(sesiones))
	nombres := map[string]string{}
	for _, sesion := range sesiones {
		minutos, err := s.Minutos.Consultar(ctx, sesion.ID(), ahora)
		switch {
		case errors.Is(err, dominio.ErrMinutosNoEncontrados):
			minutos = nil
		case err != nil:
			return nil, err
		}
		entradas = append(entradas, EntradaReporte{Sesion: sesion, Minutos: minutos})

		if _, conocido := nombres[sesion.TutorID()]; !conocido {
			nombre, err := s.nombreDe(ctx, sesion.TutorID())
			if err != nil {
				return nil, err
			}
			nombres[sesion.TutorID()] = nombre
		}
	}
	id, err := s.generarID()
	if err != nil {
		return nil, err
	}
	reporte, err := dominio.ConstruirReporte(id, filtro, coordinadorID, ahora, entradas, nombres, s.Seudonimos)
	if err != nil {
		return nil, err
	}
	if reporte.GeneradoPorNombre, err = s.nombreDe(ctx, coordinadorID); err != nil {
		return nil, err
	}
	return reporte, nil
}

// RegistrarDescarga guarda la constancia con la huella del archivo. Si falla,
// el archivo no debe entregarse (sin constancia no hay trazabilidad).
func (s ServicioReporte) RegistrarDescarga(ctx context.Context, reporte *Reporte, formato FormatoReporte, contenido []byte) (RegistroReporte, error) {
	registro, err := dominio.NuevoRegistroReporte(reporte, formato, contenido)
	if err != nil {
		return RegistroReporte{}, err
	}
	if err := s.Reportes.RegistrarReporte(ctx, registro); err != nil {
		return RegistroReporte{}, err
	}
	return registro, nil
}

// Recientes devuelve las ultimas constancias de descarga.
func (s ServicioReporte) Recientes(ctx context.Context) ([]RegistroReporte, error) {
	return s.Reportes.ReportesRecientes(ctx, LimiteRegistroReportes)
}

func (s ServicioReporte) nombreDe(ctx context.Context, usuarioID string) (string, error) {
	if s.Nombres == nil {
		return "", nil
	}
	return s.Nombres.NombreDe(ctx, usuarioID)
}

func (s ServicioReporte) generarID() (string, error) {
	if s.NuevoID != nil {
		return s.NuevoID()
	}
	aleatorio := make([]byte, 8)
	if _, err := rand.Read(aleatorio); err != nil {
		return "", err
	}
	return "rep-" + hex.EncodeToString(aleatorio), nil
}
