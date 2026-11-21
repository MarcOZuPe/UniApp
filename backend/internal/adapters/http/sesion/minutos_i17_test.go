package sesion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/adapters/http/acceso"
	memoriaAcceso "uniapp/internal/adapters/repositories/memory/acceso"
	memoriaSesion "uniapp/internal/adapters/repositories/memory/sesion"
	modeloAcceso "uniapp/internal/application/acceso"
	modeloSesion "uniapp/internal/application/sesion"
	infra "uniapp/internal/infrastructure"
)

const reservaCanceladaI17 = "reserva-cancelada-i17"

type entornoI17 struct {
	manejador http.Handler
	tutorID   string
	escolarID string
}

// prepararEntornoI17 arma acceso + I17 en memoria con dos sesiones en las que
// ambos participantes declararon asistencia: "sesion-i17" (normal) y
// "sesion-i17-cancelada" (su reserva figura cancelada en el puerto).
func prepararEntornoI17(t *testing.T) entornoI17 {
	t.Helper()
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
	for id, reserva := range map[string]string{"sesion-i17": "reserva-i17", "sesion-i17-cancelada": reservaCanceladaI17} {
		s, err := modeloSesion.NuevaSesion(id, reserva, tutor.ID, escolar.ID, inicioSesionI05, inicioSesionI05.Add(modeloSesion.DuracionSesion))
		if err != nil {
			t.Fatal(err)
		}
		if err := sesiones.CrearSesion(ctx, s); err != nil {
			t.Fatal(err)
		}
		for _, participante := range []string{tutor.ID, escolar.ID} {
			if _, err := sesiones.DeclararAsistencia(ctx, id, participante, inicioSesionI05); err != nil {
				t.Fatal(err)
			}
		}
	}

	servicio := modeloSesion.ServicioMinutos{
		Sesiones: sesiones,
		Minutos:  memoriaSesion.NuevoRepositorioMinutosMemoria(),
		Reservas: memoriaSesion.NuevasCancelacionesFijas(reservaCanceladaI17),
	}
	despuesDelFin := inicioSesionI05.Add(modeloSesion.DuracionSesion + time.Minute)
	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)
	rutasMinutos := RutasMinutos(servicio, controladorAcceso.Autenticar(), func() time.Time { return despuesDelFin })

	mux := http.NewServeMux()
	mux.Handle("/api/sesion/{id}/minutos", rutasMinutos)
	mux.Handle("/api/sesion/{id}/minutos/disputa", rutasMinutos)
	mux.Handle("/api/minutos/", rutasMinutos)
	mux.Handle("/", controladorAcceso.Rutas())
	return entornoI17{
		manejador: ctrlAcceso.Encadenar(mux, ctrlAcceso.OrigenPermitido(origenI05)),
		tutorID:   tutor.ID,
		escolarID: escolar.ID,
	}
}

func solicitarI17(t *testing.T, e entornoI17, cookie *http.Cookie, metodo, ruta, cuerpo string, esperado int) map[string]any {
	t.Helper()
	var lector *strings.Reader
	if cuerpo != "" {
		lector = strings.NewReader(cuerpo)
	}
	var peticion *http.Request
	if lector != nil {
		peticion = httptest.NewRequest(metodo, ruta, lector)
		peticion.Header.Set("Content-Type", "application/json")
	} else {
		peticion = httptest.NewRequest(metodo, ruta, nil)
	}
	peticion.Header.Set("Origin", origenI05)
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	respuesta := httptest.NewRecorder()
	e.manejador.ServeHTTP(respuesta, peticion)
	if respuesta.Code != esperado {
		t.Fatalf("%s %s: esperado %d, recibido %d: %s", metodo, ruta, esperado, respuesta.Code, respuesta.Body.String())
	}
	datos := map[string]any{}
	_ = json.Unmarshal(respuesta.Body.Bytes(), &datos)
	return datos
}

