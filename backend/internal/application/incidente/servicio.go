package incidente

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	modeloAcceso "uniapp/internal/application/acceso"
)

type Servicio struct {
	repositorio Repositorio
	almacen     Almacen
	ahora       func() time.Time
}

func NuevoServicio(repositorio Repositorio, almacen Almacen) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		almacen:     almacen,
		ahora:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Servicio) FijarReloj(f func() time.Time) { s.ahora = f }

func (s *Servicio) Ahora() time.Time { return s.ahora().UTC() }

func (s *Servicio) Tutorias(ctx context.Context, solicitante Solicitante) ([]Tutoria, error) {
	alcance, err := s.alcance(solicitante)
	if err != nil {
		return nil, err
	}
	return alcance.Tutorias(ctx)
}

func (s *Servicio) Registrar(ctx context.Context, solicitante Solicitante, datos DatosIncidente) (*Incidente, error) {
	if !solicitante.identificado() {
		return nil, ErrNoAutorizado
	}
	switch solicitante.Rol {
	case modeloAcceso.RolEscolar:
		datos.EscolarID = solicitante.ID
	case modeloAcceso.RolApoderado:
		datos.EscolarID = strings.TrimSpace(datos.EscolarID)
		if datos.EscolarID == "" {
			return nil, ErrorValidacion{Campos: map[string]string{
				"escolar_id": "Elige al escolar afectado.",
			}}
		}
		vinculado, err := s.repositorio.EsApoderadoDe(ctx, solicitante.ID, datos.EscolarID)
		if err != nil {
			return nil, err
		}
		if !vinculado {
			return nil, ErrEscolarNoVinculado
		}
	default:
		return nil, ErrNoAutorizado
	}
	datos.ReportadoPor = solicitante.ID

	tutoria, err := s.repositorio.TutoriaDeEscolar(ctx, datos.EscolarID, strings.TrimSpace(datos.ReservaID), s.Ahora())
	if err != nil {
		if errors.Is(err, ErrTutoriaNoEncontrada) {
			return nil, ErrTutoriaNoEncontrada
		}
		return nil, err
	}
	incidente, err := NuevoIncidente(datos, tutoria, s.Ahora())
	if err != nil {
		return nil, err
	}
	if s.almacen == nil {
		return nil, ErrAlmacen
	}

	guardadas := make([]string, 0, len(incidente.Evidencias))
	for indice := range incidente.Evidencias {
		clave, err := nuevaClaveAlmacen()
		if err != nil {
			return nil, errors.Join(err, s.eliminarEvidencias(context.Background(), guardadas))
		}
		if err := s.almacen.Guardar(ctx, clave, datos.Evidencias[indice].Contenido); err != nil {
			return nil, errors.Join(ErrAlmacen, s.eliminarEvidencias(context.Background(), guardadas))
		}
		guardadas = append(guardadas, clave)
		incidente.Evidencias[indice].ClaveAlmacenamiento = clave
	}

	creado, err := s.repositorio.Crear(ctx, incidente)
	if err != nil {
		return nil, errors.Join(err, s.eliminarEvidencias(context.Background(), guardadas))
	}
	return creado, nil
}

func (s *Servicio) Listar(ctx context.Context, solicitante Solicitante) ([]Incidente, error) {
	alcance, err := s.alcance(solicitante)
	if err != nil {
		return nil, err
	}
	filtro, aplicable, err := alcance.Filtrar(ctx)
	if err != nil {
		return nil, err
	}
	if !aplicable {
		return []Incidente{}, nil
	}
	return s.repositorio.Listar(ctx, filtro)
}

func (s *Servicio) PorID(ctx context.Context, solicitante Solicitante, incidenteID string) (*Incidente, error) {
	alcance, err := s.alcance(solicitante)
	if err != nil {
		return nil, err
	}
	incidente, err := s.repositorio.PorID(ctx, incidenteID)
	if err != nil {
		return nil, err
	}
	visible, err := alcance.Visible(ctx, incidente)
	if err != nil {
		return nil, err
	}
	if !visible {
		return nil, ErrNoAutorizado
	}
	return incidente, nil
}

func (s *Servicio) AbrirEvidencia(
	ctx context.Context,
	solicitante Solicitante,
	incidenteID, evidenciaID string,
) (*Incidente, Evidencia, []byte, error) {
	incidente, err := s.PorID(ctx, solicitante, incidenteID)
	if err != nil {
		return nil, Evidencia{}, nil, err
	}
	evidencia, err := s.repositorio.EvidenciaDe(ctx, incidente.ID, evidenciaID)
	if err != nil {
		return nil, Evidencia{}, nil, err
	}
	if s.almacen == nil {
		return nil, Evidencia{}, nil, ErrAlmacen
	}
	contenido, err := s.almacen.Leer(ctx, evidencia.ClaveAlmacenamiento)
	if err != nil {
		return nil, Evidencia{}, nil, ErrAlmacen
	}
	return incidente, evidencia, contenido, nil
}

func (s *Servicio) Revisar(
	ctx context.Context,
	solicitante Solicitante,
	incidenteID string,
	datos DatosRevision,
) (*Incidente, error) {
	if !solicitante.Rol.EsCoordinacion() {
		return nil, ErrNoAutorizado
	}
	incidente, err := s.PorID(ctx, solicitante, incidenteID)
	if err != nil {
		return nil, err
	}
	if err := incidente.Revisar(solicitante.ID, datos, s.Ahora()); err != nil {
		return nil, err
	}
	return s.repositorio.Revisar(
		ctx, incidente.ID, solicitante.ID, incidente.Resultado,
		incidente.ComentarioRevision, *incidente.RevisadoEn,
	)
}

func (s *Servicio) alcance(solicitante Solicitante) (alcance, error) {
	if !solicitante.identificado() {
		return nil, ErrNoAutorizado
	}
	return nuevoAlcance(solicitante, s.repositorio, s.Ahora())
}

func (s *Servicio) eliminarEvidencias(ctx context.Context, claves []string) error {
	var errores []error
	for _, clave := range claves {
		if err := s.almacen.Eliminar(ctx, clave); err != nil {
			errores = append(errores, err)
		}
	}
	return errors.Join(errores...)
}

func nuevaClaveAlmacen() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", ErrAlmacen
	}
	return hex.EncodeToString(bytes), nil
}
