package acceso

import (
	"context"
	"errors"
	"fmt"
)

const (
	EmailApoderadoSintetico                   = "apoderado.prueba@example.test"
	EmailEscolarSintetico                     = "escolar.prueba@example.test"
	EmailEscolarSinteticoDos                  = "escolar2.prueba@example.test"
	EmailCoordinadorSintetico                 = "coord.prueba@example.test"
	EmailCoordinadorGeneralSintetico          = "coord.general.prueba@example.test"
	EmailCoordinadorCiencias                  = "coord.ciencias@example.test"
	EmailCoordinadorIndustrialSistemas        = "coord.sistemas@example.test"
	EmailCoordinadorMecanica                  = "coord.mecanica@example.test"
	EmailCoordinadorArquitectura              = "coord.arquitectura@example.test"
	EmailCoordinadorGeneralCiencias           = "general.ciencias@example.test"
	EmailCoordinadorGeneralIndustrialSistemas = "general.sistemas@example.test"
	EmailCoordinadorGeneralMecanica           = "general.mecanica@example.test"
	EmailCoordinadorGeneralArquitectura       = "general.arquitectura@example.test"
)

type DatosSiembra struct {
	DominioTutor string
	Clave        string
}

type DatosSiembraVinculacion struct {
	EmailApoderado string
	EmailEscolar   string
}

func CorreosTutoresSinteticos(dominioTutor string) []string {
	dominio := primerDominioDe(dominioTutor)
	return []string{
		"tutor.prueba@" + dominio,
		"tutor2.prueba@" + dominio,
		"tutor.ciencias@" + dominio,
		"tutor.sistemas@" + dominio,
		"tutor.mecanica@" + dominio,
		"tutor.arquitectura@" + dominio,
	}
}

func primerDominioDe(texto string) string {
	dominios := DominiosDe(texto)
	if len(dominios) > 0 {
		return dominios[0]
	}
	return "example.test"
}

