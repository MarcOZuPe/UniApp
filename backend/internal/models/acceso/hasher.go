package acceso

type PasswordHasher interface {
	Hash(clave string) (string, error)
	Verify(clave, codificado string) bool
}
