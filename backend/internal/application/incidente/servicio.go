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

// Tutorias devuelve las sesiones ya finalizadas del escolar. Para el apoderado se
// incluyen las de todos sus escolares vinculados y el nombre de cada escolar.
func (s *Servicio) Tutorias(ctx context.Context, usuarioID string, rol modeloAcceso.Rol) ([]Tutoria, error) {
	switch rol {
	case modeloAcceso.RolEscolar:
		return s.repositorio.TutoriasDeEscolar(ctx, usuarioID, s.Ahora())
	case modeloAcceso.RolApoderado:
		escolares, err := s.repositorio.EscolaresDeApoderado(ctx, usuarioID)
		if err != nil {
			return nil, err
		}
		tutorias := make([]Tutoria, 0)
		for _, escolar := range escolares {
			delEscolar, err := s.repositorio.TutoriasDeEscolar(ctx, escolar.ID, s.Ahora())
			if err != nil {
				return nil, err
			}
			tutorias = append(tutorias, delEscolar...)
		}
		return tutorias, nil
	case modeloAcceso.RolCoordinador, modeloAcceso.RolCoordinadorGeneral:
		return []Tutoria{}, nil
	default:
		return nil, ErrNoAutorizado
	}
}

func (s *Servicio) Registrar(ctx context.Context, usuarioID string, rol modeloAcceso.Rol, datos DatosIncidente) (*Incidente, error) {
	if strings.TrimSpace(usuarioID) == "" {
		return nil, ErrNoAutorizado
	}
	switch rol {
	case modeloAcceso.RolEscolar:
		datos.EscolarID = usuarioID
	case modeloAcceso.RolApoderado:
		datos.EscolarID = strings.TrimSpace(datos.EscolarID)
		if datos.EscolarID == "" {
			return nil, ErrorValidacion{Campos: map[string]string{
				"escolar_id": "Elige al escolar afectado.",
			}}
		}
		vinculado, err := s.repositorio.EsApoderadoDe(ctx, usuarioID, datos.EscolarID)
		if err != nil {
			return nil, err
		}
		if !vinculado {
			return nil, ErrEscolarNoVinculado
		}
	default:
		return nil, ErrNoAutorizado
	}
	datos.ReportadoPor = usuarioID

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

func (s *Servicio) Listar(
	ctx context.Context,
	usuarioID string,
	rol modeloAcceso.Rol,
	facultadID string,
) ([]Incidente, error) {
	switch rol {
	case modeloAcceso.RolCoordinador, modeloAcceso.RolCoordinadorGeneral:
		if facultadID == "" {
			return []Incidente{}, nil
		}
		return s.repositorio.Listar(ctx, Filtro{FacultadID: facultadID})
	case modeloAcceso.RolEscolar:
		return s.repositorio.Listar(ctx, Filtro{EscolarID: usuarioID})
	case modeloAcceso.RolApoderado:
		escolares, err := s.repositorio.EscolaresDeApoderado(ctx, usuarioID)
		if err != nil {
			return nil, err
		}
		ids := make([]string, 0, len(escolares))
		for _, escolar := range escolares {
			ids = append(ids, escolar.ID)
		}
		if len(ids) == 0 {
			return []Incidente{}, nil
		}
		return s.repositorio.Listar(ctx, Filtro{EscolaresID: ids})
	default:
		return nil, ErrNoAutorizado
	}
}

func (s *Servicio) PorID(
	ctx context.Context,
	usuarioID string,
	rol modeloAcceso.Rol,
	facultadID, incidenteID string,
) (*Incidente, error) {
	incidente, err := s.repositorio.PorID(ctx, incidenteID)
	if err != nil {
		return nil, err
	}
	visible, err := s.puedeVer(ctx, incidente, usuarioID, rol, facultadID)
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
	usuarioID string,
	rol modeloAcceso.Rol,
	facultadID, incidenteID, evidenciaID string,
) (*Incidente, Evidencia, []byte, error) {
	incidente, err := s.PorID(ctx, usuarioID, rol, facultadID, incidenteID)
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
	revisorID string,
	rol modeloAcceso.Rol,
	facultadID, incidenteID string,
	datos DatosRevision,
) (*Incidente, error) {
	if !rol.EsCoordinacion() {
		return nil, ErrNoAutorizado
	}
	incidente, err := s.PorID(ctx, revisorID, rol, facultadID, incidenteID)
	if err != nil {
		return nil, err
	}
	if err := incidente.Revisar(revisorID, datos, s.Ahora()); err != nil {
		return nil, err
	}
	return s.repositorio.Revisar(
		ctx, incidente.ID, revisorID, incidente.Resultado,
		incidente.ComentarioRevision, *incidente.RevisadoEn,
	)
}

func (s *Servicio) puedeVer(
	ctx context.Context,
	incidente *Incidente,
	usuarioID string,
	rol modeloAcceso.Rol,
	facultadID string,
) (bool, error) {
	switch rol {
	case modeloAcceso.RolCoordinador, modeloAcceso.RolCoordinadorGeneral:
		return facultadID != "" && incidente.FacultadID == facultadID, nil
	case modeloAcceso.RolEscolar:
		return incidente.EscolarID == usuarioID, nil
	case modeloAcceso.RolApoderado:
		return s.repositorio.EsApoderadoDe(ctx, usuarioID, incidente.EscolarID)
	default:
		return false, nil
	}
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
