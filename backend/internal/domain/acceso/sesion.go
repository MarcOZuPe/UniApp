package acceso

import "time"

type Sesion struct {
	TokenHash []byte
	UsuarioID string
	ExpiraEn  time.Time
}
