package sesion

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"sync"
)

// RepositorioMinutosMemoria implementa RepositorioMinutos (I17) para pruebas.
type RepositorioMinutosMemoria struct {
	mutex     sync.Mutex
	registros map[string]*RegistroMinutos
}

func NuevoRepositorioMinutosMemoria() *RepositorioMinutosMemoria {
	return &RepositorioMinutosMemoria{registros: make(map[string]*RegistroMinutos)}
}

func (r *RepositorioMinutosMemoria) CrearMinutos(_ context.Context, registro *RegistroMinutos) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, existe := r.registros[registro.SesionID()]; existe {
		return ErrMinutosDuplicados
	}
	r.registros[registro.SesionID()] = registro.Clonar()
	return nil
}

func (r *RepositorioMinutosMemoria) MinutosPorSesion(_ context.Context, sesionID string) (*RegistroMinutos, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	registro, existe := r.registros[sesionID]
	if !existe {
		return nil, ErrMinutosNoEncontrados
	}
	return registro.Clonar(), nil
}

// ActualizarMinutos aplica el cambio sobre una copia y solo la guarda si el
// cambio no fallo: un rechazo no deja estados a medias.
func (r *RepositorioMinutosMemoria) ActualizarMinutos(
	_ context.Context, sesionID string, cambio func(*RegistroMinutos) error,
) (*RegistroMinutos, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	registro, existe := r.registros[sesionID]
	if !existe {
		return nil, ErrMinutosNoEncontrados
	}
	copia := registro.Clonar()
	if err := cambio(copia); err != nil {
		return nil, err
	}
	r.registros[sesionID] = copia
	return copia.Clonar(), nil
}

func (r *RepositorioMinutosMemoria) MinutosPendientes(_ context.Context) ([]*RegistroMinutos, error) {
	return r.filtrar(func(registro *RegistroMinutos) bool {
		return registro.Estado() == MinutosPropuestos || registro.Estado() == MinutosEnDisputa
	}), nil
}

func (r *RepositorioMinutosMemoria) MinutosDeTutor(_ context.Context, tutorID string) ([]*RegistroMinutos, error) {
	return r.filtrar(func(registro *RegistroMinutos) bool {
		return registro.TutorID() == tutorID
	}), nil
}

func (r *RepositorioMinutosMemoria) filtrar(incluir func(*RegistroMinutos) bool) []*RegistroMinutos {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	resultado := make([]*RegistroMinutos, 0)
	for _, registro := range r.registros {
		if incluir(registro) {
			resultado = append(resultado, registro.Clonar())
		}
	}
	// Orden estable (por fin de sesion y luego id), igual que en PostgreSQL.
	sort.Slice(resultado, func(i, j int) bool {
		if !resultado[i].FinSesion().Equal(resultado[j].FinSesion()) {
			return resultado[i].FinSesion().Before(resultado[j].FinSesion())
		}
		return resultado[i].SesionID() < resultado[j].SesionID()
	})
	return resultado
}

// CancelacionesFijas implementa el puerto EstadoReservas con un conjunto fijo
// de reservas canceladas. Es el adaptador provisional hasta que el flujo D
// publique la cancelacion (I16); reemplazarlo solo cambia cmd/api/main.go.
type CancelacionesFijas struct {
	canceladas map[string]bool
}

func NuevasCancelacionesFijas(reservaIDs ...string) CancelacionesFijas {
	canceladas := make(map[string]bool, len(reservaIDs))
	for _, id := range reservaIDs {
		canceladas[id] = true
	}
	return CancelacionesFijas{canceladas: canceladas}
}

func (c CancelacionesFijas) Cancelada(_ context.Context, reservaID string) (bool, error) {
	return c.canceladas[reservaID], nil
}

// fixtureCancelaciones es el formato de backend/fixtures/sesion-cancelaciones.json.
type fixtureCancelaciones struct {
	ReservasCanceladas []string `json:"reservas_canceladas"`
}

// CargarCancelacionesFixture lee el fixture sintetico de cancelaciones.
func CargarCancelacionesFixture(ruta string) (CancelacionesFijas, error) {
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		return CancelacionesFijas{}, err
	}
	var datos fixtureCancelaciones
	if err := json.Unmarshal(contenido, &datos); err != nil {
		return CancelacionesFijas{}, err
	}
	return NuevasCancelacionesFijas(datos.ReservasCanceladas...), nil
}
