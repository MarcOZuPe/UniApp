package acceso

import (
	"context"
	"database/sql"
	"errors"
)

func (repositorio RepositorioPostgres) GuardarSesion(ctx context.Context, sesion Sesion) error {
	_, err := repositorio.DB.ExecContext(ctx, `INSERT INTO sesiones (token_hash, usuario_id, expira_en) VALUES ($1,$2,$3)`,
		sesion.TokenHash, sesion.UsuarioID, sesion.ExpiraEn)
	return err
}

func (repositorio RepositorioPostgres) SesionPorHash(ctx context.Context, tokenHash []byte) (*Sesion, error) {
	var sesion Sesion
	err := repositorio.DB.QueryRowContext(ctx, `SELECT token_hash, usuario_id, expira_en FROM sesiones WHERE token_hash=$1`, tokenHash).
		Scan(&sesion.TokenHash, &sesion.UsuarioID, &sesion.ExpiraEn)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	return &sesion, err
}

func (repositorio RepositorioPostgres) EliminarSesion(ctx context.Context, tokenHash []byte) error {
	_, err := repositorio.DB.ExecContext(ctx, `DELETE FROM sesiones WHERE token_hash=$1`, tokenHash)
	return err
}

func (repositorio RepositorioPostgres) EliminarSesionesDeUsuario(ctx context.Context, usuarioID string) error {
	_, err := repositorio.DB.ExecContext(ctx, `DELETE FROM sesiones WHERE usuario_id=$1`, usuarioID)
	return err
}
