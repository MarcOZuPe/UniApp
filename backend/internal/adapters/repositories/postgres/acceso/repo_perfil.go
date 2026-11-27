package acceso

import (
	"context"
	"time"
)

func (repositorio RepositorioPostgres) CambiarClaveConLimite(
	ctx context.Context,
	usuarioID, hashClave string,
	ahora time.Time,
	maximoHora, maximoMinuto int,
) error {
	tx, err := repositorio.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Serializa cambios por usuario para que dos solicitudes concurrentes no puedan
	// saltarse los límites al contar antes de insertar.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, usuarioID); err != nil {
		return err
	}
	var enHora, enMinuto int
	if err := tx.QueryRowContext(ctx, `
		SELECT
			count(*) FILTER (WHERE cambiado_en > $2::timestamptz - INTERVAL '1 hour'),
			count(*) FILTER (WHERE cambiado_en > $2::timestamptz - INTERVAL '1 minute')
		FROM perfil_cambios_clave
		WHERE usuario_id = $1`, usuarioID, ahora.UTC(),
	).Scan(&enHora, &enMinuto); err != nil {
		return err
	}
	if enHora >= maximoHora || enMinuto >= maximoMinuto {
		return ErrDemasiadosCambiosClave
	}

	resultado, err := tx.ExecContext(ctx, `
		UPDATE usuarios SET clave_hash = $2 WHERE id = $1`, usuarioID, hashClave)
	if err != nil {
		return err
	}
	filas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if filas == 0 {
		return ErrNoEncontrado
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO perfil_cambios_clave (usuario_id, cambiado_en)
		VALUES ($1, $2)`, usuarioID, ahora.UTC()); err != nil {
		return err
	}
	// Revoca todas las sesiones existentes: la contraseña anterior ya no es válida.
	if _, err := tx.ExecContext(ctx, `DELETE FROM sesiones WHERE usuario_id = $1`, usuarioID); err != nil {
		return err
	}
	return tx.Commit()
}
