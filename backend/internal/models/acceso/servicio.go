package acceso

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"sync"
	"time"

	"uniapp/internal/adapters"
	"uniapp/internal/events"
)

const DuracionSesion = 12 * time.Hour

type Servicio struct {
	repositorio Repositorio
	fabrica     *Fabrica
	hasher      adapters.PasswordHasher
	bus         events.Sujeto
	ahora       func() time.Time
	hashSenuelo string

	mu     sync.Mutex
	fallos map[string]*registroFallos
}

type registroFallos struct {
	intentos int
	hasta    time.Time
}

type Dependencias struct {
	Repositorio Repositorio
	Fabrica     *Fabrica
	Hasher      adapters.PasswordHasher
	Bus         events.Sujeto
}

func NuevoServicio(dependencias Dependencias) *Servicio {

	senuelo, _ := dependencias.Hasher.Hash("clave-falsa-para-igualar-tiempos")
	return &Servicio{
		repositorio: dependencias.Repositorio,
		fabrica:     dependencias.Fabrica,
		hasher:      dependencias.Hasher,
		bus:         dependencias.Bus,
		ahora:       func() time.Time { return time.Now().UTC() },
		hashSenuelo: senuelo,
		fallos:      map[string]*registroFallos{},
	}
}

func (s *Servicio) FijarReloj(f func() time.Time) { s.ahora = f }

func normalizarEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func eventoDeUsuario(nombre string, usuario *Usuario) events.Event {
	return events.Event{Nombre: nombre, Payload: map[string]any{"usuario_id": usuario.ID, "rol": string(usuario.Rol)}}
}

func nuevoToken() (string, []byte, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, hashToken(token), nil
}

func hashToken(token string) []byte {
	suma := sha256.Sum256([]byte(token))
	return suma[:]
}
