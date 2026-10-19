package acceso

import (
	"context"
	"errors"
)

type OpcionesSiembra struct {
	DominioTutor string
	Clave        string
}

func Sembrar(ctx context.Context, servicio *Servicio, opciones OpcionesSiembra) error {
	datos := []DatosRegistro{
		{Email: "escolar.prueba@example.test", NombreUso: "Escolar Prueba", Clave: opciones.Clave, Rol: RolEscolar},
		{Email: "apoderado.prueba@example.test", NombreUso: "Apoderado Prueba", Clave: opciones.Clave, Rol: RolApoderado},
		{Email: "tutor.prueba@" + opciones.DominioTutor, NombreUso: "Tutor Prueba", Clave: opciones.Clave, Rol: RolTutor},
		{Email: "coord.prueba@example.test", NombreUso: "Coordinación Prueba", Clave: opciones.Clave, Rol: RolCoordinador},
	}
	for _, dato := range datos {
		if _, err := servicio.CrearAdministrado(ctx, dato); err != nil && !errors.Is(err, ErrEmailEnUso) {
			return err
		}
	}
	return nil
}
