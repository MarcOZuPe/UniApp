package reserva

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	ctrlAcceso "uniapp/internal/controllers/acceso"
	modeloAcceso "uniapp/internal/models/acceso"
	modelo "uniapp/internal/models/reserva"
)

const (
	origen     = "http://localhost:5173"
	clave      = "Clave-Segura-2026!"
	cookieName = "uniapp_sesion"
	cuerpoOK   = `{"slot_id":"slot-i04-01","tema":"Repaso de factorización","objetivo":"Resolver ejercicios de factorización de trinomios sin ayuda.","grado":"5to de secundaria","adjuntos":[{"nombre":"practica.pdf","tipo":"application/pdf","tamano_bytes":1024}]}`
)

type entorno struct {
	manejador http.Handler
	escolar   *http.Cookie
	tutor     *http.Cookie
}

type autorizacionFija bool

func (a autorizacionFija) AutorizacionVigente(context.Context, string) (bool, error) {
	return bool(a), nil
}

type autorizacionVariable struct{ permitido bool }

func (a *autorizacionVariable) AutorizacionVigente(context.Context, string) (bool, error) {
	return a.permitido, nil
}

func nuevoEntorno(t *testing.T) *entorno {
	t.Helper()
	return nuevoEntornoConAutorizacion(t, autorizacionFija(true))
}

func nuevoEntornoConAutorizacion(t *testing.T, autorizacion modelo.PuertoAutorizacion) *entorno {
	t.Helper()
	hasher := modeloAcceso.PBKDF2Hasher{Iteraciones: 1000}
	servicioAcceso := modeloAcceso.NuevoServicio(modeloAcceso.Dependencias{
		Repositorio: modeloAcceso.NuevoRepositorioMemoria(),
		Fabrica:     modeloAcceso.NuevaFabrica(hasher, "uni.edu.pe"),
		Hasher:      hasher,
		Bus:         modeloAcceso.NuevoBus(),
	})
	if err := modeloAcceso.Sembrar(context.Background(), servicioAcceso, modeloAcceso.OpcionesSiembra{DominioTutor: "uni.edu.pe", Clave: clave}); err != nil {
		t.Fatal(err)
	}
	controladorAcceso := ctrlAcceso.Nuevo(servicioAcceso, false)

	ahora := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	servicioReserva := modelo.NuevoServicio(modelo.Dependencias{
		Repositorio:  modelo.NuevoRepositorioMemoria(),
		Oferta:       modelo.NuevaOfertaSintetica(ahora),
		Autorizacion: autorizacion,
	})
	servicioReserva.FijarReloj(func() time.Time { return ahora })

	manejadorReserva := Nuevo(servicioReserva, controladorAcceso.Autenticar()).Rutas()
	raiz := http.NewServeMux()
	// Igual que cmd/server: el prefijo exacto evita el redirect 301 de POST /api/reserva.
	raiz.Handle("/api/reserva", manejadorReserva)
	raiz.Handle("/api/reserva/", manejadorReserva)
	raiz.Handle("/", controladorAcceso.Rutas())
	manejador := ctrlAcceso.Encadenar(raiz, ctrlAcceso.OrigenPermitido(origen))

	return &entorno{
		manejador: manejador,
		escolar:   ingresar(t, manejador, "escolar.prueba@example.test"),
		tutor:     ingresar(t, manejador, "tutor.prueba@uni.edu.pe"),
	}
}

func llamar(manejador http.Handler, metodo, ruta, cuerpo string, cookie *http.Cookie) *httptest.ResponseRecorder {
	peticion := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	peticion.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		peticion.AddCookie(cookie)
	}
	grabador := httptest.NewRecorder()
	manejador.ServeHTTP(grabador, peticion)
	return grabador
}

func ingresar(t *testing.T, manejador http.Handler, email string) *http.Cookie {
	t.Helper()
	respuesta := llamar(manejador, "POST", "/api/auth/ingreso", `{"email":"`+email+`","clave":"`+clave+`"}`, nil)
	if respuesta.Code != http.StatusOK {
		t.Fatalf("ingreso %s: %d %s", email, respuesta.Code, respuesta.Body)
	}
	for _, cookie := range respuesta.Result().Cookies() {
		if cookie.Name == cookieName {
			return cookie
		}
	}
	t.Fatal("sin cookie de sesión")
	return nil
}

