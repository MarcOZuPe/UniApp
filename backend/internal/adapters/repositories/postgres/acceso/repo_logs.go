package acceso

import (
	"context"
	"encoding/json"
	"fmt"
)

func (repositorio RepositorioPostgres) RegistrarLogSistema(ctx context.Context, log LogSistema) error {
	detalle := log.Detalle
	if detalle == nil {
		detalle = map[string]any{}
	}
	crudo, err := json.Marshal(detalle)
	if err != nil {
		return err
	}
	_, err = repositorio.DB.ExecContext(ctx, `
		INSERT INTO logs_sistema (categoria, evento, actor_id, actor_rol, ip, detalle)
		VALUES ($1, $2, NULLIF($3,'')::uuid, $4, $5, $6)`,
		log.Categoria, log.Evento, log.ActorID, log.ActorRol, log.IP, crudo,
	)
	return err
}

func (repositorio RepositorioPostgres) LogsSistema(
	ctx context.Context,
	filtro FiltroLogsSistema,
) ([]LogSistema, error) {
	filas, err := repositorio.DB.QueryContext(ctx, `
		SELECT id, creado_en, categoria, evento, COALESCE(actor_id::text, ''),
		       actor_rol, ip, detalle
		FROM logs_sistema
		ORDER BY creado_en DESC, id DESC
		LIMIT $1`, filtro.Limite)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	logs := make([]LogSistema, 0)
	for filas.Next() {
		var (
			log   LogSistema
			crudo []byte
		)
		if err := filas.Scan(
			&log.ID, &log.CreadoEn, &log.Categoria, &log.Evento,
			&log.ActorID, &log.ActorRol, &log.IP, &crudo,
		); err != nil {
			return nil, err
		}
		log.Detalle = map[string]any{}
		if len(crudo) > 0 {
			if err := json.Unmarshal(crudo, &log.Detalle); err != nil {
				return nil, fmt.Errorf("leer detalle de log: %w", err)
			}
		}
		logs = append(logs, log)
	}
	return logs, filas.Err()
}
