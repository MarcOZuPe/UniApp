package acceso

import (
	"strings"
)

type PerfilRol interface {
	Rol() Rol
	Politica() PoliticaClave
	EsPublico() bool

	ValidarExtra(d DatosRegistro) map[string]string
}

type perfilBase struct {
	rol      Rol
	politica PoliticaClave
	publico  bool
}

func (p perfilBase) Rol() Rol                                     { return p.rol }
func (p perfilBase) Politica() PoliticaClave                      { return p.politica }
func (p perfilBase) EsPublico() bool                              { return p.publico }
func (p perfilBase) ValidarExtra(DatosRegistro) map[string]string { return nil }

type perfilEscolar struct{ perfilBase }

func nuevoPerfilEscolar() PerfilRol {
	return &perfilEscolar{perfilBase{rol: RolEscolar, politica: PoliticaBasica{Min: 10}, publico: true}}
}

type perfilApoderado struct{ perfilBase }

func nuevoPerfilApoderado() PerfilRol {
	return &perfilApoderado{perfilBase{rol: RolApoderado, politica: PoliticaBasica{Min: 10}, publico: true}}
}

type perfilTutor struct {
	perfilBase
	sufijo string
}

func nuevoPerfilTutor(dominio string) PerfilRol {
	return &perfilTutor{
		perfilBase: perfilBase{rol: RolTutor, politica: PoliticaEstricta{Min: 12}, publico: true},
		sufijo:     "@" + strings.ToLower(dominio),
	}
}

func (p *perfilTutor) ValidarExtra(d DatosRegistro) map[string]string {
	if !strings.HasSuffix(d.Email, p.sufijo) {
		return map[string]string{"email": "los tutores usan su correo institucional (" + p.sufijo + ")"}
	}
	return nil
}

type perfilCoordinador struct{ perfilBase }

func nuevoPerfilCoordinador() PerfilRol {
	return &perfilCoordinador{perfilBase{rol: RolCoordinador, politica: PoliticaEstricta{Min: 12}, publico: false}}
}

func ObtenerPerfil(rol Rol, dominioTutor string) (PerfilRol, error) {
	switch rol {
	case RolEscolar:
		return nuevoPerfilEscolar(), nil
	case RolApoderado:
		return nuevoPerfilApoderado(), nil
	case RolTutor:
		return nuevoPerfilTutor(dominioTutor), nil
	case RolCoordinador:
		return nuevoPerfilCoordinador(), nil
	}
	return nil, ErrValidacion{Campos: map[string]string{"rol": "rol desconocido"}}
}
