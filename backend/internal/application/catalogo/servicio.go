package catalogo

import (
	"context"
	"strings"
	"time"
)

type Servicio struct {
	repositorio Repositorio
	ahora       func() time.Time
}

func NuevoServicio(repositorio Repositorio) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		ahora:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Servicio) Importar(ctx context.Context, datos DatosImportacion) (*Importacion, error) {
	importacion, err := NuevaImportacion(datos, s.ahora())
	if err != nil {
		return nil, err
	}
	if err := s.repositorio.Importar(ctx, importacion); err != nil {
		return nil, err
	}
	return CopiarImportacion(importacion), nil
}

func (s *Servicio) CatalogoPublico(ctx context.Context) (*Importacion, error) {
	importacion, err := s.repositorio.ImportacionActual(ctx)
	if err != nil {
		return nil, err
	}
	publica := CopiarImportacion(importacion)
	publica.Cursos = make([]Curso, 0, len(importacion.Cursos))
	for _, curso := range importacion.Cursos {
		if curso.Estado == EstadoAprobado {
			publica.Cursos = append(publica.Cursos, curso)
		}
	}
	return publica, nil
}

func (s *Servicio) CursosPendientes(ctx context.Context) ([]Curso, error) {
	importacion, err := s.repositorio.ImportacionActual(ctx)
	if err != nil {
		return nil, err
	}
	pendientes := make([]Curso, 0)
	for _, curso := range importacion.Cursos {
		if curso.Estado == EstadoPendiente {
			pendientes = append(pendientes, curso)
		}
	}
	return pendientes, nil
}

func (s *Servicio) RevisarCurso(ctx context.Context, cursoID int64, revisorID string, estado EstadoCurso, motivo string) (Curso, error) {
	curso, err := s.repositorio.Curso(ctx, cursoID)
	if err != nil {
		return Curso{}, err
	}
	if err := curso.Revisar(estado, revisorID, motivo, s.ahora()); err != nil {
		return Curso{}, err
	}
	if err := s.repositorio.RevisarCurso(ctx, curso, revisorID); err != nil {
		return Curso{}, err
	}
	return curso, nil
}

func (s *Servicio) SolicitarColegio(ctx context.Context, requesterID string, datos SolicitudColegioEntrada) (SolicitudColegio, error) {
	if strings.TrimSpace(requesterID) == "" {
		return SolicitudColegio{}, ErrSolicitudInvalida
	}
	colegioSolicitado, err := NuevaSolicitudColegio(datos, s.ahora())
	if err != nil {
		return SolicitudColegio{}, err
	}
	importacion, err := s.repositorio.ImportacionActual(ctx)
	if err != nil {
		return SolicitudColegio{}, err
	}
	for _, colegio := range importacion.Colegios {
		if ClaveNormalizada(colegio.Nombre)+"\x00"+ClaveNormalizada(colegio.Distrito) ==
			ClaveNormalizada(colegioSolicitado.Nombre)+"\x00"+ClaveNormalizada(colegioSolicitado.Distrito) {
			return SolicitudColegio{}, ErrColegioYaExiste
		}
	}
	if err := s.repositorio.SolicitarColegio(ctx, colegioSolicitado, requesterID); err != nil {
		return SolicitudColegio{}, err
	}
	return colegioSolicitado, nil
}

func (s *Servicio) SolicitudesPendientes(ctx context.Context) ([]SolicitudColegio, error) {
	return s.repositorio.SolicitudesPendientes(ctx)
}
