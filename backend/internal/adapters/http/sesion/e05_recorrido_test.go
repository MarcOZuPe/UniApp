package sesion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// cancelacionesE05 es el puerto hacia el flujo D: permite cancelar una
// reserva a mitad del recorrido, como haria I16.
type cancelacionesE05 struct {
	mutex      sync.Mutex
	canceladas map[string]bool
}

func (c *cancelacionesE05) Cancelada(_ context.Context, reservaID string) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.canceladas[reservaID], nil
}

func (c *cancelacionesE05) cancelar(reservaID string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.canceladas[reservaID] = true
}

// TestE05RecorridoSesionAReporte recorre el flujo E completo por HTTP, con
// los mismos controladores que usa main.go y repositorios en memoria:
//
//	I05 asistencia -> I11 ausencia y resultado -> I17 minutos, disputa y
//	resolucion -> cancelacion (flujo D) -> I18 reporte CSV/PDF con constancia.
//
// Cubre los cuatro casos de E05: sesion completa con disputa, sesion parcial
// (el escolar no asistio), reserva cancelada y el reporte final.
func TestE05RecorridoSesionAReporte(t *testing.T) {
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

	// Tres sesiones del 20/10/2026, cada una con su reserva (flujo D).
	sesiones := memoriaSesion.NuevoRepositorioMemoria()
	ids := []string{"e05-completa", "e05-parcial", "e05-cancelada"}
	inicios := map[string]time.Time{
		"e05-completa":  inicioSesionI05,
		"e05-parcial":   inicioSesionI05.Add(2 * time.Hour),
		"e05-cancelada": inicioSesionI05.Add(4 * time.Hour),
	}
	for _, id := range ids {
		inicio := inicios[id]
		s, err := modeloSesion.NuevaSesion(id, "reserva-"+id, tutor.ID, escolar.ID, inicio, inicio.Add(modeloSesion.DuracionSesion))
		if err != nil {
			t.Fatal(err)
		}
		if err := sesiones.CrearSesion(ctx, s); err != nil {
			t.Fatal(err)
		}
	}

	cancelaciones := &cancelacionesE05{canceladas: map[string]bool{}}
	minutos := modeloSesion.ServicioMinutos{Sesiones: sesiones, Minutos: memoriaSesion.NuevoRepositorioMinutosMemoria(), Reservas: cancelaciones}
	seudonimos, err := modeloSesion.NuevoSeudonimizador([]byte("clave-de-prueba-e05-suficiente"))
	if err != nil {
		t.Fatal(err)
	}
	reporte := modeloSesion.ServicioReporte{
		Reportes:   memoriaSesion.NuevoRepositorioReportesMemoria(sesiones),
		Minutos:    minutos,
		Nombres:    nombresDeAcceso{repo: repositorioAcceso},
		Seudonimos: seudonimos,
	}

	// Reloj controlado: el recorrido avanza por las ventanas de cada plazo.
	var relojMutex sync.Mutex
	momento := inicioSesionI05
	reloj := func() time.Time {
		relojMutex.Lock()
		defer relojMutex.Unlock()
		return momento
	}
	fijar := func(m time.Time) {
		relojMutex.Lock()
		defer relojMutex.Unlock()
		momento = m
	}

	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	autenticar := controladorAcceso.Autenticar()
	rutasSesion := Rutas(sesiones, autenticar, origenI05, ConReloj(reloj))
	rutasMinutos := RutasMinutos(minutos, autenticar, reloj)
	mux := http.NewServeMux()
	mux.Handle("/api/sesion/", rutasSesion)
	mux.Handle("/api/sesion/{id}/minutos", rutasMinutos)
	mux.Handle("/api/sesion/{id}/minutos/disputa", rutasMinutos)
	mux.Handle("/api/minutos/", rutasMinutos)
	mux.Handle("/api/reportes/", RutasReporte(reporte, autenticar, reloj))
	mux.Handle("/", controladorAcceso.Rutas())
	e := entornoI17{manejador: ctrlAcceso.Encadenar(mux, ctrlAcceso.OrigenPermitido(origenI05))}

	cTutor := iniciarSesionI05(t, e.manejador, "tutor.prueba@uni.edu.pe")
	cEscolar := iniciarSesionI05(t, e.manejador, "escolar.prueba@example.test")
	cCoord := iniciarSesionI05(t, e.manejador, "coord.prueba@example.test")

	// 1. Durante cada sesion: I05 asistencia e I11 ausencia (sesion parcial).
	for _, id := range ids {
		fijar(inicios[id].Add(10 * time.Minute))
		solicitarI17(t, e, cTutor, "POST", "/api/sesion/"+id+"/asistencias", "", 201)
		if id == "e05-parcial" {
			solicitarI17(t, e, cEscolar, "POST", "/api/sesion/"+id+"/ausencias", "", 201)
			continue
		}
		solicitarI17(t, e, cEscolar, "POST", "/api/sesion/"+id+"/asistencias", "", 201)
	}

	// 2. Despues de la ultima sesion: I11 resultado e I17 minutos.
	fijar(inicios["e05-cancelada"].Add(2 * time.Hour))
	solicitarI17(t, e, cTutor, "POST", "/api/sesion/e05-completa/resultado", `{"resumen":"Repasamos ecuaciones lineales."}`, 201)
	solicitarI17(t, e, cTutor, "POST", "/api/sesion/e05-completa/minutos", `{"minutos":45}`, 201)
	solicitarI17(t, e, cTutor, "POST", "/api/sesion/e05-cancelada/minutos", `{"minutos":60}`, 201)

	// Sesion parcial: sin asistencia de ambos no hay minutos que proponer.
	parcial := solicitarI17(t, e, cTutor, "POST", "/api/sesion/e05-parcial/minutos", `{"minutos":30}`, 422)
	if parcial["codigo"] != "asistencia_incompleta" {
		t.Fatalf("la sesion parcial debe rechazarse por asistencia incompleta: %v", parcial)
	}

	// 3. Coordinacion aprueba; el escolar disputa; coordinacion resuelve.
	solicitarI17(t, e, cCoord, "POST", "/api/minutos/e05-completa/aprobar", "", 200)
	solicitarI17(t, e, cCoord, "POST", "/api/minutos/e05-cancelada/aprobar", "", 200)
	disputa := solicitarI17(t, e, cEscolar, "POST", "/api/sesion/e05-completa/minutos/disputa", `{"motivo":"Empezamos 10 minutos tarde"}`, 200)
	if disputa["minutos_computables"] != float64(0) {
		t.Fatalf("una disputa congela los minutos: %v", disputa)
	}
	solicitarI17(t, e, cCoord, "POST", "/api/minutos/e05-completa/resolver", `{"minutos":35,"motivo":"Confirmado con ambos"}`, 200)

	total := solicitarI17(t, e, cTutor, "GET", "/api/minutos/total", "", 200)
	if total["minutos_verificados"] != float64(95) {
		t.Fatalf("antes de cancelar el tutor debe sumar 35 + 60 = 95: %v", total)
	}

	// 4. El flujo D cancela una reserva: esos minutos pasan a sumar cero.
	cancelaciones.cancelar("reserva-e05-cancelada")

	// 5. I18: coordinacion descarga el reporte del periodo.
	respuesta := pedirReporteE05(t, e, cCoord, "/api/reportes/minutos?desde=2026-10-20&hasta=2026-10-20&formato=csv")
	csv := respuesta.Body.String()
	if strings.Contains(csv, escolar.ID) {
		t.Fatal("el reporte no puede exponer el ID del escolar")
	}
	for _, esperado := range []string{"e05-completa", "e05-parcial", "e05-cancelada", "Sin validez oficial", "ESC-"} {
		if !strings.Contains(csv, esperado) {
			t.Fatalf("el CSV debe contener %q:\n%s", esperado, csv)
		}
	}
	filas := map[string]string{}
	for _, linea := range strings.Split(csv, "\n") {
		for _, id := range ids {
			if strings.Contains(linea, id+",") {
				filas[id] = linea
			}
		}
	}
	if !strings.Contains(filas["e05-completa"], "verificado") || !strings.Contains(filas["e05-completa"], ",45,35,") {
		t.Fatalf("la sesion completa debe salir verificada con 45 propuestos y 35 que cuentan: %s", filas["e05-completa"])
	}
	if !strings.Contains(filas["e05-parcial"], "ausencia_reportada") || !strings.Contains(filas["e05-parcial"], "sin_registro") {
		t.Fatalf("la sesion parcial debe mostrar la ausencia y no tener minutos: %s", filas["e05-parcial"])
	}
	if !strings.Contains(filas["e05-cancelada"], "anulado") || !strings.Contains(filas["e05-cancelada"], ",60,0,") {
		t.Fatalf("la sesion cancelada debe salir anulada y sumar 0: %s", filas["e05-cancelada"])
	}

	pdf := pedirReporteE05(t, e, cCoord, "/api/reportes/minutos?desde=2026-10-20&hasta=2026-10-20&formato=pdf")
	if !strings.HasPrefix(pdf.Body.String(), "%PDF-1.4") || strings.Contains(pdf.Body.String(), escolar.ID) {
		t.Fatal("el PDF debe generarse y no exponer al escolar")
	}

	// 6. Trazabilidad: una constancia por descarga, con la huella entregada.
	registro := pedirReporteE05(t, e, cCoord, "/api/reportes/registro")
	var constancias struct {
		Reportes []struct {
			SHA256 string `json:"sha256"`
			Filas  int    `json:"filas"`
		} `json:"reportes"`
	}
	if err := json.Unmarshal(registro.Body.Bytes(), &constancias); err != nil {
		t.Fatal(err)
	}
	if len(constancias.Reportes) != 2 || constancias.Reportes[0].Filas != 3 {
		t.Fatalf("esperaba 2 constancias de 3 filas: %+v", constancias.Reportes)
	}
	huellas := map[string]bool{constancias.Reportes[0].SHA256: true, constancias.Reportes[1].SHA256: true}
	if !huellas[respuesta.Header().Get("X-Reporte-Sha256")] || !huellas[pdf.Header().Get("X-Reporte-Sha256")] {
		t.Fatal("cada constancia debe guardar la huella del archivo entregado")
	}

	// El total del tutor ya refleja la cancelacion (solo cuentan los 35).
	total = solicitarI17(t, e, cTutor, "GET", "/api/minutos/total", "", 200)
	if total["minutos_verificados"] != float64(35) {
		t.Fatalf("tras la cancelacion el tutor debe sumar 35: %v", total)
	}
}

func pedirReporteE05(t *testing.T, e entornoI17, cookie *http.Cookie, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	peticion := httptest.NewRequest("GET", ruta, nil)
	peticion.Header.Set("Origin", origenI05)
	peticion.AddCookie(cookie)
	respuesta := httptest.NewRecorder()
	e.manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("GET %s: esperado 200, recibido %d: %s", ruta, respuesta.Code, respuesta.Body.String())
	}
	return respuesta
}
