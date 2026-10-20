package acceso

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RepositorioMemoria struct {
	mu                 sync.Mutex
	secuenciaID        int
	usuarios           map[string]*Usuario
	sesiones           map[string]Sesion
	tokens             map[string]*tokenMem
	tokensConfirmacion map[string]*tokenConfirmacionMem
}

type tokenMem struct {
	TokenRecuperacion
	usado bool
}

type tokenConfirmacionMem struct {
	TokenConfirmacion
	usado bool
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		usuarios:           map[string]*Usuario{},
		sesiones:           map[string]Sesion{},
		tokens:             map[string]*tokenMem{},
		tokensConfirmacion: map[string]*tokenConfirmacionMem{},
	}
}

func (r *RepositorioMemoria) CrearUsuario(_ context.Context, usuario *Usuario) (*Usuario, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existente := range r.usuarios {
		if existente.Email == usuario.Email {
			return nil, ErrEmailEnUso
		}
	}
	r.secuenciaID++
	registrado := *usuario
	registrado.ID = fmt.Sprintf("u-%d", r.secuenciaID)
	r.usuarios[registrado.ID] = &registrado
	copia := registrado
	return &copia, nil
}

func (r *RepositorioMemoria) UsuarioPorEmail(_ context.Context, email string) (*Usuario, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, usuario := range r.usuarios {
		if usuario.Email == email {
			copia := *usuario
			return &copia, nil
		}
	}
	return nil, ErrNoEncontrado
}

func (r *RepositorioMemoria) UsuarioPorID(_ context.Context, id string) (*Usuario, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usuario, ok := r.usuarios[id]; ok {
		copia := *usuario
		return &copia, nil
	}
	return nil, ErrNoEncontrado
}

func (r *RepositorioMemoria) ActualizarClave(_ context.Context, id, claveHash string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usuario, ok := r.usuarios[id]; ok {
		usuario.ClaveHash = claveHash
		return nil
	}
	return ErrNoEncontrado
}

func (r *RepositorioMemoria) MarcarCorreoConfirmado(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usuario, ok := r.usuarios[id]; ok {
		usuario.EmailConfirmado = true
		return nil
	}
	return ErrNoEncontrado
}

func (r *RepositorioMemoria) RegistrarIngreso(_ context.Context, id string, momento time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usuario, ok := r.usuarios[id]; ok {
		usuario.UltimoIngreso = &momento
	}
	return nil
}

func (r *RepositorioMemoria) GuardarSesion(_ context.Context, sesion Sesion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sesiones[string(sesion.TokenHash)] = sesion
	return nil
}

func (r *RepositorioMemoria) SesionPorHash(_ context.Context, tokenHash []byte) (*Sesion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if sesion, ok := r.sesiones[string(tokenHash)]; ok {
		return &sesion, nil
	}
	return nil, ErrNoEncontrado
}

func (r *RepositorioMemoria) EliminarSesion(_ context.Context, tokenHash []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sesiones, string(tokenHash))
	return nil
}

func (r *RepositorioMemoria) EliminarSesionesDeUsuario(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for clave, sesion := range r.sesiones {
		if sesion.UsuarioID == id {
			delete(r.sesiones, clave)
		}
	}
	return nil
}

func (r *RepositorioMemoria) GuardarTokenRecuperacion(_ context.Context, token TokenRecuperacion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokens[string(token.TokenHash)] = &tokenMem{TokenRecuperacion: token}
	return nil
}

func (r *RepositorioMemoria) ConsumirTokenRecuperacion(_ context.Context, tokenHash []byte, ahora time.Time) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	token := r.tokens[string(tokenHash)]
	if !tokenVigente(token, ahora) {
		return "", ErrTokenInvalido
	}
	token.usado = true
	return token.UsuarioID, nil
}

func (r *RepositorioMemoria) UsuarioDeTokenRecuperacion(_ context.Context, tokenHash []byte, ahora time.Time) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	token := r.tokens[string(tokenHash)]
	if !tokenVigente(token, ahora) {
		return "", ErrTokenInvalido
	}
	return token.UsuarioID, nil
}

func (r *RepositorioMemoria) GuardarTokenConfirmacion(_ context.Context, token TokenConfirmacion) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tokensConfirmacion[string(token.TokenHash)] = &tokenConfirmacionMem{TokenConfirmacion: token}
	return nil
}

func (r *RepositorioMemoria) ConsumirTokenConfirmacion(_ context.Context, tokenHash []byte, ahora time.Time) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	token := r.tokensConfirmacion[string(tokenHash)]
	if token == nil || token.usado || !ahora.Before(token.ExpiraEn) {
		return "", ErrTokenInvalido
	}
	token.usado = true
	return token.UsuarioID, nil
}

// EliminarUsuariosSinConfirmar libera los correos de registros que nunca
// confirmaron: solo conserva las cuentas sin confirmar con algún enlace vivo.
func (r *RepositorioMemoria) EliminarUsuariosSinConfirmar(_ context.Context, ahora time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var liberados int64
	for id, usuario := range r.usuarios {
		if usuario.EmailConfirmado || r.tieneEnlaceVigente(id, ahora) {
			continue
		}
		for clave, token := range r.tokensConfirmacion {
			if token.UsuarioID == id {
				delete(r.tokensConfirmacion, clave)
			}
		}
		delete(r.usuarios, id)
		liberados++
	}
	return liberados, nil
}

func (r *RepositorioMemoria) tieneEnlaceVigente(usuarioID string, ahora time.Time) bool {
	for _, token := range r.tokensConfirmacion {
		if token.UsuarioID == usuarioID && !token.usado && ahora.Before(token.ExpiraEn) {
			return true
		}
	}
	return false
}

func tokenVigente(token *tokenMem, ahora time.Time) bool {
	return token != nil && !token.usado && ahora.Before(token.ExpiraEn)
}
