package infra

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"hash"
	"strconv"
	"strings"
)

type derivadorPBKDF2 struct{}

func (derivadorPBKDF2) DerivarClave(clave, sal []byte, iteraciones, longitud int) []byte {
	return pbkdf2(clave, sal, iteraciones, longitud)
}

type PBKDF2Hasher struct {
	Iteraciones int
	derivador   derivadorPBKDF2
}

const longitudClave = 32

func NuevoHasher() PBKDF2Hasher { return PBKDF2Hasher{Iteraciones: 600_000} }

func (h PBKDF2Hasher) Hash(clave string) (string, error) {
	sal := make([]byte, 16)
	if _, err := rand.Read(sal); err != nil {
		return "", err
	}
	claveDerivada := h.derivador.DerivarClave([]byte(clave), sal, h.Iteraciones, longitudClave)
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", h.Iteraciones,
		base64.RawStdEncoding.EncodeToString(sal), base64.RawStdEncoding.EncodeToString(claveDerivada)), nil
}

func (h PBKDF2Hasher) Verify(clave, codificado string) bool {
	partes := strings.Split(codificado, "$")
	if len(partes) != 4 || partes[0] != "pbkdf2-sha256" {
		return false
	}
	iteraciones, err := strconv.Atoi(partes[1])
	if err != nil || iteraciones < 1 {
		return false
	}
	sal, errSal := base64.RawStdEncoding.DecodeString(partes[2])
	esperado, errHash := base64.RawStdEncoding.DecodeString(partes[3])
	if errSal != nil || errHash != nil {
		return false
	}
	calculado := h.derivador.DerivarClave([]byte(clave), sal, iteraciones, len(esperado))
	return subtle.ConstantTimeCompare(calculado, esperado) == 1
}

func pbkdf2(clave, sal []byte, iteraciones, longitud int) []byte {
	funcion := hmac.New(sha256.New, clave)
	longitudBloque := funcion.Size()
	numeroBloques := (longitud + longitudBloque - 1) / longitudBloque
	claveDerivada := make([]byte, 0, numeroBloques*longitudBloque)
	for bloque := 1; bloque <= numeroBloques; bloque++ {
		claveDerivada = append(claveDerivada, derivarBloque(funcion, sal, bloque, iteraciones)...)
	}
	return claveDerivada[:longitud]
}

func derivarBloque(funcion hash.Hash, sal []byte, indice, iteraciones int) []byte {
	funcion.Reset()
	funcion.Write(sal)
	funcion.Write([]byte{byte(indice >> 24), byte(indice >> 16), byte(indice >> 8), byte(indice)})
	acumulado := funcion.Sum(nil)
	ultimo := make([]byte, len(acumulado))
	copy(ultimo, acumulado)
	for i := 2; i <= iteraciones; i++ {
		funcion.Reset()
		funcion.Write(ultimo)
		ultimo = funcion.Sum(ultimo[:0])
		for j := range ultimo {
			acumulado[j] ^= ultimo[j]
		}
	}
	return acumulado
}