func TestI17FlujoCompletoHTTP(t *testing.T) {
	e := prepararEntornoI17(t)
	tutor := iniciarSesionI05(t, e.manejador, "tutor.prueba@uni.edu.pe")
	escolar := iniciarSesionI05(t, e.manejador, "escolar.prueba@example.test")
	coordinador := iniciarSesionI05(t, e.manejador, "coord.prueba@example.test")
	ajeno := iniciarSesionI05(t, e.manejador, "apoderado.prueba@example.test")
	base := "/api/sesion/sesion-i17/minutos"

	// Proponer: solo el tutor, con entrada estricta.
	solicitarI17(t, e, nil, "POST", base, `{"minutos":45}`, 401)
	solicitarI17(t, e, escolar, "POST", base, `{"minutos":45}`, 403)
	for _, cuerpo := range []string{`{}`, `{"minutos":45,"estado":"verificado"}`, `{"minutos":"45"}`, `{"minutos":45}{"minutos":45}`, `{"minutos":45,"motivo":"x"}`} {
		solicitarI17(t, e, tutor, "POST", base, cuerpo, 400)
	}
	solicitarI17(t, e, tutor, "POST", base, `{"minutos":61}`, 422)
	creado := solicitarI17(t, e, tutor, "POST", base, `{"minutos":45}`, 201)
	if creado["estado"] != "propuesto" || creado["minutos_computables"] != float64(0) {
		t.Fatalf("propuesta incorrecta: %v", creado)
	}
	solicitarI17(t, e, tutor, "POST", base, `{"minutos":45}`, 409)

	// Consultar: participantes y coordinacion; nadie mas.
	solicitarI17(t, e, escolar, "GET", base, "", 200)
	solicitarI17(t, e, coordinador, "GET", base, "", 200)
	solicitarI17(t, e, ajeno, "GET", base, "", 403)

	// Bandeja y aprobacion: solo coordinacion (no autoaprueba).
	solicitarI17(t, e, tutor, "GET", "/api/minutos/pendientes", "", 403)
	pendientes := solicitarI17(t, e, coordinador, "GET", "/api/minutos/pendientes", "", 200)
	if lista, _ := pendientes["pendientes"].([]any); len(lista) != 1 {
		t.Fatalf("la bandeja debe tener 1 pendiente: %v", pendientes)
	}
	solicitarI17(t, e, tutor, "POST", "/api/minutos/sesion-i17/aprobar", "", 403)
	aprobado := solicitarI17(t, e, coordinador, "POST", "/api/minutos/sesion-i17/aprobar", "", 200)
	if aprobado["estado"] != "verificado" || aprobado["minutos_computables"] != float64(45) {
		t.Fatalf("45 minutos deben sumar 45: %v", aprobado)
	}

	// Disputa del escolar: congela.
	solicitarI17(t, e, escolar, "POST", base+"/disputa", `{"motivo":""}`, 422)
	solicitarI17(t, e, escolar, "POST", base+"/disputa", `{"minutos":10,"motivo":"x"}`, 400)
	disputa := solicitarI17(t, e, escolar, "POST", base+"/disputa", `{"motivo":"Empezamos 15 minutos tarde"}`, 200)
	if disputa["estado"] != "en_disputa" || disputa["minutos_computables"] != float64(0) {
		t.Fatalf("la disputa debe congelar: %v", disputa)
	}
	solicitarI17(t, e, coordinador, "POST", "/api/minutos/sesion-i17/aprobar", "", 409)

	// Resolucion con motivo y total.
	solicitarI17(t, e, coordinador, "POST", "/api/minutos/sesion-i17/resolver", `{"minutos":30}`, 422)
	resuelto := solicitarI17(t, e, coordinador, "POST", "/api/minutos/sesion-i17/resolver", `{"minutos":30,"motivo":"Revisado con ambos"}`, 200)
	if resuelto["minutos_computables"] != float64(30) {
		t.Fatalf("resolucion incorrecta: %v", resuelto)
	}
	total := solicitarI17(t, e, tutor, "GET", "/api/minutos/total", "", 200)
	if total["minutos_verificados"] != float64(30) {
		t.Fatalf("total del tutor incorrecto: %v", total)
	}
	solicitarI17(t, e, tutor, "GET", "/api/minutos/total?tutor_id=otro", "", 403)
	solicitarI17(t, e, escolar, "GET", "/api/minutos/total", "", 403)
	solicitarI17(t, e, coordinador, "GET", "/api/minutos/total", "", 400)
	solicitarI17(t, e, coordinador, "GET", "/api/minutos/total?tutor_id="+e.tutorID, "", 200)

	// Auditoria: proponer, aprobar, disputar, resolver.
	solicitarI17(t, e, tutor, "GET", "/api/minutos/sesion-i17/auditoria", "", 403)
	auditoria := solicitarI17(t, e, coordinador, "GET", "/api/minutos/sesion-i17/auditoria", "", 200)
	entradas, _ := auditoria["auditoria"].([]any)
	if len(entradas) != 4 {
		t.Fatalf("esperadas 4 entradas de auditoria: %v", auditoria)
	}
	ultima, _ := entradas[3].(map[string]any)
	if ultima["accion"] != "resolver" || ultima["motivo"] != "Revisado con ambos" || ultima["valor_nuevo"] != float64(30) {
		t.Fatalf("entrada de resolucion incorrecta: %v", ultima)
	}
}

func TestI17CancelacionHTTP(t *testing.T) {
	e := prepararEntornoI17(t)
	tutor := iniciarSesionI05(t, e.manejador, "tutor.prueba@uni.edu.pe")
	datos := solicitarI17(t, e, tutor, "POST", "/api/sesion/sesion-i17-cancelada/minutos", `{"minutos":60}`, 409)
	if datos["codigo"] != "sesion_cancelada" {
		t.Fatalf("una reserva cancelada debe rechazarse con sesion_cancelada: %v", datos)
	}
	total := solicitarI17(t, e, tutor, "GET", "/api/minutos/total", "", 200)
	if total["minutos_verificados"] != float64(0) {
		t.Fatalf("una cancelacion suma cero: %v", total)
	}
}
