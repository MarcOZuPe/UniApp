package acceso

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"strings"
	"time"
)

const DuracionSesion = 12 * time.Hour

type Nucleo struct {
	RepositorioUsuarios       RepositorioUsuarios
	RepositorioSesiones       RepositorioSesiones
	RepositorioTokens         RepositorioTokens
	RepositorioVinculaciones  RepositorioVinculaciones
	RepositorioAutorizaciones RepositorioAutorizaciones
	CreadorUsuarios           CreadorUsuarios
	ValidadorClaves           ValidadorClaves
	Hasher                    PasswordHasher
	Bus                       Sujeto
	Tokens                    GeneradorTokens
	Reloj                     Reloj
	HashSenuelo               string
}

type CreadorUsuarios interface {
	CrearPublico(datos SolicitudRegistro) (*Usuario, error)
	CrearInterno(datos SolicitudRegistro) (*Usuario, error)
}

type ValidadorClaves interface {
	ValidadorPara(rol Rol) *ValidadorClave
}

type Reloj interface {
	Ahora() time.Time
}

type relojSistema struct{}

func (relojSistema) Ahora() time.Time { return time.Now().UTC() }

type relojFuncion func() time.Time

func (reloj relojFuncion) Ahora() time.Time { return reloj().UTC() }

func (nucleo *Nucleo) FijarReloj(obtenerTiempo func() time.Time) {
	nucleo.Reloj = relojFuncion(obtenerTiempo)
}

func (nucleo *Nucleo) Ahora() time.Time { return nucleo.Reloj.Ahora() }

type GeneradorTokens interface {
	Nuevo() (string, []byte, error)
	Hash(token string) []byte
}

type tokensAleatorios struct{}

func (tokensAleatorios) Nuevo() (string, []byte, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, tokensAleatorios{}.Hash(token), nil
}

func (tokensAleatorios) Hash(token string) []byte {
	suma := sha256.Sum256([]byte(token))
	return suma[:]
}

func normalizarEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func eventoDeUsuario(nombre string, usuario *Usuario) Event {
	return Event{Nombre: nombre, Payload: map[string]any{"usuario_id": usuario.ID, "rol": string(usuario.Rol)}}
}
