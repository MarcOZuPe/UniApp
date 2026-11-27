package reserva

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	memoriaReserva "uniapp/internal/adapters/repositories/memory/reserva"
)

// autorizacionQueSeRevoca aprueba la comprobacion inicial y revoca la
// autorizacion inmediatamente despues, simulando una revocacion concurrente
// entre la comprobacion (F-09) y la transaccion de reserva.
type autorizacionQueSeRevoca struct {
	mu        sync.Mutex
	revocada  bool
	consultas int
}

func (a *autorizacionQueSeRevoca) AutorizacionVigente(context.Context, string) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.consultas++
	if a.revocada {
		return false, nil
	}
	a.revocada = true
	return true, nil
}

func datosReservaValidos(slotID string) DatosReserva {
	return DatosReserva{
		SlotID:   slotID,
		Tema:     "Resolver ecuaciones de primer grado",
		Objetivo: "Practicar ejercicios y aclarar dudas del tema.",
		Grado:    "5to de secundaria",
	}
}

// F-09: la comprobacion de autorizacion y la creacion de la reserva no deben
// estar separadas por una ventana en la que la autorizacion se revoque.
func TestF09ReservaNoDebeCrearseSiLaAutorizacionSeRevocaAntesDeConfirmar(t *testing.T) {
	base := memoriaReserva.NuevoRepositorioMemoria()
	autorizacion := &autorizacionQueSeRevoca{}
	oferta := NuevaOfertaSintetica(time.Now().UTC())
	servicio := NuevoServicio(Dependencias{
		Repositorio:  base,
		Oferta:       oferta,
		Autorizacion: autorizacion,
	})

	_, _, err := servicio.Reservar(context.Background(), "e-f09", datosReservaValidos("slot-i04-01"))
	if !errors.Is(err, ErrSinAutorizacion) {
		t.Fatalf("F-09: se creo la reserva aunque la autorizacion se revoco antes de confirmar (err=%v, consultas=%d)", err, autorizacion.consultas)
	}
}

// repoConBarreraAdjuntos sincroniza dos ContarAdjuntos para que ambas subidas
// lean el mismo conteo antes de insertar (F-10).
type repoConBarreraAdjuntos struct {
	*memoriaReserva.RepositorioMemoria
	mu       sync.Mutex
	llegadas int
	liberar  chan struct{}
}

func (r *repoConBarreraAdjuntos) ContarAdjuntos(ctx context.Context, reservaID string) (int, error) {
	total, err := r.RepositorioMemoria.ContarAdjuntos(ctx, reservaID)
	r.mu.Lock()
	r.llegadas++
	if r.llegadas == 2 {
		close(r.liberar)
	}
	r.mu.Unlock()
	select {
	case <-r.liberar:
	case <-time.After(3 * time.Second):
	}
	return total, err
}

type almacenEnMemoria struct {
	mu    sync.Mutex
	datos map[string][]byte
}

func nuevoAlmacenEnMemoria() *almacenEnMemoria {
	return &almacenEnMemoria{datos: map[string][]byte{}}
}

func (a *almacenEnMemoria) Guardar(_ context.Context, clave string, contenido []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.datos[clave] = append([]byte(nil), contenido...)
	return nil
}

func (a *almacenEnMemoria) Leer(_ context.Context, clave string) ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	contenido, existe := a.datos[clave]
	if !existe {
		return nil, ErrAlmacen
	}
	return append([]byte(nil), contenido...), nil
}

func (a *almacenEnMemoria) Eliminar(_ context.Context, clave string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.datos, clave)
	return nil
}

