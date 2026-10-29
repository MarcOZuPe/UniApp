package sesion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestI11AusenciaHTTP(t *testing.T) {
	e := prepararEntornoI05(t)
	tutor := iniciarSesionI05(t, e.handler, "tutor.prueba@uni.edu.pe")
	escolar := iniciarSesionI05(t, e.handler, "escolar.prueba@example.test")
	ajeno := iniciarSesionI05(t, e.handler, "apoderado.prueba@example.test")
	ruta := "/api/sesion/sesion-i05"
	comprobar := func(metodo, url string, cookie *http.Cookie, origen string, estado int) *httptest.ResponseRecorder {
		t.Helper()
		r := solicitarI05(e.handler, metodo, url, cookie, origen)
		if r.Code != estado {
			t.Fatalf("%s %s: esperado %d recibido %d: %s", metodo, url, estado, r.Code, r.Body.String())
		}
		return r
	}
	comprobar("POST", ruta+"/ausencias", nil, origenI05, 401)
	comprobar("POST", ruta+"/ausencias", ajeno, origenI05, 403)
	comprobar("POST", ruta+"/ausencias", tutor, "https://ajeno.test", 403)
	comprobar("POST", "/api/sesion/inexistente/ausencias", tutor, origenI05, 404)
	// La consulta no registra declaraciones.
	r := comprobar("GET", ruta, tutor, "", 200)
	var datos struct {
		Ausencias []struct {
			UsuarioID   string    `json:"usuario_id"`
			Rol         string    `json:"rol"`
			ReportadaEn time.Time `json:"reportada_en"`
		} `json:"ausencias"`
		Asistencias []json.RawMessage `json:"asistencias"`
	}
	if err := json.Unmarshal(r.Body.Bytes(), &datos); err != nil {
		t.Fatal(err)
	}
	if len(datos.Ausencias) != 0 || len(datos.Asistencias) != 0 {
		t.Fatal("consultar registro declaraciones")
	}
	// No se puede atribuir la ausencia a otro usuario mediante el cuerpo.
	p := httptest.NewRequest("POST", ruta+"/ausencias", strings.NewReader(`{"usuario_id":"`+e.escolarID+`"}`))
	p.AddCookie(tutor)
	p.Header.Set("Origin", origenI05)
	respuesta := httptest.NewRecorder()
	e.handler.ServeHTTP(respuesta, p)
	if respuesta.Code != 400 {
		t.Fatalf("suplantacion: %d", respuesta.Code)
	}
	comprobar("POST", ruta+"/ausencias", tutor, origenI05, 201)
	comprobar("POST", ruta+"/ausencias", tutor, origenI05, 409)
	comprobar("POST", ruta+"/asistencias", tutor, origenI05, 409)
	comprobar("POST", ruta+"/asistencias", escolar, origenI05, 201)
	comprobar("POST", ruta+"/ausencias", escolar, origenI05, 409)
	r = comprobar("GET", ruta, escolar, "", 200)
	if err := json.Unmarshal(r.Body.Bytes(), &datos); err != nil {
		t.Fatal(err)
	}
	if len(datos.Ausencias) != 1 || len(datos.Asistencias) != 1 {
		t.Fatal("declaraciones incorrectas")
	}
	a := datos.Ausencias[0]
	if a.UsuarioID != e.tutorID || a.Rol != "tutor" || a.ReportadaEn.IsZero() {
		t.Fatal("ausencia incorrecta")
	}
}