func decodificar(t *testing.T, respuesta *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var cuerpo map[string]any
	if err := json.Unmarshal(respuesta.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("JSON inválido (%d): %s", respuesta.Code, respuesta.Body)
	}
	return cuerpo
}

func TestReservaBaseDeExtremoAExtremo(t *testing.T) {
	entorno := nuevoEntorno(t)

	listado := llamar(entorno.manejador, "GET", "/api/reserva/slots", "", entorno.escolar)
	if listado.Code != http.StatusOK {
		t.Fatalf("slots: %d %s", listado.Code, listado.Body)
	}
	var slots struct {
		Slots []struct {
			ID         string `json:"id"`
			Disponible bool   `json:"disponible"`
		} `json:"slots"`
	}
	if err := json.Unmarshal(listado.Body.Bytes(), &slots); err != nil || len(slots.Slots) == 0 {
		t.Fatalf("lista de horarios: %v %s", err, listado.Body)
	}

	creada := llamar(entorno.manejador, "POST", "/api/reserva", cuerpoOK, entorno.escolar)
	if creada.Code != http.StatusCreated {
		t.Fatalf("crear: %d %s", creada.Code, creada.Body)
	}
	reserva := decodificar(t, creada)
	if reserva["estado"] != "confirmada" || reserva["curso"] != "Álgebra" {
		t.Fatalf("reserva confirmada con el horario elegido: %v", reserva)
	}
	if reserva["inicio_lima"] == "" || reserva["tema"] == "" {
		t.Fatalf("la respuesta presenta la hora de Lima: %v", reserva)
	}
	adjuntos, ok := reserva["adjuntos"].([]any)
	if !ok || len(adjuntos) != 1 {
		t.Fatalf("adjunto opcional guardado: %v", reserva["adjuntos"])
	}

	consultada := llamar(entorno.manejador, "GET", "/api/reserva/"+reserva["id"].(string), "", entorno.escolar)
	if consultada.Code != http.StatusOK {
		t.Fatalf("consultar: %d %s", consultada.Code, consultada.Body)
	}
}

func TestRevocarAutorizacionNoAfectaReservasConfirmadas(t *testing.T) {
	autorizacion := &autorizacionVariable{permitido: true}
	entorno := nuevoEntornoConAutorizacion(t, autorizacion)

	creada := llamar(entorno.manejador, "POST", "/api/reserva", cuerpoOK, entorno.escolar)
	if creada.Code != http.StatusCreated {
		t.Fatalf("reserva inicial: %d %s", creada.Code, creada.Body)
	}
	id := decodificar(t, creada)["id"].(string)

	autorizacion.permitido = false
	nueva := llamar(entorno.manejador, "POST", "/api/reserva", cuerpoOK, entorno.escolar)
	if nueva.Code != http.StatusForbidden || decodificar(t, nueva)["codigo"] != "sin_autorizacion" {
		t.Fatalf("sin autorización no se reserva: %d %s", nueva.Code, nueva.Body)
	}

	existente := llamar(entorno.manejador, "GET", "/api/reserva/"+id, "", entorno.escolar)
	if existente.Code != http.StatusOK {
		t.Fatalf("la reserva ya confirmada sigue vigente: %d %s", existente.Code, existente.Body)
	}
}

func TestReservaRequiereSesionDeEscolar(t *testing.T) {
	entorno := nuevoEntorno(t)
	if respuesta := llamar(entorno.manejador, "POST", "/api/reserva", cuerpoOK, nil); respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", respuesta.Code)
	}
	respuesta := llamar(entorno.manejador, "POST", "/api/reserva", cuerpoOK, entorno.tutor)
	if respuesta.Code != http.StatusForbidden || decodificar(t, respuesta)["codigo"] != "solo_escolar" {
		t.Fatalf("un tutor no reserva: %d %s", respuesta.Code, respuesta.Body)
	}
	if respuesta := llamar(entorno.manejador, "GET", "/api/reserva/slots", "", nil); respuesta.Code != http.StatusUnauthorized {
		t.Fatalf("slots sin sesión: %d", respuesta.Code)
	}
}

