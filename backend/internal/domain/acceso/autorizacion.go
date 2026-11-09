package acceso

import "time"

type EstadoAutorizacion string

const (
	AutorizacionOtorgada EstadoAutorizacion = "autorizado"
	AutorizacionRevocada EstadoAutorizacion = "no_autorizado"
)

type Autorizacion struct {
	ID              string
	VinculacionID   string
	ApoderadoID     string
	EscolarID       string
	Estado          EstadoAutorizacion
	CreadaEn        time.Time
	ApoderadoNombre string
	ApoderadoEmail  string
	EscolarNombre   string
	EscolarEmail    string
}

func (autorizacion Autorizacion) Vigente() bool {
	return autorizacion.Estado == AutorizacionOtorgada
}

type CambioAutorizacion struct {
	VinculacionID string
	EscolarID     string
	ApoderadoID   string
	Estado        EstadoAutorizacion
	Momento       time.Time
}
