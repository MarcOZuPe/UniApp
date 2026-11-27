package panel

import modeloAcceso "uniapp/internal/application/acceso"

// Solicitante representa al usuario que consulta el panel. Mantiene unidas las
// dos condiciones de acceso (rol y facultad) para que las firmas del servicio
// no crezcan y la autorización se valide en un único lugar.
type Solicitante struct {
	Rol        modeloAcceso.Rol
	FacultadID string
}

func (solicitante Solicitante) Autorizado() bool {
	return solicitante.Rol.EsCoordinacion() && solicitante.FacultadID != ""
}
