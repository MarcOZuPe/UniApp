package acceso

type escaneadorFila interface {
	Scan(destinos ...any) error
}
