package sesion

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoriaSesion "uniapp/internal/adapters/repositories/memory/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloSesion "uniapp/internal/application/sesion"
	infra "uniapp/internal/infrastructure"
)

// reservasConmutables permite cancelar una reserva a mitad de la prueba.
type reservasConmutables struct {
	mutex      sync.Mutex
	canceladas map[string]bool
}

func (r *reservasConmutables) Cancelada(_ context.Context, reservaID string) (bool, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.canceladas[reservaID], nil
}

func (r *reservasConmutables) cancelar(reservaID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.canceladas[reservaID] = true
}

// F-28 (error de seguridad / orden de autorización): GET
// /api/sesion/{id}/minutos hecho por un usuario ajeno ejecuta primero
// servicio.Consultar, que puede anular y PERSISTIR el registro, y recién
// después responde 403. Un usuario sin permiso provoca una escritura.
func TestF28AjenoNoDebeProvocarEscriturasAntesDel403(t *testing.T) {
	ctx := context.Background()
	repositorioAcceso := memoriaAcceso.NuevoRepositorioMemoria()
	hasher := infra.PBKDF2Hasher{Iteraciones: 1000}
	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: repositorioAcceso,
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         infra.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(ctx, servicioAcceso, modeloAcceso.DatosSiembra{DominioTutor: "uni.edu.pe", Clave: claveI05}); err != nil {
		t.Fatal(err)
	}
	tutor, err := repositorioAcceso.UsuarioPorEmail(ctx, "tutor.prueba@uni.edu.pe")
	if err != nil {
		t.Fatal(err)
	}
	escolar, err := repositorioAcceso.UsuarioPorEmail(ctx, "escolar.prueba@example.test")
	if err != nil {
		t.Fatal(err)
	}

	sesiones := memoriaSesion.NuevoRepositorioMemoria()
	s, err := modeloSesion.NuevaSesion("sesion-f28", "reserva-f28", tutor.ID, escolar.ID, inicioSesionI05, inicioSesionI05.Add(modeloSesion.DuracionSesion))
	if err != nil {
		t.Fatal(err)
	}
	if err := sesiones.CrearSesion(ctx, s); err != nil {
		t.Fatal(err)
	}
	for _, participante := range []string{tutor.ID, escolar.ID} {
		if _, err := sesiones.DeclararAsistencia(ctx, "sesion-f28", participante, inicioSesionI05); err != nil {
			t.Fatal(err)
		}
	}

	repositorioMinutos := memoriaSesion.NuevoRepositorioMinutosMemoria()
	reservas := &reservasConmutables{canceladas: map[string]bool{}}
	servicio := modeloSesion.ServicioMinutos{Sesiones: sesiones, Minutos: repositorioMinutos, Reservas: reservas}
	despuesDelFin := inicioSesionI05.Add(modeloSesion.DuracionSesion + time.Minute)
	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	rutas := RutasMinutos(servicio, controladorAcceso.Autenticar(), func() time.Time { return despuesDelFin })
	mux := http.NewServeMux()
	mux.Handle("/api/sesion/{id}/minutos", rutas)
	mux.Handle("/", controladorAcceso.Rutas())
	e := entornoI17{manejador: ctrlAcceso.Encadenar(mux, ctrlAcceso.OrigenPermitido(origenI05))}

	cookieTutor := iniciarSesionI05(t, e.manejador, "tutor.prueba@uni.edu.pe")
	ajeno := iniciarSesionI05(t, e.manejador, "apoderado.prueba@example.test")
	solicitarI17(t, e, cookieTutor, "POST", "/api/sesion/sesion-f28/minutos", `{"minutos":45}`, 201)

	reservas.cancelar("reserva-f28")
	solicitarI17(t, e, ajeno, "GET", "/api/sesion/sesion-f28/minutos", "", 403)

	guardado, err := repositorioMinutos.MinutosPorSesion(ctx, "sesion-f28")
	if err != nil {
		t.Fatal(err)
	}
	if guardado.Estado() != modeloSesion.MinutosPropuestos {
		t.Fatalf("F-28: un GET rechazado con 403 dejó el registro en %q (escribió antes de autorizar)", guardado.Estado())
	}
}
