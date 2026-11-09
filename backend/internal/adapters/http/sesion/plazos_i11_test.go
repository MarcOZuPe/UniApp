package sesion

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	modeloSesion "uniapp/internal/application/sesion"
)

func TestI11HTTPRechazaAccionesFueraDePlazo(t *testing.T) {
	finSesion := inicioSesionI05.Add(modeloSesion.DuracionSesion)
	momentos := map[string]time.Time{
		"antes de la sesion":       inicioSesionI05.Add(-modeloSesion.AnticipacionDeclaracion - time.Minute),
		"despues de todo el plazo": finSesion.Add(modeloSesion.PlazoResultado + time.Minute),
	}

	for nombre, momento := range momentos {
		t.Run(nombre, func(t *testing.T) {
			entorno := prepararEntornoI05EnMomento(t, momento)
			tutor := iniciarSesionI05(t, entorno.manejador, "tutor.prueba@uni.edu.pe")

			for _, ruta := range []string{"/asistencias", "/ausencias"} {
				respuesta := solicitarI05(opcionesSolicitudI05{
					manejador: entorno.manejador, metodo: http.MethodPost,
					ruta: "/api/sesion/sesion-i05" + ruta, cookie: tutor, origen: origenI05,
				})
				comprobarFueraDePlazo(t, ruta, respuesta.Code, respuesta.Body.Bytes())
			}

			respuesta := solicitarResultadoI05(t, entorno.manejador, tutor, `{"resumen":"Resumen"}`)
			comprobarFueraDePlazo(t, "/resultado", respuesta.Code, respuesta.Body.Bytes())
		})
	}
}

func TestI11HTTPConsultaPublicaLosPlazos(t *testing.T) {
	entorno := prepararEntornoI05(t)
	tutor := iniciarSesionI05(t, entorno.manejador, "tutor.prueba@uni.edu.pe")

	respuesta := solicitarI05(opcionesSolicitudI05{
		manejador: entorno.manejador, metodo: http.MethodGet,
		ruta: "/api/sesion/sesion-i05", cookie: tutor,
	})
	if respuesta.Code != http.StatusOK {
		t.Fatalf("esperado 200, obtenido %d: %s", respuesta.Code, respuesta.Body.String())
	}

	var datos struct {
		Plazos struct {
			Declaracion struct{ Desde, Hasta time.Time } `json:"declaracion"`
			Resultado   struct{ Desde, Hasta time.Time } `json:"resultado"`
		} `json:"plazos"`
	}
	if err := json.Unmarshal(respuesta.Body.Bytes(), &datos); err != nil {
		t.Fatal(err)
	}

	finSesion := inicioSesionI05.Add(modeloSesion.DuracionSesion)
	if !datos.Plazos.Declaracion.Desde.Equal(inicioSesionI05.Add(-modeloSesion.AnticipacionDeclaracion)) ||
		!datos.Plazos.Declaracion.Hasta.Equal(finSesion.Add(modeloSesion.PlazoDeclaracion)) ||
		!datos.Plazos.Resultado.Desde.Equal(inicioSesionI05) ||
		!datos.Plazos.Resultado.Hasta.Equal(finSesion.Add(modeloSesion.PlazoResultado)) {
		t.Fatalf("plazos publicados incorrectos: %+v", datos.Plazos)
	}
}

func comprobarFueraDePlazo(t *testing.T, ruta string, estado int, cuerpo []byte) {
	t.Helper()
	var error struct {
		Codigo string `json:"codigo"`
	}
	_ = json.Unmarshal(cuerpo, &error)
	if estado != http.StatusUnprocessableEntity || error.Codigo != "fuera_de_plazo" {
		t.Fatalf("%s: esperado 422 fuera_de_plazo, obtenido %d: %s", ruta, estado, cuerpo)
	}
}
