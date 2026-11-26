package acceso

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// F-12: el mapeador HTTP responde 500 para errores no reconocidos pero no deja
// rastro del error original, lo que impide diagnosticar fallos reales.
func TestF12MapeadorRegistraElErrorInterno(t *testing.T) {
	var registro bytes.Buffer
	anterior := log.Writer()
	log.SetOutput(&registro)
	t.Cleanup(func() { log.SetOutput(anterior) })

	mapeador := NuevoMapeadorDeErrores()
	respuesta := httptest.NewRecorder()
	mapeador.Escribir(respuesta, errors.New("fallo de PostgreSQL 08006 conexión perdida"))

	if respuesta.Code != http.StatusInternalServerError {
		t.Fatalf("F-12: se esperaba 500, se obtuvo %d", respuesta.Code)
	}
	if !strings.Contains(registro.String(), "08006") {
		t.Fatalf(
			"F-12: el error interno no quedó registrado en el log; log=%q",
			registro.String(),
		)
	}
}