// F-10: dos subidas concurrentes no deben superar el maximo de adjuntos de la
// reserva. El conteo previo y la insercion deben ser atomicos.
func TestF10AdjuntosNoDebenSuperarElMaximoEnSubidasConcurrentes(t *testing.T) {
	ctx := context.Background()
	ahora := time.Now().UTC()
	base := memoriaReserva.NuevoRepositorioMemoria()

	reserva, err := NuevaReserva("e-f10", datosReservaValidos("slot-f10"), Slot{
		ID: "slot-f10", TutorID: "t-f10", Curso: "Álgebra",
		Inicio: ahora.Add(72 * time.Hour), Fin: ahora.Add(73 * time.Hour),
	}, ahora)
	if err != nil {
		t.Fatal(err)
	}
	guardada, _, err := base.ReservarTransaccional(ctx, reserva, LimitesPorDefecto)
	if err != nil {
		t.Fatal(err)
	}
	// Cuatro adjuntos previos: queda un solo cupo de los cinco permitidos.
	for i := 0; i < MaxAdjuntos-1; i++ {
		if _, err := base.AgregarAdjunto(ctx, guardada.ID, &Adjunto{
			Nombre: "previo.pdf", Tipo: "application/pdf", TamanoBytes: 10,
			Estado: AdjuntoValidado, Huella: "huella",
		}); err != nil {
			t.Fatal(err)
		}
	}

	repositorio := &repoConBarreraAdjuntos{RepositorioMemoria: base, liberar: make(chan struct{})}
	servicio := NuevoServicio(Dependencias{
		Repositorio: repositorio,
		Oferta:      NuevaOfertaSintetica(ahora),
		Almacen:     nuevoAlmacenEnMemoria(),
		Analizador:  AnalizadorBasico{},
	})

	pdf := []byte("%PDF-1.4\ncontenido de prueba\n%%EOF\n")
	var espera sync.WaitGroup
	errores := make(chan error, 2)
	for i := 0; i < 2; i++ {
		espera.Add(1)
		go func() {
			defer espera.Done()
			_, err := servicio.Adjuntar(ctx, "e-f10", guardada.ID, "nuevo.pdf", pdf)
			errores <- err
		}()
	}
	espera.Wait()
	close(errores)
	for err := range errores {
		if err != nil {
			t.Fatalf("una subida valida no debia fallar: %v", err)
		}
	}

	total, err := base.ContarAdjuntos(ctx, guardada.ID)
	if err != nil {
		t.Fatal(err)
	}
	if total > MaxAdjuntos {
		t.Fatalf("F-10: la reserva quedo con %d adjuntos tras dos subidas concurrentes; el maximo es %d", total, MaxAdjuntos)
	}
}

// repoQueCuentaSlotTomado detecta el N+1 de Slots.
type repoQueCuentaSlotTomado struct {
	*memoriaReserva.RepositorioMemoria
	mu        sync.Mutex
	consultas int
}

func (r *repoQueCuentaSlotTomado) SlotTomado(ctx context.Context, slotID string) (bool, error) {
	r.mu.Lock()
	r.consultas++
	r.mu.Unlock()
	return r.RepositorioMemoria.SlotTomado(ctx, slotID)
}

// F-17: listar N horarios no debe disparar N consultas de ocupacion.
func TestF17SlotsNoDebeConsultarOcupacionPorFila(t *testing.T) {
	ahora := time.Now().UTC()
	repositorio := &repoQueCuentaSlotTomado{RepositorioMemoria: memoriaReserva.NuevoRepositorioMemoria()}
	servicio := NuevoServicio(Dependencias{
		Repositorio: repositorio,
		Oferta:      NuevaOfertaSintetica(ahora),
	})
	servicio.FijarReloj(func() time.Time { return ahora })

	disposiciones, err := servicio.Slots(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	validos := 0
	for _, disposicion := range disposiciones {
		if disposicion.Disponible() {
			validos++
		}
	}
	if validos < 2 {
		t.Fatalf("la oferta sintetica debia tener varios horarios validos, hubo %d", validos)
	}
	repositorio.mu.Lock()
	consultas := repositorio.consultas
	repositorio.mu.Unlock()
	if consultas > 1 {
		t.Fatalf("F-17: Slots consulto la ocupacion %d veces para %d horarios validos (N+1); debe resolverse en una sola consulta", consultas, validos)
	}
}
