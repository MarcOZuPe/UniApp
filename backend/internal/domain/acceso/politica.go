package acceso

import (
	"strconv"
	"strings"
	"unicode"
)

type PoliticaClave interface {
	Validar(clave, email string) []string
}

type ValidadorClave struct{ politica PoliticaClave }

func NuevoValidadorClave(politica PoliticaClave) *ValidadorClave {
	return &ValidadorClave{politica: politica}
}

func (validador *ValidadorClave) FijarPolitica(politica PoliticaClave) {
	validador.politica = politica
}

func (validador *ValidadorClave) Validar(clave, email string) []string {
	return validador.politica.Validar(clave, email)
}

type PoliticaBasica struct{ Min int }

func (politica PoliticaBasica) Validar(clave, email string) []string {
	var observaciones []string
	if len([]rune(clave)) < politica.Min {
		observaciones = append(observaciones, "debe tener al menos "+strconv.Itoa(politica.Min)+" caracteres")
	}
	if nombreLocal := nombreLocalDeCorreo(email); nombreLocal != "" && strings.EqualFold(clave, nombreLocal) {
		observaciones = append(observaciones, "no puede ser igual a tu correo")
	}
	return observaciones
}

type PoliticaEstricta struct{ Min int }

func (politica PoliticaEstricta) Validar(clave, email string) []string {
	observaciones := PoliticaBasica(politica).Validar(clave, email)
	if !combinaTodasLasClases(clave) {
		observaciones = append(observaciones, "debe combinar mayúscula, minúscula, número y símbolo")
	}
	return observaciones
}

func nombreLocalDeCorreo(email string) string {
	return strings.ToLower(strings.SplitN(email, "@", 2)[0])
}

func combinaTodasLasClases(clave string) bool {
	var tieneMayuscula, tieneMinuscula, tieneDigito, tieneSimbolo bool
	for _, caracter := range clave {
		switch {
		case unicode.IsUpper(caracter):
			tieneMayuscula = true
		case unicode.IsLower(caracter):
			tieneMinuscula = true
		case unicode.IsDigit(caracter):
			tieneDigito = true
		default:
			tieneSimbolo = true
		}
	}
	return tieneMayuscula && tieneMinuscula && tieneDigito && tieneSimbolo
}
