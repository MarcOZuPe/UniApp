package acceso

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

type Fabrica struct {
	hasher        PasswordHasher
	dominiosTutor string
}

func NuevaFabrica(hasher PasswordHasher, dominiosTutor string) *Fabrica {
	return &Fabrica{hasher: hasher, dominiosTutor: dominiosTutor}
}

func (fabrica *Fabrica) PoliticaPara(rol Rol) PoliticaClave {
	if perfil, err := fabrica.perfilDe(rol); err == nil {
		return perfil.Politica()
	}
	return PoliticaEstricta{Min: 12}
}

func (fabrica *Fabrica) ValidadorPara(rol Rol) *ValidadorClave {
	return NuevoValidadorClave(fabrica.PoliticaPara(rol))
}

func (fabrica *Fabrica) CrearPublico(datos SolicitudRegistro) (*Usuario, error) {
	perfil, err := fabrica.perfilDe(datos.Rol)
	if err != nil {
		return nil, err
	}
	if !perfil.EsPublico() {
		return nil, ErrRolNoPermitido
	}
	return fabrica.construirValidado(datos, perfil, false)
}

func (fabrica *Fabrica) CrearInterno(datos SolicitudRegistro) (*Usuario, error) {
	perfil, err := fabrica.perfilDe(datos.Rol)
	if err != nil {
		return nil, err
	}
	return fabrica.construirValidado(datos, perfil, true)
}

func (fabrica *Fabrica) perfilDe(rol Rol) (PerfilRol, error) {
	return ObtenerPerfil(rol, fabrica.dominiosTutor)
}

func (fabrica *Fabrica) construirValidado(datos SolicitudRegistro, perfil PerfilRol, cuentaConfirmada bool) (*Usuario, error) {
	datos = normalizarRegistro(datos)
	if errores := validarRegistro(datos, perfil); len(errores) > 0 {
		return nil, ErrValidacion{Campos: errores}
	}
	return fabrica.construirUsuario(datos, cuentaConfirmada)
}

func normalizarRegistro(datos SolicitudRegistro) SolicitudRegistro {
	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))
	datos.NombreUso = strings.TrimSpace(datos.NombreUso)
	return datos
}

func validarRegistro(datos SolicitudRegistro, perfil PerfilRol) map[string]string {
	erroresCampos := map[string]string{}
	if !correoValido(datos.Email) {
		erroresCampos["email"] = "correo no válido"
	}
	if longitud := utf8.RuneCountInString(datos.NombreUso); longitud < 2 || longitud > 60 {
		erroresCampos["nombre_uso"] = "entre 2 y 60 caracteres"
	}
	for campo, mensaje := range perfil.ValidarExtra(datos) {
		if _, yaReportado := erroresCampos[campo]; !yaReportado {
			erroresCampos[campo] = mensaje
		}
	}
	if observaciones := NuevoValidadorClave(perfil.Politica()).Validar(datos.Clave, datos.Email); len(observaciones) > 0 {
		erroresCampos["clave"] = strings.Join(observaciones, "; ")
	}
	return erroresCampos
}

func correoValido(email string) bool {
	direccion, err := mail.ParseAddress(email)
	return err == nil && direccion.Address == email && len(email) <= 254
}

func (fabrica *Fabrica) construirUsuario(datos SolicitudRegistro, cuentaConfirmada bool) (*Usuario, error) {
	claveHash, err := fabrica.hasher.Hash(datos.Clave)
	if err != nil {
		return nil, err
	}
	return &Usuario{
		Email:           datos.Email,
		NombreUso:       datos.NombreUso,
		ClaveHash:       claveHash,
		Rol:             datos.Rol,
		Activo:          true,
		EmailConfirmado: cuentaConfirmada,
		CreadoEn:        time.Now().UTC(),
	}, nil
}
