package acceso

import (
	"context"
	"errors"
	"fmt"
)

const (
	EmailApoderadoSintetico = "apoderado.prueba@example.test"
	EmailEscolarSintetico   = "escolar.prueba@example.test"
)

type DatosSiembra struct {
	DominioTutor string
	Clave        string
}

type DatosSiembraVinculacion struct {
	EmailApoderado string
	EmailEscolar   string
}

func Sembrar(ctx context.Context, servicio *Servicio, datos DatosSiembra) error {
	correoTutor := "tutor.prueba@example.test"
	if dominios := DominiosDe(datos.DominioTutor); len(dominios) > 0 {
		correoTutor = "tutor.prueba@" + dominios[0]
	}
	cuentas := []SolicitudRegistro{
		{Email: EmailEscolarSintetico, NombreUso: "Escolar Prueba", Clave: datos.Clave, Rol: RolEscolar},
		{Email: EmailApoderadoSintetico, NombreUso: "Apoderado Prueba", Clave: datos.Clave, Rol: RolApoderado},
		{Email: correoTutor, NombreUso: "Tutor Prueba", Clave: datos.Clave, Rol: RolTutor},
		{Email: "coord.prueba@example.test", NombreUso: "Coordinación Prueba", Clave: datos.Clave, Rol: RolCoordinador},
	}
	for _, cuenta := range cuentas {
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
