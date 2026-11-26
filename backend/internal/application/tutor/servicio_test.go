package tutor

import (
	"context"
	"sync"
	"testing"
	"time"

	memoriaTutor "uniapp/internal/adapters/repositories/memory/tutor"
)

type almacenDocumentosEnMemoria struct {
	mu    sync.Mutex
	datos map[string][]byte
}

func nuevoAlmacenDocumentosEnMemoria() *almacenDocumentosEnMemoria {
	return &almacenDocumentosEnMemoria{datos: map[string][]byte{}}
}

func (a *almacenDocumentosEnMemoria) Guardar(_ context.Context, id string, contenido []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.datos[id] = append([]byte(nil), contenido...)
	return nil
}

func (a *almacenDocumentosEnMemoria) Leer(_ context.Context, id string) ([]byte, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	contenido, existe := a.datos[id]
	if !existe {
		return nil, ErrNoEncontrado
	}
	return append([]byte(nil), contenido...), nil
}

func (a *almacenDocumentosEnMemoria) Eliminar(_ context.Context, id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.datos, id)
	return nil
}

func (a *almacenDocumentosEnMemoria) cantidad() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.datos)
}

// repoConBarreraPresentaciones fuerza a que dos Presentar lean el mismo
// expediente anterior antes de que cualquiera guarde (B-16).
type repoConBarreraPresentaciones struct {
	*memoriaTutor.RepositorioMemoria
	mu       sync.Mutex
	llegadas int
	liberar  chan struct{}
}

func (r *repoConBarreraPresentaciones) PorTutor(ctx context.Context, tutorID string) (*Expediente, error) {
	expediente, err := r.RepositorioMemoria.PorTutor(ctx, tutorID)
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
	return expediente, err
}

func datosExpedientePrueba(nombre string) DatosExpediente {
	return DatosExpediente{
		Presentacion: "Presentación de prueba con experiencia docente.",
		Cursos:       "Álgebra y geometría",
		Experiencia:  "Experiencia de prueba en acompañamiento académico.",
		Documentos: []Archivo{{
			Nombre:    nombre,
			Contenido: []byte("%PDF-1.4\nconstancia de prueba\n%%EOF\n"),
		}},
	}
}

// B-16: dos presentaciones concurrentes no deben dejar documentos huerfanos
// en el almacen. Solo deben sobrevivir los documentos del expediente vigente.
func TestB16PresentacionesConcurrentesNoDebenDejarDocumentosHuerfanos(t *testing.T) {
	ctx := context.Background()
	base := memoriaTutor.NuevoRepositorioMemoria()
	almacen := nuevoAlmacenDocumentosEnMemoria()

	inicial := NuevoServicio(base, almacen)
	if _, err := inicial.Presentar(ctx, "t-b16", datosExpedientePrueba("anterior.pdf")); err != nil {
		t.Fatal(err)
	}

	concurrente := &repoConBarreraPresentaciones{RepositorioMemoria: base, liberar: make(chan struct{})}
	servicio := NuevoServicio(concurrente, almacen)

	var espera sync.WaitGroup
	errores := make(chan error, 2)
	for i := 0; i < 2; i++ {
		espera.Add(1)
		go func(indice int) {
			defer espera.Done()
			nombre := "nuevo-a.pdf"
			if indice == 1 {
				nombre = "nuevo-b.pdf"
			}
			_, err := servicio.Presentar(ctx, "t-b16", datosExpedientePrueba(nombre))
			errores <- err
		}(i)
	}
	espera.Wait()
	close(errores)
	for err := range errores {
		if err != nil {
			t.Fatalf("una presentación válida no debía fallar: %v", err)
		}
	}

	vigente, err := base.PorTutor(ctx, "t-b16")
	if err != nil {
		t.Fatal(err)
	}
	if cantidad := almacen.cantidad(); cantidad != len(vigente.Documentos) {
		t.Fatalf("B-16: quedaron %d documentos en el almacén pero el expediente vigente referencia %d; las presentaciones concurrentes dejaron huérfanos", cantidad, len(vigente.Documentos))
	}
}
