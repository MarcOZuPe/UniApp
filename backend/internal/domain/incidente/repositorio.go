package incidente

import (
	"context"
	"time"
)

type Repositorio interface {
	TutoriasDeEscolar(ctx context.Context, escolarID string, ahora time.Time) ([]Tutoria, error)
	TutoriaDeEscolar(ctx context.Context, escolarID, reservaID string, ahora time.Time) (*Tutoria, error)
	EscolaresDeApoderado(ctx context.Context, apoderadoID string) ([]EscolarVinculado, error)
	EsApoderadoDe(ctx context.Context, apoderadoID, escolarID string) (bool, error)

	Crear(ctx context.Context, incidente *Incidente) (*Incidente, error)
	PorID(ctx context.Context, incidenteID string) (*Incidente, error)
	Listar(ctx context.Context, filtro Filtro) ([]Incidente, error)
	Revisar(ctx context.Context, incidenteID, revisorID string, resultado Resultado, comentario string, ahora time.Time) (*Incidente, error)
	EvidenciaDe(ctx context.Context, incidenteID, evidenciaID string) (Evidencia, error)
}

type Almacen interface {
	Guardar(ctx context.Context, clave string, contenido []byte) error
	Leer(ctx context.Context, clave string) ([]byte, error)
	Eliminar(ctx context.Context, clave string) error
}
