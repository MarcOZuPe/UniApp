package acceso

type PasswordHasher interface {
	Hash(clave string) (string, error)
	Verify(clavePlana, hashAlmacenado string) bool
}
