package acceso

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
)

const claveSesionDesarrollo = "uniapp-dev-session-secret-no-usar-en-produccion"

// TokensSesion genera y valida credenciales autocontenidas. El servidor no
// guarda sesiones: toda la información necesaria viaja firmada en el token.
type TokensSesion interface {
	Firmar(usuario *Usuario, expira time.Time) (string, error)
	Verificar(token string, ahora time.Time) (*SesionFirmada, error)
}

type SesionFirmada struct {
	UsuarioID   string
	HuellaClave string
	ExpiraEn    time.Time
}

type firmadorSesion struct {
	clave []byte
}

type cargaSesion struct {
	UsuarioID   string `json:"sub"`
	HuellaClave string `json:"cv"`
	ExpiraEn    int64  `json:"exp"`
}

func NuevoFirmadorSesion(clave string) TokensSesion {
	if strings.TrimSpace(clave) == "" {
		clave = claveSesionDesarrollo
	}
	return firmadorSesion{clave: []byte(clave)}
}

func (firmador firmadorSesion) Firmar(usuario *Usuario, expira time.Time) (string, error) {
	if usuario == nil || usuario.ID == "" || expira.IsZero() {
		return "", ErrTokenInvalido
	}
	carga, err := json.Marshal(cargaSesion{
		UsuarioID:   usuario.ID,
		HuellaClave: huellaClave(usuario.ClaveHash),
		ExpiraEn:    expira.UTC().Unix(),
	})
	if err != nil {
		return "", err
	}
	contenido := base64.RawURLEncoding.EncodeToString(carga)
	firma := firmador.firmar(contenido)
	return contenido + "." + firma, nil
}

func (firmador firmadorSesion) Verificar(token string, ahora time.Time) (*SesionFirmada, error) {
	partes := strings.Split(token, ".")
	if len(partes) != 2 || partes[0] == "" || partes[1] == "" {
		return nil, ErrTokenInvalido
	}
	esperada := firmador.firmar(partes[0])
	if !hmac.Equal([]byte(esperada), []byte(partes[1])) {
		return nil, ErrTokenInvalido
	}
	carga, err := decodificarCarga(partes[0])
	if err != nil {
		return nil, ErrTokenInvalido
	}
	expira := time.Unix(carga.ExpiraEn, 0).UTC()
	if carga.UsuarioID == "" || carga.HuellaClave == "" || !ahora.UTC().Before(expira) {
		return nil, ErrTokenInvalido
	}
	return &SesionFirmada{
		UsuarioID:   carga.UsuarioID,
		HuellaClave: carga.HuellaClave,
		ExpiraEn:    expira,
	}, nil
}

func (firmador firmadorSesion) firmar(contenido string) string {
	mac := hmac.New(sha256.New, firmador.clave)
	_, _ = mac.Write([]byte(contenido))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func decodificarCarga(contenido string) (cargaSesion, error) {
	datos, err := base64.RawURLEncoding.DecodeString(contenido)
	if err != nil {
		return cargaSesion{}, err
	}
	var carga cargaSesion
	if err := json.Unmarshal(datos, &carga); err != nil {
		return cargaSesion{}, err
	}
	return carga, nil
}

func huellaClave(hashClave string) string {
	suma := sha256.Sum256([]byte(hashClave))
	return hex.EncodeToString(suma[:])
}

func tokensSesionDePrueba() TokensSesion {
	return NuevoFirmadorSesion(claveSesionDesarrollo)
}
