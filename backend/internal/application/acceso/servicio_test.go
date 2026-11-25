package acceso

import (
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/acceso"
	infra "uniapp/internal/infrastructure"
	"uniapp/internal/infrastructure/lockout"
)

const clavePrueba = "Clave-Segura-2026!"

func servicioPrueba(t *testing.T, repositorio Repositorio, ahora func() time.Time) (*Servicio, *infra.Bus) {
	t.Helper()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	bus := infra.NuevoBus()
	servicio := NuevoServicio(Dependencias{
		Repositorio: repositorio,
		Fabrica:     NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         bus,
		Lockout:     lockout.NuevaMemoriaConReloj(ahora),
	})
	servicio.FijarReloj(ahora)
	return servicio, bus
}

func nuevoRepositorioPrueba() *memoria.RepositorioMemoria {
	return memoria.NuevoRepositorioMemoria()
}

func capturarToken(bus *infra.Bus, nombre, identificador string, destino *string) {
	bus.Registrar(nombre, infra.NuevoObservadorFunc(identificador, func(evento Event) {
		*destino, _ = evento.Payload["token"].(string)
	}))
}