func Sembrar(ctx context.Context, servicio *Servicio, datos DatosSiembra) error {
	correoTutor := "tutor.prueba@example.test"
	correoTutorDos := "tutor2.prueba@example.test"
	if dominios := DominiosDe(datos.DominioTutor); len(dominios) > 0 {
		correoTutor = "tutor.prueba@" + dominios[0]
		correoTutorDos = "tutor2.prueba@" + dominios[0]
	}
	dominio := primerDominioDe(datos.DominioTutor)
	cuentas := []SolicitudRegistro{
		{Email: EmailEscolarSintetico, NombreUso: "Escolar Prueba", Clave: datos.Clave, Rol: RolEscolar},
		{Email: EmailEscolarSinteticoDos, NombreUso: "Escolar Prueba Dos", Clave: datos.Clave, Rol: RolEscolar},
		{Email: EmailApoderadoSintetico, NombreUso: "Apoderado Prueba", Clave: datos.Clave, Rol: RolApoderado},
		{Email: correoTutor, NombreUso: "Tutor Prueba", Clave: datos.Clave, Rol: RolTutor},
		{Email: correoTutorDos, NombreUso: "Tutor Prueba Dos", Clave: datos.Clave, Rol: RolTutor},
		{Email: EmailCoordinadorSintetico, NombreUso: "Coordinación Prueba", Clave: datos.Clave, Rol: RolCoordinador},
		{Email: EmailCoordinadorGeneralSintetico, NombreUso: "Coordinación General Prueba", Clave: datos.Clave, Rol: RolCoordinadorGeneral},

		{Email: EmailCoordinadorGeneralCiencias, NombreUso: "Coordinación General Ciencias", Clave: datos.Clave, Rol: RolCoordinadorGeneral, FacultadID: "1"},
		{Email: EmailCoordinadorCiencias, NombreUso: "Coordinación Ciencias", Clave: datos.Clave, Rol: RolCoordinador, FacultadID: "1"},
		{Email: "tutor.ciencias@" + dominio, NombreUso: "Tutor Ciencias", Clave: datos.Clave, Rol: RolTutor, FacultadID: "1"},

		{Email: EmailCoordinadorGeneralIndustrialSistemas, NombreUso: "Coordinación General Industrial y Sistemas", Clave: datos.Clave, Rol: RolCoordinadorGeneral, FacultadID: "2"},
		{Email: EmailCoordinadorIndustrialSistemas, NombreUso: "Coordinación Industrial y Sistemas", Clave: datos.Clave, Rol: RolCoordinador, FacultadID: "2"},
		{Email: "tutor.sistemas@" + dominio, NombreUso: "Tutor Industrial y Sistemas", Clave: datos.Clave, Rol: RolTutor, FacultadID: "2"},

		{Email: EmailCoordinadorGeneralMecanica, NombreUso: "Coordinación General Mecánica", Clave: datos.Clave, Rol: RolCoordinadorGeneral, FacultadID: "3"},
		{Email: EmailCoordinadorMecanica, NombreUso: "Coordinación Mecánica", Clave: datos.Clave, Rol: RolCoordinador, FacultadID: "3"},
		{Email: "tutor.mecanica@" + dominio, NombreUso: "Tutor Mecánica", Clave: datos.Clave, Rol: RolTutor, FacultadID: "3"},

		{Email: EmailCoordinadorGeneralArquitectura, NombreUso: "Coordinación General Arquitectura", Clave: datos.Clave, Rol: RolCoordinadorGeneral, FacultadID: "4"},
		{Email: EmailCoordinadorArquitectura, NombreUso: "Coordinación Arquitectura", Clave: datos.Clave, Rol: RolCoordinador, FacultadID: "4"},
		{Email: "tutor.arquitectura@" + dominio, NombreUso: "Tutor Arquitectura", Clave: datos.Clave, Rol: RolTutor, FacultadID: "4"},
	}
	for _, cuenta := range cuentas {
		if _, err := servicio.RepositorioUsuarios.UsuarioPorEmail(ctx, normalizarEmail(cuenta.Email)); err == nil {
			continue
		} else if !errors.Is(err, ErrNoEncontrado) {
			return err
		}

		if _, err := servicio.Registro.CrearAdministrado(ctx, cuenta); err != nil && !errors.Is(err, ErrEmailEnUso) {
			return err
		}
	}
	return nil
}

func SembrarVinculacion(ctx context.Context, servicio *Servicio, datos DatosSiembraVinculacion) error {
	apoderado, err := servicio.RepositorioUsuarios.UsuarioPorEmail(ctx, normalizarEmail(datos.EmailApoderado))
	if err != nil {
		return fmt.Errorf("buscar apoderado sintético: %w", err)
	}
	escolar, err := servicio.RepositorioUsuarios.UsuarioPorEmail(ctx, normalizarEmail(datos.EmailEscolar))
	if err != nil {
		return fmt.Errorf("buscar escolar sintético: %w", err)
	}

	if _, err := servicio.RepositorioVinculaciones.VinculacionAceptadaDeEscolar(ctx, escolar.ID); err == nil {
		return nil
	} else if !errors.Is(err, ErrNoEncontrado) {
		return err
	}

	token, tokenHash, err := servicio.Tokens.Nuevo()
	if err != nil {
		return err
	}
	momento := servicio.Ahora()
	if _, err := servicio.RepositorioVinculaciones.GuardarSolicitudVinculacion(ctx, Vinculacion{
		ApoderadoID: apoderado.ID,
		EscolarID:   escolar.ID,
		Estado:      EstadoVinculacionPendiente,
		TokenHash:   tokenHash,
		ExpiraEn:    momento.Add(duracionVinculacion),
		CreadaEn:    momento,
	}); err != nil {
		return fmt.Errorf("crear vínculo sintético: %w", err)
	}
	if _, err := servicio.RepositorioVinculaciones.ConsumirSolicitudVinculacion(ctx, ConsultaVinculacionPorToken{
		TokenHash: servicio.Tokens.Hash(token),
		EscolarID: escolar.ID,
		Ahora:     momento,
	}); err != nil {
		return fmt.Errorf("aceptar vínculo sintético: %w", err)
	}
	return nil
}
