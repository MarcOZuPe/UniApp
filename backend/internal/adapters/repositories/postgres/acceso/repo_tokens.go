package acceso

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (repositorio RepositorioPostgres) GuardarTokenRecuperacion(ctx context.Context, tokenRecuperacion TokenRecuperacion) error {
	_, err := repositorio.DB.ExecContext(ctx, `INSERT INTO tokens_recuperacion (token_hash, usuario_id, expira_en) VALUES ($1,$2,$3)`,
		tokenRecuperacion.TokenHash, tokenRecuperacion.UsuarioID, tokenRecuperacion.ExpiraEn)
	return err
}

func (repositorio RepositorioPostgres) ConsumirTokenRecuperacion(ctx context.Context, tokenHash []byte, momento time.Time) (string, error) {
	var usuarioID string
	err := repositorio.DB.QueryRowContext(ctx,
		`UPDATE tokens_recuperacion SET usado_en=$2 WHERE token_hash=$1 AND usado_en IS NULL AND expira_en > $2 RETURNING usuario_id`,
		tokenHash, momento).Scan(&usuarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTokenInvalido
	}
	return usuarioID, err
}

func (repositorio RepositorioPostgres) UsuarioDeTokenRecuperacion(ctx context.Context, tokenHash []byte, momento time.Time) (string, error) {
	var usuarioID string
	err := repositorio.DB.QueryRowContext(ctx,
		`SELECT usuario_id FROM tokens_recuperacion WHERE token_hash=$1 AND usado_en IS NULL AND expira_en > $2`, tokenHash, momento).Scan(&usuarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTokenInvalido
	}
	return usuarioID, err
}

func (repositorio RepositorioPostgres) GuardarTokenConfirmacion(ctx context.Context, tokenConfirmacion TokenConfirmacion) error {
	_, err := repositorio.DB.ExecContext(ctx, `INSERT INTO tokens_confirmacion (token_hash, usuario_id, expira_en) VALUES ($1,$2,$3)`,
		tokenConfirmacion.TokenHash, tokenConfirmacion.UsuarioID, tokenConfirmacion.ExpiraEn)
	return err
}

func (repositorio RepositorioPostgres) ConsumirTokenConfirmacion(ctx context.Context, tokenHash []byte, momento time.Time) (string, error) {
	var usuarioID string
	err := repositorio.DB.QueryRowContext(ctx,
		`UPDATE tokens_confirmacion SET usado_en=$2 WHERE token_hash=$1 AND usado_en IS NULL AND expira_en > $2 RETURNING usuario_id`,
		tokenHash, momento).Scan(&usuarioID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTokenInvalido
	}
	return usuarioID, err
}