func TestReservaRechazaDatosYReglas(t *testing.T) {
	entorno := nuevoEntorno(t)
	casos := []struct {
		nombre string
		cuerpo string
		estado int
		codigo string
		campo  string
	}{
		{"tema corto", `{"slot_id":"slot-i04-01","tema":"corto","objetivo":"Resolver ejercicios de factorización de trinomios sin ayuda.","grado":"5to"}`, http.StatusUnprocessableEntity, "validacion", "tema"},
		{"sin grado", `{"slot_id":"slot-i04-01","tema":"Repaso de factorización","objetivo":"Resolver ejercicios de factorización de trinomios sin ayuda.","grado":""}`, http.StatusUnprocessableEntity, "validacion", "grado"},
		{"menos de 48 horas", `{"slot_id":"slot-i04-cercano","tema":"Repaso de factorización","objetivo":"Resolver ejercicios de factorización de trinomios sin ayuda.","grado":"5to"}`, http.StatusUnprocessableEntity, "anticipacion_insuficiente", "slot_id"},
		{"fuera del horizonte", `{"slot_id":"slot-i04-lejano","tema":"Repaso de factorización","objetivo":"Resolver ejercicios de factorización de trinomios sin ayuda.","grado":"5to"}`, http.StatusUnprocessableEntity, "horizonte_excedido", "slot_id"},
		{"slot inexistente", `{"slot_id":"slot-fantasma","tema":"Repaso de factorización","objetivo":"Resolver ejercicios de factorización de trinomios sin ayuda.","grado":"5to"}`, http.StatusNotFound, "slot_inexistente", ""},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			respuesta := llamar(entorno.manejador, "POST", "/api/reserva", caso.cuerpo, entorno.escolar)
			if respuesta.Code != caso.estado {
				t.Fatalf("estado %d, esperado %d: %s", respuesta.Code, caso.estado, respuesta.Body)
			}
			cuerpo := decodificar(t, respuesta)
			if cuerpo["codigo"] != caso.codigo {
				t.Fatalf("código %v, esperado %s", cuerpo["codigo"], caso.codigo)
			}
			if caso.campo != "" {
				campos, _ := cuerpo["campos"].(map[string]any)
				if campos[caso.campo] == nil {
					t.Fatalf("falta el campo %q en %v", caso.campo, cuerpo["campos"])
				}
			}
		})
	}
}

func TestReservaNoDuplicaSlotYSoloLaVeSuDueno(t *testing.T) {
	entorno := nuevoEntorno(t)
	primera := llamar(entorno.manejador, "POST", "/api/reserva", cuerpoOK, entorno.escolar)
	if primera.Code != http.StatusCreated {
		t.Fatalf("primera reserva: %d %s", primera.Code, primera.Body)
	}
	repetida := llamar(entorno.manejador, "POST", "/api/reserva", cuerpoOK, entorno.escolar)
	if repetida.Code != http.StatusConflict || decodificar(t, repetida)["codigo"] != "slot_ocupado" {
		t.Fatalf("segunda reserva del mismo slot: %d %s", repetida.Code, repetida.Body)
	}
	id := decodificar(t, primera)["id"].(string)
	if respuesta := llamar(entorno.manejador, "GET", "/api/reserva/"+id, "", entorno.tutor); respuesta.Code != http.StatusForbidden {
		t.Fatalf("el tutor no ve la reserva ajena: %d", respuesta.Code)
	}
	if respuesta := llamar(entorno.manejador, "GET", "/api/reserva/00000000-0000-0000-0000-000000000000", "", entorno.escolar); respuesta.Code != http.StatusNotFound {
		t.Fatalf("reserva inexistente: %d", respuesta.Code)
	}
}
