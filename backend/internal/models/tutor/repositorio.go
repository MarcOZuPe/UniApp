package tutor

import "context"

type Repositorio interface {
	Guardar(context.Context, *Expediente) error
	PorTutor(context.Context, string) (*Expediente, error)
	Pendientes(context.Context) ([]Expediente, error)
	Revisar(context.Context, *Expediente) error
	Documento(context.Context, string) (Documento, string, error)
}

type AlmacenDocumentos interface {
	Guardar(context.Context, string, []byte) error
	Leer(context.Context, string) ([]byte, error)
	Eliminar(context.Context, string) error
}
