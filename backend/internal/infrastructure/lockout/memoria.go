package lockout

import (
	"context"
	"sync"
	"time"
)

const (
	maxFallos      = 5
	ventanaBloqueo = 15 * time.Minute
)

type registroFallos struct {
	intentos int
	hasta    time.Time
}

// Memoria implementa LockoutManager para un único proceso. En producción debe
// sustituirse por una implementación distribuida.
type Memoria struct {
	mu     sync.Mutex
	fallos map[string]*registroFallos
	ahora  func() time.Time
}

func NuevaMemoria() *Memoria {
	return NuevaMemoriaConReloj(func() time.Time { return time.Now().UTC() })
}

func NuevaMemoriaConReloj(ahora func() time.Time) *Memoria {
	return &Memoria{fallos: map[string]*registroFallos{}, ahora: ahora}
}

func (m *Memoria) EstaBloqueado(_ context.Context, email string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	registro := m.fallos[email]
	return registro != nil && m.ahora().Before(registro.hasta), nil
}

func (m *Memoria) RegistrarFallo(_ context.Context, email string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	registro := m.fallos[email]
	if m.conteoVencido(registro) {
		registro = &registroFallos{}
		m.fallos[email] = registro
	}
	registro.intentos++
	if registro.intentos >= maxFallos {
		registro.hasta = m.ahora().Add(ventanaBloqueo)
	}
	return nil
}

func (m *Memoria) LimpiarFallos(_ context.Context, email string) error {
	m.mu.Lock()
	delete(m.fallos, email)
	m.mu.Unlock()
	return nil
}

func (m *Memoria) conteoVencido(registro *registroFallos) bool {
	if registro == nil {
		return true
	}
	return !m.ahora().Before(registro.hasta) && registro.intentos >= maxFallos
}
