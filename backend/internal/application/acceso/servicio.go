package acceso

type Dependencias struct {
	Repositorio Repositorio
	Fabrica     *Fabrica
	Hasher      PasswordHasher
	Bus         Sujeto
	Tokens      GeneradorTokens
	Lockout     LockoutManager
	Reloj       Reloj
}

type Servicio struct {
	*Nucleo

	Registro     *RegistroService
	Sesion       *SesionService
	Recuperacion *RecuperacionService
	Confirmacion *ConfirmacionService
	Vinculacion  *VinculacionService
	Autorizacion *AutorizacionService
}

func NuevoServicio(dependencias Dependencias) *Servicio {
	tokens := dependencias.Tokens
	if tokens == nil {
		tokens = tokensAleatorios{}
	}
	senuelo, _ := dependencias.Hasher.Hash("clave-falsa-para-igualar-tiempos")
	reloj := dependencias.Reloj
	if reloj == nil {
		reloj = relojSistema{}
	}
	nucleo := &Nucleo{
		RepositorioUsuarios:       dependencias.Repositorio,
		RepositorioSesiones:       dependencias.Repositorio,
		RepositorioTokens:         dependencias.Repositorio,
		RepositorioVinculaciones:  dependencias.Repositorio,
		RepositorioAutorizaciones: dependencias.Repositorio,
		CreadorUsuarios:           dependencias.Fabrica,
		ValidadorClaves:           dependencias.Fabrica,
		Hasher:                    dependencias.Hasher,
		Bus:                       dependencias.Bus,
		Tokens:                    tokens,
		Reloj:                     reloj,
		HashSenuelo:               senuelo,
	}
	bloqueo := dependencias.Lockout
	if bloqueo == nil {
		bloqueo = lockoutNulo{}
	}
	confirmacion := &ConfirmacionService{Nucleo: nucleo}
	if nucleo.Bus != nil {
		nucleo.Bus.Registrar(UsuarioRegistrado, NuevoConfirmacionListener(confirmacion))
	}
	return &Servicio{
		Nucleo:       nucleo,
		Registro:     &RegistroService{Nucleo: nucleo},
		Sesion:       &SesionService{Nucleo: nucleo, lockout: bloqueo},
		Recuperacion: &RecuperacionService{Nucleo: nucleo},
		Confirmacion: confirmacion,
		Vinculacion:  &VinculacionService{Nucleo: nucleo},
		Autorizacion: &AutorizacionService{Nucleo: nucleo},
	}
}
