package acceso

type Rol string

const (
	RolEscolar            Rol = "escolar"
	RolApoderado          Rol = "apoderado"
	RolTutor              Rol = "tutor"
	RolCoordinador        Rol = "coordinador"
	RolCoordinadorGeneral Rol = "coordinador_general"
)

func (rol Rol) EsCoordinacion() bool {
	return rol == RolCoordinador || rol == RolCoordinadorGeneral
}

func (rol Rol) RequiereFacultad() bool {
	return rol == RolTutor || rol == RolCoordinador || rol == RolCoordinadorGeneral
}
