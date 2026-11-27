package acceso

import (
	"context"
	"fmt"
	"time"
)

func (repositorio *RepositorioMemoria) CrearUsuario(_ context.Context, usuario *Usuario) (*Usuario, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	for _, existente := range repositorio.usuarios {
		if existente.Email == usuario.Email {
			return nil, ErrEmailEnUso
		}
	}
	repositorio.secuenciaID++
	registrado := *usuario
	registrado.ID = fmt.Sprintf("u-%d", repositorio.secuenciaID)
	repositorio.usuarios[registrado.ID] = &registrado
	copia := registrado
	return &copia, nil
}

func (repositorio *RepositorioMemoria) UsuarioPorEmail(_ context.Context, email string) (*Usuario, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	for _, usuario := range repositorio.usuarios {
		if usuario.Email == email {
			copia := *usuario
			return &copia, nil
		}
	}
	return nil, ErrNoEncontrado
}

func (repositorio *RepositorioMemoria) UsuarioPorID(_ context.Context, usuarioID string) (*Usuario, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	if usuario, ok := repositorio.usuarios[usuarioID]; ok {
		copia := *usuario
		return &copia, nil
	}
	return nil, ErrNoEncontrado
}

func (repositorio *RepositorioMemoria) UsuariosActivosPorRol(
	_ context.Context,
	rol Rol,
	facultadID string,
) ([]*Usuario, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	usuarios := make([]*Usuario, 0)
	for _, usuario := range repositorio.usuarios {
		if usuario.Rol != rol || !usuario.Activo || (facultadID != "" && usuario.FacultadID != facultadID) {
			continue
		}
		copia := *usuario
		usuarios = append(usuarios, &copia)
	}
	return usuarios, nil
}

func (repositorio *RepositorioMemoria) DesactivarUsuario(_ context.Context, usuarioID string) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	usuario, existe := repositorio.usuarios[usuarioID]
	if !existe {
		return ErrNoEncontrado
	}
	usuario.Activo = false
	return nil
}

func (repositorio *RepositorioMemoria) ActualizarClave(_ context.Context, usuarioID, claveHash string) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	if usuario, ok := repositorio.usuarios[usuarioID]; ok {
		usuario.ClaveHash = claveHash
		return nil
	}
	return ErrNoEncontrado
}

func (repositorio *RepositorioMemoria) MarcarCorreoConfirmado(_ context.Context, usuarioID string) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	if usuario, ok := repositorio.usuarios[usuarioID]; ok {
		usuario.EmailConfirmado = true
		return nil
	}
	return ErrNoEncontrado
}

func (repositorio *RepositorioMemoria) RegistrarIngreso(_ context.Context, usuarioID string, momento time.Time) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	if usuario, ok := repositorio.usuarios[usuarioID]; ok {
		usuario.UltimoIngreso = &momento
	}
	return nil
}

func (repositorio *RepositorioMemoria) EliminarUsuariosSinConfirmar(_ context.Context, momento time.Time) (int64, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()
	var usuariosLiberados int64
	for usuarioID, usuario := range repositorio.usuarios {
		if usuario.EmailConfirmado || repositorio.tieneEnlaceVigente(usuarioID, momento) {
			continue
		}
		for claveToken, token := range repositorio.tokensConfirmacion {
			if token.UsuarioID == usuarioID {
				delete(repositorio.tokensConfirmacion, claveToken)
			}
		}
		delete(repositorio.usuarios, usuarioID)
		usuariosLiberados++
	}
	return usuariosLiberados, nil
}
