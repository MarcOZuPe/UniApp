package acceso

type ErrValidacion struct {
	Campos map[string]string
}

func (errorValidacion ErrValidacion) Error() string {
	return "datos inválidos"
}
