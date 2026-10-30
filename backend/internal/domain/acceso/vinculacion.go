package acceso

import (
	"errors"
	"time"
)

type EstadoVinculacion string

const (
	EstadoVinculacionPendiente EstadoVinculacion = "pendiente"
	EstadoVinculacionAceptada  EstadoVinculacion = "aceptada"
)

type ConsultaVinculacionPorToken struct {
	TokenHash []byte
	EscolarID string
	Ahora     time.Time
}

type Vinculacion struct {
	ID          string
	ApoderadoID string
	EscolarID   string
	Estado      EstadoVinculacion

	TokenHash  []byte
	ExpiraEn   time.Time
	CreadaEn   time.Time
	AceptadaEn *time.Time

	ApoderadoNombre string
	ApoderadoEmail  string
	EscolarNombre   string
	EscolarEmail    string

	Autorizacion *Autorizacion
}

var (
	ErrNoEsApoderado       = errors.New("solo un apoderado puede realizar esta acción")
	ErrNoEsEscolar         = errors.New("solo un escolar puede realizar esta acción")
	ErrEscolarNoEncontrado = errors.New("no existe un escolar con ese correo")
	ErrEscolarNoValido     = errors.New("el usuario indicado no es un escolar activo y confirmado")
	ErrYaVinculado         = errors.New("el escolar ya tiene un apoderado vinculado")
	ErrSolicitudPendiente  = errors.New("el escolar ya tiene una solicitud de vinculación pendiente")
	ErrSinApoderado        = errors.New("el escolar no tiene un apoderado vinculado")
	ErrVinculacionAjena    = errors.New("la vinculación no pertenece al apoderado")
)
