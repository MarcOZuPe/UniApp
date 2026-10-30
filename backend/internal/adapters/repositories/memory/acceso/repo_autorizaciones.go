package acceso

import (
	"context"
	"fmt"
)

func (repositorio *RepositorioMemoria) AutorizacionVigenteDeEscolar(_ context.Context, escolarID string) (*Autorizacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	var vinculacion *Vinculacion
	for _, vinculacionCandidata := range repositorio.vinculaciones {
		if vinculacionCandidata.EscolarID == escolarID && vinculacionCandidata.Estado == EstadoVinculacionAceptada {
			vinculacion = vinculacionCandidata
			break
		}
	}
	if vinculacion == nil {
		return nil, ErrNoEncontrado
	}
	autorizacion := repositorio.ultimaAutorizacionDeVinculacion(vinculacion.ID)
	if autorizacion == nil {
		return nil, ErrNoEncontrado
	}
	return repositorio.copiarAutorizacion(autorizacion), nil
}

func (repositorio *RepositorioMemoria) CambiarAutorizacion(_ context.Context, cambio CambioAutorizacion) (*Autorizacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	vinculacion, ok := repositorio.vinculaciones[cambio.VinculacionID]
	if !ok || vinculacion.Estado != EstadoVinculacionAceptada ||
		vinculacion.EscolarID != cambio.EscolarID || vinculacion.ApoderadoID != cambio.ApoderadoID {
		return nil, ErrNoEncontrado
	}
	if vigente := repositorio.ultimaAutorizacionDeVinculacion(cambio.VinculacionID); vigente != nil {
		if vigente.Estado == cambio.Estado {
			return repositorio.copiarAutorizacion(vigente), nil
		}
	}
	repositorio.secuenciaID++
	autorizacion := &Autorizacion{
		ID:            fmt.Sprintf("a-%d", repositorio.secuenciaID),
		VinculacionID: cambio.VinculacionID,
		ApoderadoID:   cambio.ApoderadoID,
		EscolarID:     cambio.EscolarID,
		Estado:        cambio.Estado,
		CreadaEn:      cambio.Momento,
	}
	repositorio.autorizaciones[autorizacion.ID] = autorizacion
	repositorio.ordenAutorizaciones = append(repositorio.ordenAutorizaciones, autorizacion.ID)
	return repositorio.copiarAutorizacion(autorizacion), nil
}

func (repositorio *RepositorioMemoria) AutorizacionesDeVinculacion(_ context.Context, vinculacionID string) ([]Autorizacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	autorizaciones := make([]Autorizacion, 0)
	for i := len(repositorio.ordenAutorizaciones) - 1; i >= 0; i-- {
		autorizacion := repositorio.autorizaciones[repositorio.ordenAutorizaciones[i]]
		if autorizacion == nil || autorizacion.VinculacionID != vinculacionID {
			continue
		}
		autorizaciones = append(autorizaciones, *repositorio.copiarAutorizacion(autorizacion))
	}
	return autorizaciones, nil
}
