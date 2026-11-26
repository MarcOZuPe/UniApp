package acceso

import (
	"context"
	"time"
)

func (repositorio *RepositorioMemoria) CambiarClaveConLimite(
	_ context.Context,
	usuarioID, hashClave string,
	ahora time.Time,
	maximoHora, maximoMinuto int,
) error {
	repositorio.mu.Lock()
	defer repositorio.mu.Unlock()

	usuario, existe := repositorio.usuarios[usuarioID]
	if !existe {
		return ErrNoEncontrado
	}
	enHora, enMinuto := 0, 0
	for _, cambio := range repositorio.cambiosClave[usuarioID] {
		if cambio.After(ahora.Add(-time.Hour)) {
			enHora++
		}
		if cambio.After(ahora.Add(-time.Minute)) {
			enMinuto++
		}
	}
	if enHora >= maximoHora || enMinuto >= maximoMinuto {
		return ErrDemasiadosCambiosClave
	}

	usuario.ClaveHash = hashClave
	repositorio.cambiosClave[usuarioID] = append(repositorio.cambiosClave[usuarioID], ahora)
	for claveSesion, sesion := range repositorio.sesiones {
		if sesion.UsuarioID == usuarioID {
			delete(repositorio.sesiones, claveSesion)
		}
	}
	return nil
}
