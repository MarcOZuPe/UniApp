package acceso

import "context"

func (repositorio *RepositorioMemoria) GuardarSesion(_ context.Context, sesion Sesion) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	repositorio.sesiones[string(sesion.TokenHash)] = sesion
	return nil
}

func (repositorio *RepositorioMemoria) SesionPorHash(_ context.Context, tokenHash []byte) (*Sesion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	if sesion, ok := repositorio.sesiones[string(tokenHash)]; ok {
		return &sesion, nil
	}
	return nil, ErrNoEncontrado
}

func (repositorio *RepositorioMemoria) EliminarSesion(_ context.Context, tokenHash []byte) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	delete(repositorio.sesiones, string(tokenHash))
	return nil
}

func (repositorio *RepositorioMemoria) EliminarSesionesDeUsuario(_ context.Context, usuarioID string) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	for claveSesion, sesion := range repositorio.sesiones {
		if sesion.UsuarioID == usuarioID {
			delete(repositorio.sesiones, claveSesion)
		}
	}
	return nil
}
