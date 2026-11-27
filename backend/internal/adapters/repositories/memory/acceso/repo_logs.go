package acceso

import (
	"context"
	"time"
)

func (repositorio *RepositorioMemoria) RegistrarLogSistema(_ context.Context, log LogSistema) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	if log.CreadoEn.IsZero() {
		log.CreadoEn = time.Now().UTC()
	}
	repositorio.logsSistema = append(repositorio.logsSistema, log)
	return nil
}

func (repositorio *RepositorioMemoria) LogsSistema(
	_ context.Context,
	filtro FiltroLogsSistema,
) ([]LogSistema, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	logs := make([]LogSistema, 0, len(repositorio.logsSistema))
	for indice := len(repositorio.logsSistema) - 1; indice >= 0; indice-- {
		logs = append(logs, repositorio.logsSistema[indice])
		if filtro.Limite > 0 && len(logs) >= filtro.Limite {
			break
		}
	}
	return logs, nil
}
