package acceso

import "context"

type LockoutManager interface {
	EstaBloqueado(ctx context.Context, email string) (bool, error)
	RegistrarFallo(ctx context.Context, email string) error
	LimpiarFallos(ctx context.Context, email string) error
}

type lockoutNulo struct{}

func (lockoutNulo) EstaBloqueado(context.Context, string) (bool, error) { return false, nil }
func (lockoutNulo) RegistrarFallo(context.Context, string) error        { return nil }
func (lockoutNulo) LimpiarFallos(context.Context, string) error         { return nil }
