package sesion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestI11ResultadoHTTP(t *testing.T) {
	e := prepararEntornoI05(t)
	tutor := iniciarSesionI05(t, e.manejador, "tutor.prueba@uni.edu.pe")
	escolar := iniciarSesionI05(t, e.manejador, "escolar.prueba@example.test")
	ajeno := iniciarSesionI05(t, e.manejador, "apoderado.prueba@example.test")
	ruta := "/api/sesion/sesion-i05"
	enviar := func(cookie *http.Cookie, cuerpo string, estado int) *httptest.ResponseRecorder {
		t.Helper()
		p := httptest.NewRequest("POST", ruta+"/resultado", strings.NewReader(cuerpo))
		p.Header.Set("Origin", origenI05)
		p.Header.Set("Content-Type", "application/json")
		if cookie != nil {
			p.AddCookie(cookie)
		}
		r := httptest.NewRecorder()
		e.manejador.ServeHTTP(r, p)
		if r.Code != estado {
			t.Fatalf("esperado %d recibido %d: %s", estado, r.Code, r.Body.String())
		}
		return r
	}
	valido := `{"resumen":"  Ecuaciones resueltas.  "}`
	enviar(nil, valido, 401)
	enviar(escolar, valido, 403)
	enviar(ajeno, valido, 403)
	for _, cuerpo := range []string{"", "{", `{"resumen":"Texto","usuario_id":"otro"}`, `{"resumen":"Texto","registrado_en":"2020-01-01"}`, valido + valido, `{"resumen":5}`, `{"resumen":"` + strings.Repeat("a", 17000) + `"}`} {
		enviar(tutor, cuerpo, 400)
	}
	for _, cuerpo := range []string{`{}`, `{"resumen":"  "}`, `{"resumen":"` + strings.Repeat("a", 2001) + `"}`} {
		enviar(tutor, cuerpo, 422)
	}
	var datos struct {
		Resultado *struct {
			UsuarioID    string    `json:"usuario_id"`
			Resumen      string    `json:"resumen"`
			RegistradoEn time.Time `json:"registrado_en"`
		} `json:"resultado"`
		Asistencias []json.RawMessage `json:"asistencias"`
		Ausencias   []json.RawMessage `json:"ausencias"`
	}
	consulta := solicitarI05(opcionesSolicitudI05{manejador: e.manejador, metodo: "GET", ruta: ruta, cookie: tutor, origen: ""})
	if err := json.Unmarshal(consulta.Body.Bytes(), &datos); err != nil {
		t.Fatal(err)
	}
	if datos.Resultado != nil {
		t.Fatal("rechazos o consulta crearon resultado")
	}
	respuesta := enviar(tutor, valido, 201)
	var creado struct {
		UsuarioID string `json:"usuario_id"`
		Resumen   string `json:"resumen"`
	}
	if err := json.Unmarshal(respuesta.Body.Bytes(), &creado); err != nil {
		t.Fatal(err)
	}
	if creado.UsuarioID != e.tutorID || creado.Resumen != "Ecuaciones resueltas." {
		t.Fatal("respuesta incorrecta")
	}
	enviar(tutor, valido, 409)
	for _, cookie := range []*http.Cookie{tutor, escolar} {
		consulta = solicitarI05(opcionesSolicitudI05{manejador: e.manejador, metodo: "GET", ruta: ruta, cookie: cookie, origen: ""})
		if consulta.Code != 200 {
			t.Fatalf("consulta: %d", consulta.Code)
		}
		if err := json.Unmarshal(consulta.Body.Bytes(), &datos); err != nil {
			t.Fatal(err)
		}
		if datos.Resultado == nil || datos.Resultado.UsuarioID != e.tutorID || datos.Resultado.Resumen != "Ecuaciones resueltas." || datos.Resultado.RegistradoEn.IsZero() {
			t.Fatal("resultado no recuperado")
		}
		if len(datos.Asistencias) != 0 || len(datos.Ausencias) != 0 {
			t.Fatal("el resultado modifico declaraciones")
		}
	}
	if r := solicitarI05(opcionesSolicitudI05{manejador: e.manejador, metodo: "GET", ruta: ruta, cookie: ajeno, origen: ""}); r.Code != 403 {
		t.Fatalf("consulta ajena: %d", r.Code)
	}
}
