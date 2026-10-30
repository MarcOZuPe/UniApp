package acceso

import (
	"sync"
	"time"
)

type RepositorioMemoria struct {
	mu                  sync.Mutex
	secuenciaID         int
	usuarios            map[string]*Usuario
	sesiones            map[string]Sesion
	tokens              map[string]*tokenRecuperacionMemoria
	tokensConfirmacion  map[string]*tokenConfirmacionMemoria
	vinculaciones       map[string]*Vinculacion
	autorizaciones      map[string]*Autorizacion
	ordenAutorizaciones []string
}

type tokenRecuperacionMemoria struct {
	TokenRecuperacion
	usado bool
}

type tokenConfirmacionMemoria struct {
	TokenConfirmacion
	usado bool
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		usuarios:           map[string]*Usuario{},
		sesiones:           map[string]Sesion{},
		tokens:             map[string]*tokenRecuperacionMemoria{},
		tokensConfirmacion: map[string]*tokenConfirmacionMemoria{},
		vinculaciones:      map[string]*Vinculacion{},
		autorizaciones:     map[string]*Autorizacion{},
	}
}

func tokenRecuperacionVigente(token *tokenRecuperacionMemoria, ahora time.Time) bool {
	return token != nil && !token.usado && ahora.Before(token.ExpiraEn)
}

func (repositorio *RepositorioMemoria) tieneEnlaceVigente(usuarioID string, ahora time.Time) bool {
	for _, token := range repositorio.tokensConfirmacion {
		if token.UsuarioID == usuarioID && !token.usado && ahora.Before(token.ExpiraEn) {
			return true
		}
	}
	return false
}

func (repositorio *RepositorioMemoria) copiarVinculacion(vinculacion *Vinculacion) *Vinculacion {
	copia := *vinculacion
	if usuario, ok := repositorio.usuarios[copia.ApoderadoID]; ok {
		copia.ApoderadoNombre = usuario.NombreUso
		copia.ApoderadoEmail = usuario.Email
	}
	if usuario, ok := repositorio.usuarios[copia.EscolarID]; ok {
		copia.EscolarNombre = usuario.NombreUso
		copia.EscolarEmail = usuario.Email
	}
	if autorizacion := repositorio.ultimaAutorizacionDeVinculacion(copia.ID); autorizacion != nil {
		copia.Autorizacion = repositorio.copiarAutorizacion(autorizacion)
	}
	return &copia
}

func (repositorio *RepositorioMemoria) ultimaAutorizacionDeVinculacion(vinculacionID string) *Autorizacion {
	var encontrada *Autorizacion
	for _, autorizacion := range repositorio.autorizaciones {
		if autorizacion.VinculacionID != vinculacionID {
			continue
		}
		if encontrada == nil || autorizacion.CreadaEn.After(encontrada.CreadaEn) ||
			(autorizacion.CreadaEn.Equal(encontrada.CreadaEn) && autorizacion.ID > encontrada.ID) {
			encontrada = autorizacion
		}
	}
	return encontrada
}

func (repositorio *RepositorioMemoria) copiarAutorizacion(autorizacion *Autorizacion) *Autorizacion {
	copia := *autorizacion
	if usuario, ok := repositorio.usuarios[copia.ApoderadoID]; ok {
		copia.ApoderadoNombre = usuario.NombreUso
		copia.ApoderadoEmail = usuario.Email
	}
	if usuario, ok := repositorio.usuarios[copia.EscolarID]; ok {
		copia.EscolarNombre = usuario.NombreUso
		copia.EscolarEmail = usuario.Email
	}
	return &copia
}
