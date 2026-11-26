package acceso

type Dependencias struct {
	Repositorio  Repositorio
	Fabrica      *Fabrica
	Hasher       PasswordHasher
	Bus          Sujeto
	Tokens       GeneradorTokens
	TokensSesion TokensSesion
	Lockout      LockoutManager
	Reloj        Reloj
}

type Servicio struct {
	*Nucleo

	Registro       *RegistroService
	Sesion         *SesionService
	Recuperacion   *RecuperacionService
	Confirmacion   *ConfirmacionService
	Vinculacion    *VinculacionService
	Autorizacion   *AutorizacionService
	Coordinadores  *CoordinadoresService
	Administracion *AdministracionService
	Perfil         *PerfilService
}

func NuevoServicio(dependencias Dependencias) *Servicio {
	tokens := dependencias.Tokens
	if tokens == nil {
		tokens = tokensAleatorios{}
	}
	tokensSesion := dependencias.TokensSesion
	if tokensSesion == nil {
		tokensSesion = tokensSesionDePrueba()
	}
	senuelo, _ := dependencias.Hasher.Hash("clave-falsa-para-igualar-tiempos")
	reloj := dependencias.Reloj
	if reloj == nil {
		reloj = relojSistema{}
	}
	nucleo := &Nucleo{
		RepositorioUsuarios:       dependencias.Repositorio,
		RepositorioFacultades:     dependencias.Repositorio,
		RepositorioPerfil:         dependencias.Repositorio,
		RepositorioLogs:           dependencias.Repositorio,
		RepositorioTokens:         dependencias.Repositorio,
		RepositorioVinculaciones:  dependencias.Repositorio,
		RepositorioAutorizaciones: dependencias.Repositorio,
		CreadorUsuarios:           dependencias.Fabrica,
		ValidadorClaves:           dependencias.Fabrica,
		Hasher:                    dependencias.Hasher,
		Bus:                       dependencias.Bus,
		Tokens:                    tokens,
		TokensSesion:              tokensSesion,
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
		Nucleo:         nucleo,
		Registro:       &RegistroService{Nucleo: nucleo},
		Sesion:         &SesionService{Nucleo: nucleo, lockout: bloqueo},
		Recuperacion:   &RecuperacionService{Nucleo: nucleo},
		Confirmacion:   confirmacion,
		Vinculacion:    &VinculacionService{Nucleo: nucleo},
		Autorizacion:   &AutorizacionService{Nucleo: nucleo},
		Coordinadores:  &CoordinadoresService{Nucleo: nucleo},
		Administracion: &AdministracionService{Nucleo: nucleo},
		Perfil:         &PerfilService{Nucleo: nucleo},
	}
}
