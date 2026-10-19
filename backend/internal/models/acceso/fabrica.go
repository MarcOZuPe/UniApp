package acceso

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"uniapp/internal/adapters"
)

type DatosRegistro struct {
	Email, NombreUso, Clave string
	Rol                     Rol
}

type Fabrica struct {
	hasher       adapters.PasswordHasher
	dominioTutor string
}

func NuevaFabrica(hasher adapters.PasswordHasher, dominioTutor string) *Fabrica {
	return &Fabrica{hasher: hasher, dominioTutor: dominioTutor}
}

func (f *Fabrica) PoliticaPara(rol Rol) PoliticaClave {
	if perfil, err := f.perfilDe(rol); err == nil {
		return perfil.Politica()
	}
	return PoliticaEstricta{Min: 12}
}

func (f *Fabrica) ValidadorPara(rol Rol) *ValidadorClave {
	return NuevoValidadorClave(f.PoliticaPara(rol))
}

func (f *Fabrica) CrearPublico(datos DatosRegistro) (*Usuario, error) {
	perfil, err := f.perfilDe(datos.Rol)
	if err != nil {
		return nil, err
	}
	if !perfil.EsPublico() {
		return nil, ErrRolNoPermitido
	}
	return f.construirValidado(datos, perfil)
}

func (f *Fabrica) CrearInterno(datos DatosRegistro) (*Usuario, error) {
	perfil, err := f.perfilDe(datos.Rol)
	if err != nil {
		return nil, err
	}
	return f.construirValidado(datos, perfil)
}

func (f *Fabrica) perfilDe(rol Rol) (PerfilRol, error) {
	return ObtenerPerfil(rol, f.dominioTutor)
}

func (f *Fabrica) construirValidado(datos DatosRegistro, perfil PerfilRol) (*Usuario, error) {
	datos = normalizarRegistro(datos)
	if errores := validarRegistro(datos, perfil); len(errores) > 0 {
		return nil, ErrValidacion{Campos: errores}
	}
	return f.construirUsuario(datos)
}

func normalizarRegistro(datos DatosRegistro) DatosRegistro {
	datos.Email = strings.ToLower(strings.TrimSpace(datos.Email))
	datos.NombreUso = strings.TrimSpace(datos.NombreUso)
	return datos
}

func validarRegistro(datos DatosRegistro, perfil PerfilRol) map[string]string {
	errores := map[string]string{}
	if !correoValido(datos.Email) {
		errores["email"] = "correo no válido"
	}
	if longitud := utf8.RuneCountInString(datos.NombreUso); longitud < 2 || longitud > 60 {
		errores["nombre_uso"] = "entre 2 y 60 caracteres"
	}
	for campo, mensaje := range perfil.ValidarExtra(datos) {
		if _, yaReportado := errores[campo]; !yaReportado {
			errores[campo] = mensaje
		}
	}
	if observaciones := NuevoValidadorClave(perfil.Politica()).Validar(datos.Clave, datos.Email); len(observaciones) > 0 {
		errores["clave"] = strings.Join(observaciones, "; ")
	}
	return errores
}

func correoValido(email string) bool {
	direccion, err := mail.ParseAddress(email)
	return err == nil && direccion.Address == email && len(email) <= 254
}

func (f *Fabrica) construirUsuario(datos DatosRegistro) (*Usuario, error) {
	claveHash, err := f.hasher.Hash(datos.Clave)
	if err != nil {
		return nil, err
	}
	return &Usuario{
		Email:     datos.Email,
		NombreUso: datos.NombreUso,
		ClaveHash: claveHash,
		Rol:       datos.Rol,
		Activo:    true,
		CreadoEn:  time.Now().UTC(),
	}, nil
}
