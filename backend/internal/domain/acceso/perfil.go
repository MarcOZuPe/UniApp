package acceso

import "strings"

type PerfilRol interface {
	Rol() Rol
	Politica() PoliticaClave
	EsPublico() bool
	ValidarExtra(solicitud SolicitudRegistro) map[string]string
}

func ObtenerPerfil(rol Rol, dominiosTutor string) (PerfilRol, error) {
	fabrica, err := FabricaParaRol(rol)
	if err != nil {
		return nil, err
	}
	perfil := fabrica.CrearPerfil(dominiosTutor)
	if perfil == nil {
		return nil, ErrValidacion{Campos: map[string]string{"rol": "rol desconocido"}}
	}
	return perfil, nil
}

type perfilBase struct {
	rol      Rol
	politica PoliticaClave
	publico  bool
}

func (perfil perfilBase) Rol() Rol                { return perfil.rol }
func (perfil perfilBase) Politica() PoliticaClave { return perfil.politica }
func (perfil perfilBase) EsPublico() bool         { return perfil.publico }
func (perfil perfilBase) ValidarExtra(_ SolicitudRegistro) map[string]string {
	return nil
}

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
	sufijos []string
}

func nuevoPerfilTutor(dominios string) PerfilRol {
	return &perfilTutor{
		perfilBase: perfilBase{rol: RolTutor, politica: PoliticaEstricta{Min: 12}, publico: true},
		sufijos:    sufijosDeDominios(dominios),
	}
}

func (perfil *perfilTutor) ValidarExtra(solicitud SolicitudRegistro) map[string]string {
	for _, sufijo := range perfil.sufijos {
		if strings.HasSuffix(solicitud.Email, sufijo) {
			return nil
		}
	}
	if len(perfil.sufijos) == 0 {
		return map[string]string{"email": "los tutores usan su correo institucional"}
	}
	return map[string]string{"email": "los tutores usan su correo institucional (" + strings.Join(perfil.sufijos, " o ") + ")"}
}

type perfilCoordinador struct{ perfilBase }

func nuevoPerfilCoordinador() PerfilRol {
	return &perfilCoordinador{perfilBase{rol: RolCoordinador, politica: PoliticaEstricta{Min: 12}, publico: false}}
}

func DominiosDe(texto string) []string {
	dominiosLimpios := make([]string, 0, 2)
	for _, parte := range strings.Split(texto, ",") {
		if dominio := strings.ToLower(strings.TrimSpace(parte)); dominio != "" {
			dominiosLimpios = append(dominiosLimpios, dominio)
		}
	}
	return dominiosLimpios
}

func sufijosDeDominios(texto string) []string {
	dominios := DominiosDe(texto)
	sufijos := make([]string, 0, len(dominios))
	for _, dominio := range dominios {
		sufijos = append(sufijos, "@"+dominio)
	}
	return sufijos
}
