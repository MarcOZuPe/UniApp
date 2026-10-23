package tutor

import (
	"context"
	"time"
)

type Repositorio interface {
	Guardar(context.Context, *Expediente) error
	PorTutor(context.Context, string) (*Expediente, error)
	Pendientes(context.Context) ([]Expediente, error)
	Revisar(context.Context, *Expediente) error
	Documento(context.Context, string) (Documento, string, error)
	PerfilesPublicos(context.Context, time.Time) ([]PerfilListado, error)
}

type AlmacenDocumentos interface {
	Guardar(context.Context, string, []byte) error
	Leer(context.Context, string) ([]byte, error)
	Eliminar(context.Context, string) error
}
