package acceso

import (
	"context"
	"time"

	"uniapp/internal/events"
)

const (
	maxFallos      = 5
	ventanaBloqueo = 15 * time.Minute
)

func (s *Servicio) Ingresar(ctx context.Context, email, clave string) (string, *Usuario, error) {
	email = normalizarEmail(email)
	if s.estaBloqueado(email) {
		return "", nil, ErrDemasiadosIntentos
	}
	usuario, err := s.autenticar(ctx, email, clave)
	if err != nil {
		return "", nil, err
	}
	token, err := s.abrirSesion(ctx, usuario)
	if err != nil {
		return "", nil, err
	}
	return token, usuario, nil
}

func (s *Servicio) autenticar(ctx context.Context, email, clave string) (*Usuario, error) {
	usuario, err := s.repositorio.UsuarioPorEmail(ctx, email)
	hash := s.hashSenuelo
	if err == nil {
		hash = usuario.ClaveHash
	}
	credencialesValidas := s.hasher.Verify(clave, hash)
	if err != nil || !credencialesValidas || !usuario.Activo || !usuario.EmailConfirmado {
		s.registrarFallo(email)
		s.bus.Notificar(events.Event{Nombre: events.IngresoFallido, Payload: map[string]any{"usuario_id": "-", "rol": "-"}})
		return nil, ErrCredenciales
	}
	s.limpiarFallos(email)
	return usuario, nil
}

func (s *Servicio) abrirSesion(ctx context.Context, usuario *Usuario) (string, error) {
	token, tokenHash, err := nuevoToken()
	if err != nil {
		return "", err
	}
	momento := s.ahora()
	sesion := Sesion{TokenHash: tokenHash, UsuarioID: usuario.ID, ExpiraEn: momento.Add(DuracionSesion)}
	if err := s.repositorio.GuardarSesion(ctx, sesion); err != nil {
		return "", err
	}
	_ = s.repositorio.RegistrarIngreso(ctx, usuario.ID, momento)
	s.bus.Notificar(eventoDeUsuario(events.IngresoExitoso, usuario))
	return token, nil
}

func (s *Servicio) Salir(ctx context.Context, token string) error {
	return s.repositorio.EliminarSesion(ctx, hashToken(token))
}

func (s *Servicio) UsuarioDeSesion(ctx context.Context, token string) (*Usuario, error) {
	if token == "" {
		return nil, ErrNoEncontrado
	}
	tokenHash := hashToken(token)
	sesionGuardada, err := s.repositorio.SesionPorHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if !s.ahora().Before(sesionGuardada.ExpiraEn) {
		_ = s.repositorio.EliminarSesion(ctx, tokenHash)
		return nil, ErrNoEncontrado
	}
	usuario, err := s.repositorio.UsuarioPorID(ctx, sesionGuardada.UsuarioID)
	if err != nil || !usuario.Activo {
		return nil, ErrNoEncontrado
	}
	return usuario, nil
}

func (s *Servicio) estaBloqueado(email string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	registro := s.fallos[email]
	return registro != nil && s.ahora().Before(registro.hasta)
}

func (s *Servicio) registrarFallo(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	registro := s.fallos[email]
	if s.conteoVencido(registro) {
		registro = &registroFallos{}
		s.fallos[email] = registro
	}
	registro.intentos++
	if registro.intentos >= maxFallos {
		registro.hasta = s.ahora().Add(ventanaBloqueo)
	}
}

func (s *Servicio) conteoVencido(registro *registroFallos) bool {
	if registro == nil {
		return true
	}
	return !s.ahora().Before(registro.hasta) && registro.intentos >= maxFallos
}

func (s *Servicio) limpiarFallos(email string) {
	s.mu.Lock()
	delete(s.fallos, email)
	s.mu.Unlock()
}
