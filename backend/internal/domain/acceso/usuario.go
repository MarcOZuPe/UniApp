package acceso

import "time"

type SolicitudRegistro struct {
	Email     string
	NombreUso string
	Clave     string
	Rol       Rol
}

type Usuario struct {
	ID              string
	Email           string
	NombreUso       string
	ClaveHash       string
	Rol             Rol
	Activo          bool
	EmailConfirmado bool
	CreadoEn        time.Time
	UltimoIngreso   *time.Time
}
