package acceso

import "time"

type SolicitudRegistro struct {
	Email      string
	NombreUso  string
	Clave      string
	Rol        Rol
	FacultadID string
}

type Usuario struct {
	ID              string
	Email           string
	NombreUso       string
	ClaveHash       string
	Rol             Rol
	FacultadID      string
	FacultadNombre  string
	Activo          bool
	EmailConfirmado bool
	CreadoEn        time.Time
	UltimoIngreso   *time.Time
}
