package sesion

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/sesion"
	dominio "uniapp/internal/domain/sesion"
)

type nombresFalsos map[string]string

func (n nombresFalsos) NombreDe(_ context.Context, id string) (string, error) { return n[id], nil }

type nombresConError struct{}

func (nombresConError) NombreDe(context.Context, string) (string, error) {
	return "", errors.New("acceso caido")
}

var desdeI18 = time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)

type escenarioI18 struct {
	servicio ServicioReporte
	reservas *reservasFalsas
	ahora    time.Time
	filtro   FiltroReporte
}

// prepararReporteI18: sesion-1 (45 verificados), sesion-2 (30 propuestos),
// sesion-3 (sin minutos), sesion-fuera (fuera del periodo).
func prepararReporteI18(t *testing.T) escenarioI18 {
	t.Helper()
	ctx := context.Background()
	sesiones := memoria.NuevoRepositorioMemoria()
	reservas := &reservasFalsas{canceladas: map[string]bool{}}
	minutos := ServicioMinutos{Sesiones: sesiones, Minutos: memoria.NuevoRepositorioMinutosMemoria(), Reservas: reservas}

	crear := func(id, tutor, escolar string, inicio time.Time) {
		s, err := dominio.NuevaSesion(id, "reserva-"+id, tutor, escolar, inicio, inicio.Add(dominio.DuracionSesion))
		if err != nil {
			t.Fatal(err)
		}
		if err := sesiones.CrearSesion(ctx, s); err != nil {
			t.Fatal(err)
		}
		for _, u := range []string{tutor, escolar} {
			if _, err := sesiones.DeclararAsistencia(ctx, id, u, inicio); err != nil {
				t.Fatal(err)
			}
		}
	}
	base := desdeI18.Add(24 * time.Hour)
	crear("sesion-1", "tutor-a", "escolar-1", base)
	crear("sesion-2", "tutor-a", "escolar-2", base.Add(2*time.Hour))
	crear("sesion-3", "tutor-b", "escolar-1", base.Add(4*time.Hour))
	crear("sesion-fuera", "tutor-a", "escolar-1", desdeI18.AddDate(0, 0, 40))

	ahora := base.Add(6 * time.Hour)
	if _, err := minutos.Proponer(ctx, "sesion-1", "tutor-a", 45, ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := minutos.Aprobar(ctx, "sesion-1", "coord", ahora); err != nil {
		t.Fatal(err)
	}
	if _, err := minutos.Proponer(ctx, "sesion-2", "tutor-a", 30, ahora); err != nil {
		t.Fatal(err)
	}

	seudonimos, err := dominio.NuevoSeudonimizador([]byte("clave-de-prueba-i18-suficiente"))
	if err != nil {
		t.Fatal(err)
	}
	filtro, err := dominio.NuevoFiltroReporte(desdeI18, desdeI18.AddDate(0, 0, 30), "")
	if err != nil {
		t.Fatal(err)
	}
	return escenarioI18{
		servicio: ServicioReporte{
			Reportes:   memoria.NuevoRepositorioReportesMemoria(sesiones),
			Minutos:    minutos,
			Nombres:    nombresFalsos{"tutor-a": "Ana Quispe", "coord": "Coordinación Prueba"},
			Seudonimos: seudonimos,
			NuevoID:    func() (string, error) { return "rep-fijo", nil },
		},
		reservas: reservas,
		ahora:    ahora,
		filtro:   filtro,
	}
}

func TestI18ServicioGeneraConMinutosDeI17(t *testing.T) {
	e := prepararReporteI18(t)
	reporte, err := e.servicio.Generar(context.Background(), e.filtro, "coord", e.ahora)
	if err != nil {
		t.Fatal(err)
	}
	if reporte.GeneradoPorNombre != "Coordinación Prueba" {
		t.Fatalf("debe resolver el nombre del coordinador: %q", reporte.GeneradoPorNombre)
	}
	if reporte.ID != "rep-fijo" || len(reporte.Filas) != 3 {
		t.Fatalf("esperaba 3 filas del periodo: %+v", reporte)
	}
	if reporte.Filas[0].MinutosComputables != 45 || reporte.Filas[1].MinutosComputables != 0 || reporte.Filas[2].EstadoMinutos != dominio.EstadoSinRegistro {
		t.Fatalf("filas inesperadas: %+v", reporte.Filas)
	}
	if reporte.Filas[0].TutorNombre != "Ana Quispe" || reporte.Filas[2].TutorNombre != "tutor-b" {
		t.Fatal("debe usar el nombre conocido y, si no hay, el ID")
	}
	if reporte.Totales[0].MinutosVerificados != 45 {
		t.Fatalf("Ana debe sumar 45: %+v", reporte.Totales)
	}
}

func TestI18ServicioCancelacionSumaCero(t *testing.T) {
	e := prepararReporteI18(t)
	e.reservas.canceladas["reserva-sesion-1"] = true
	reporte, err := e.servicio.Generar(context.Background(), e.filtro, "coord", e.ahora)
	if err != nil {
		t.Fatal(err)
	}
	if reporte.Filas[0].EstadoMinutos != dominio.MinutosAnulados || reporte.Filas[0].MinutosComputables != 0 {
		t.Fatalf("una reserva cancelada debe salir anulada y sumar 0: %+v", reporte.Filas[0])
	}
	if reporte.Filas[0].UltimaAccion == nil || reporte.Filas[0].UltimaAccion.Accion != dominio.AccionAnular {
		t.Fatal("la anulacion debe verse como ultima accion de auditoria")
	}
}

// El reporte es una lectura: ver una reserva cancelada no debe escribir la
// anulacion en el repositorio (no repetir F-06 de Consultar).
func TestI18ServicioNoEscribeAlVerReservaCancelada(t *testing.T) {
	e := prepararReporteI18(t)
	e.reservas.canceladas["reserva-sesion-1"] = true
	if _, err := e.servicio.Generar(context.Background(), e.filtro, "coord", e.ahora); err != nil {
		t.Fatal(err)
	}
	guardado, err := e.servicio.Minutos.Minutos.MinutosPorSesion(context.Background(), "sesion-1")
	if err != nil {
		t.Fatal(err)
	}
	if guardado.Estado() != dominio.MinutosVerificados {
		t.Fatalf("generar el reporte no debe persistir la anulacion: %s", guardado.Estado())
	}
}

func TestI18VistaSinRegistroDevuelveNoEncontrado(t *testing.T) {
	e := prepararReporteI18(t)
	if _, err := e.servicio.Minutos.Vista(context.Background(), "sesion-3", e.ahora); !errors.Is(err, dominio.ErrMinutosNoEncontrados) {
		t.Fatalf("esperaba ErrMinutosNoEncontrados: %v", err)
	}
}

func TestI18ServicioRegistraDescargaYLista(t *testing.T) {
	e := prepararReporteI18(t)
	ctx := context.Background()
	reporte, err := e.servicio.Generar(ctx, e.filtro, "coord", e.ahora)
	if err != nil {
		t.Fatal(err)
	}
	registro, err := e.servicio.RegistrarDescarga(ctx, reporte, dominio.FormatoCSV, []byte("a,b\n"))
	if err != nil {
		t.Fatal(err)
	}
	if registro.Filas != 3 || len(registro.SHA256) != 64 {
		t.Fatalf("constancia inesperada: %+v", registro)
	}
	if _, err := e.servicio.RegistrarDescarga(ctx, reporte, dominio.FormatoCSV, []byte("otra")); err == nil {
		t.Fatal("el mismo ID de reporte no puede registrarse dos veces")
	}
	recientes, err := e.servicio.Recientes(ctx)
	if err != nil || len(recientes) != 1 || recientes[0].ID != "rep-fijo" {
		t.Fatalf("esperaba una constancia: %+v %v", recientes, err)
	}
}

func TestI18ServicioPropagaErrorDelDirectorio(t *testing.T) {
	e := prepararReporteI18(t)
	e.servicio.Nombres = nombresConError{}
	if _, err := e.servicio.Generar(context.Background(), e.filtro, "coord", e.ahora); err == nil || !strings.Contains(err.Error(), "acceso caido") {
		t.Fatalf("un fallo al leer nombres no debe ocultarse: %v", err)
	}
}

func TestI18ServicioIDAleatorio(t *testing.T) {
	e := prepararReporteI18(t)
	e.servicio.NuevoID = nil
	a, err := e.servicio.Generar(context.Background(), e.filtro, "coord", e.ahora)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := e.servicio.Generar(context.Background(), e.filtro, "coord", e.ahora)
	if !strings.HasPrefix(a.ID, "rep-") || a.ID == b.ID {
		t.Fatalf("los IDs deben ser unicos: %s %s", a.ID, b.ID)
	}
}
