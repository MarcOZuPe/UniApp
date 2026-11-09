package sesion

import (
	"context"
	"errors"
	"time"
)

var (
	ErrSesionNoEncontrada = errors.New("sesion no encontrada")
	ErrSesionDuplicada    = errors.New("sesion ya registrada")
)

type Repositorio interface {
	CrearSesion(
		ctx context.Context,
		sesion *Sesion,
	) error

	SesionPorID(
		ctx context.Context,
		id string,
	) (*Sesion, error)

	SesionPorReservaID(
		ctx context.Context,
		reservaID string,
	) (*Sesion, error)

	RegistrarResultado(
		ctx context.Context,
		sesionID string,
		usuarioID string,
		resumen string,
		ahora time.Time,
	) (Resultado, error)

	ReportarAusencia(
		ctx context.Context,
		sesionID string,
		usuarioID string,
		ahora time.Time,
	) (Ausencia, error)

	DeclararAsistencia(
		ctx context.Context,
		sesionID string,
		usuarioID string,
		ahora time.Time,
	) (Asistencia, error)
}
