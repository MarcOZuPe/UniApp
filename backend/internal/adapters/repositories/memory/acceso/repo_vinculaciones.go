package acceso

import (
	"context"
	"fmt"
	"sort"
	"time"
)

func (repositorio *RepositorioMemoria) GuardarSolicitudVinculacion(_ context.Context, vinculacion Vinculacion) (*Vinculacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	registrada := vinculacion
	repositorio.secuenciaID++
	registrada.ID = fmt.Sprintf("v-%d", repositorio.secuenciaID)
	if registrada.CreadaEn.IsZero() {
		registrada.CreadaEn = time.Now().UTC()
	}
	repositorio.vinculaciones[registrada.ID] = &registrada
	return repositorio.copiarVinculacion(&registrada), nil
}

func (repositorio *RepositorioMemoria) SolicitudPendienteDeEscolar(_ context.Context, escolarID string, momento time.Time) (*Vinculacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	var vinculacionEncontrada *Vinculacion
	for _, vinculacion := range repositorio.vinculaciones {
		if vinculacion.EscolarID != escolarID || vinculacion.Estado != EstadoVinculacionPendiente || !momento.Before(vinculacion.ExpiraEn) {
			continue
		}
		if vinculacionEncontrada == nil || vinculacion.CreadaEn.After(vinculacionEncontrada.CreadaEn) {
			vinculacionEncontrada = vinculacion
		}
	}
	if vinculacionEncontrada == nil {
		return nil, ErrNoEncontrado
	}
	return repositorio.copiarVinculacion(vinculacionEncontrada), nil
}

func (repositorio *RepositorioMemoria) VinculacionAceptadaDeEscolar(_ context.Context, escolarID string) (*Vinculacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	var vinculacionEncontrada *Vinculacion
	for _, vinculacion := range repositorio.vinculaciones {
		if vinculacion.EscolarID != escolarID || vinculacion.Estado != EstadoVinculacionAceptada {
			continue
		}
		if vinculacionEncontrada == nil || vinculacion.CreadaEn.After(vinculacionEncontrada.CreadaEn) {
			vinculacionEncontrada = vinculacion
		}
	}
	if vinculacionEncontrada == nil {
		return nil, ErrNoEncontrado
	}
	return repositorio.copiarVinculacion(vinculacionEncontrada), nil
}

func (repositorio *RepositorioMemoria) ConsumirSolicitudVinculacion(_ context.Context, consulta ConsultaVinculacionPorToken) (*Vinculacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	for _, vinculacion := range repositorio.vinculaciones {
		if vinculacion.EscolarID != consulta.EscolarID || vinculacion.Estado != EstadoVinculacionPendiente ||
			!consulta.Ahora.Before(vinculacion.ExpiraEn) || string(vinculacion.TokenHash) != string(consulta.TokenHash) {
			continue
		}
		vinculacion.Estado = EstadoVinculacionAceptada
		momento := consulta.Ahora
		vinculacion.AceptadaEn = &momento
		repositorio.secuenciaID++
		autorizacionID := fmt.Sprintf("a-%d", repositorio.secuenciaID)
		repositorio.autorizaciones[autorizacionID] = &Autorizacion{
			ID:            autorizacionID,
			VinculacionID: vinculacion.ID,
			ApoderadoID:   vinculacion.ApoderadoID,
			EscolarID:     vinculacion.EscolarID,
			Estado:        AutorizacionOtorgada,
			CreadaEn:      momento,
		}
		repositorio.ordenAutorizaciones = append(repositorio.ordenAutorizaciones, autorizacionID)
		return repositorio.copiarVinculacion(vinculacion), nil
	}
	return nil, ErrTokenInvalido
}

func (repositorio *RepositorioMemoria) VinculacionesDeApoderado(_ context.Context, apoderadoID string, momento time.Time) ([]Vinculacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	vinculaciones := make([]Vinculacion, 0)
	for _, vinculacion := range repositorio.vinculaciones {
		if vinculacion.ApoderadoID != apoderadoID {
			continue
		}
		if vinculacion.Estado == EstadoVinculacionPendiente && !momento.Before(vinculacion.ExpiraEn) {
			continue
		}
		vinculaciones = append(vinculaciones, *repositorio.copiarVinculacion(vinculacion))
	}
	sort.Slice(vinculaciones, func(indiceA, indiceB int) bool {
		if vinculaciones[indiceA].Estado != vinculaciones[indiceB].Estado {
			return vinculaciones[indiceA].Estado == EstadoVinculacionAceptada
		}
		return vinculaciones[indiceA].CreadaEn.After(vinculaciones[indiceB].CreadaEn)
	})
	return vinculaciones, nil
}

func (repositorio *RepositorioMemoria) SolicitudVinculacionPorToken(_ context.Context, consulta ConsultaVinculacionPorToken) (*Vinculacion, error) {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	for _, vinculacion := range repositorio.vinculaciones {
		if vinculacion.EscolarID == consulta.EscolarID && vinculacion.Estado == EstadoVinculacionPendiente &&
			consulta.Ahora.Before(vinculacion.ExpiraEn) && string(vinculacion.TokenHash) == string(consulta.TokenHash) {
			return repositorio.copiarVinculacion(vinculacion), nil
		}
	}
	return nil, ErrTokenInvalido
}
