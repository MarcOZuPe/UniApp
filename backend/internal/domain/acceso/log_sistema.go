package acceso

import (
	"context"
	"time"
)

const (
	CategoriaSeguridad      = "seguridad"
	CategoriaAdministracion = "administracion"
	CategoriaSistema        = "sistema"
)

type LogSistema struct {
	ID        int64
	CreadoEn  time.Time
	Categoria string
	Evento    string
	ActorID   string
	ActorRol  string
	IP        string
	Detalle   map[string]any
}

type FiltroLogsSistema struct {
	Limite int
}

type RepositorioLogs interface {
	RegistrarLogSistema(ctx context.Context, log LogSistema) error
	LogsSistema(ctx context.Context, filtro FiltroLogsSistema) ([]LogSistema, error)
}
