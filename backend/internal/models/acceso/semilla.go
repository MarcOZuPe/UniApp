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
	correoTutor := "tutor.prueba@example.test"
	if dominios := dominiosDe(opciones.DominioTutor); len(dominios) > 0 {
		correoTutor = "tutor.prueba@" + dominios[0]
	}
	datos := []DatosRegistro{
		{Email: "escolar.prueba@example.test", NombreUso: "Escolar Prueba", Clave: opciones.Clave, Rol: RolEscolar},
		{Email: "apoderado.prueba@example.test", NombreUso: "Apoderado Prueba", Clave: opciones.Clave, Rol: RolApoderado},
		{Email: correoTutor, NombreUso: "Tutor Prueba", Clave: opciones.Clave, Rol: RolTutor},
		{Email: "coord.prueba@example.test", NombreUso: "Coordinación Prueba", Clave: opciones.Clave, Rol: RolCoordinador},
	}
	for _, dato := range datos {
		if _, err := servicio.CrearAdministrado(ctx, dato); err != nil && !errors.Is(err, ErrEmailEnUso) {
			return err
		}
	}
	return nil
}
