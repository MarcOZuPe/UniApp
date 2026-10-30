package acceso

import (
	"context"
	"time"
)

func (repositorio *RepositorioMemoria) GuardarTokenRecuperacion(_ context.Context, tokenRecuperacion TokenRecuperacion) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	repositorio.tokens[string(tokenRecuperacion.TokenHash)] = &tokenRecuperacionMemoria{TokenRecuperacion: tokenRecuperacion}
	return nil
}

func (repositorio *RepositorioMemoria) ConsumirTokenRecuperacion(_ context.Context, tokenHash []byte, momento time.Time) (string, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	tokenRecuperacion := repositorio.tokens[string(tokenHash)]
	if !tokenRecuperacionVigente(tokenRecuperacion, momento) {
		return "", ErrTokenInvalido
	}
	tokenRecuperacion.usado = true
	return tokenRecuperacion.UsuarioID, nil
}

func (repositorio *RepositorioMemoria) UsuarioDeTokenRecuperacion(_ context.Context, tokenHash []byte, momento time.Time) (string, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	tokenRecuperacion := repositorio.tokens[string(tokenHash)]
	if !tokenRecuperacionVigente(tokenRecuperacion, momento) {
		return "", ErrTokenInvalido
	}
	return tokenRecuperacion.UsuarioID, nil
}

func (repositorio *RepositorioMemoria) GuardarTokenConfirmacion(_ context.Context, tokenConfirmacion TokenConfirmacion) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	repositorio.tokensConfirmacion[string(tokenConfirmacion.TokenHash)] = &tokenConfirmacionMemoria{TokenConfirmacion: tokenConfirmacion}
	return nil
}

func (repositorio *RepositorioMemoria) ConsumirTokenConfirmacion(_ context.Context, tokenHash []byte, momento time.Time) (string, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	tokenConfirmacion := repositorio.tokensConfirmacion[string(tokenHash)]
	if tokenConfirmacion == nil || tokenConfirmacion.usado || !momento.Before(tokenConfirmacion.ExpiraEn) {
		return "", ErrTokenInvalido
	}
	tokenConfirmacion.usado = true
	return tokenConfirmacion.UsuarioID, nil
}
