package acceso

import "time"

type TokenRecuperacion struct {
	TokenHash []byte
	UsuarioID string
	ExpiraEn  time.Time
}

type TokenConfirmacion struct {
	TokenHash []byte
	UsuarioID string
	ExpiraEn  time.Time
}
