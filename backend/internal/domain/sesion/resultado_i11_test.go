package sesion

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestI11ResultadoDelTutorIndependienteDeAsistencia(t *testing.T) {
	s := nuevaSesionPrueba(t)
	if s.Resultado() != nil {
		t.Fatal("una sesion nueva no debe tener resultado")
	}
	ahora := s.Fin().In(time.FixedZone("Lima", -5*60*60))
	if err := s.RegistrarResultado(s.TutorID(), "  Se practicaron ecuaciones.  ", ahora); err != nil {
		t.Fatal(err)
	}
	r := s.Resultado()
	if r.UsuarioID() != s.TutorID() || r.Resumen() != "Se practicaron ecuaciones." || !r.RegistradoEn().Equal(ahora) || r.RegistradoEn().Location() != time.UTC {
		t.Fatalf("resultado incorrecto: %+v", r)
	}
	r.resumen = "alterado"
	copia := s.Clonar()
	copia.resultado.resumen = "alterado"
	if s.Resultado().Resumen() != "Se practicaron ecuaciones." {
		t.Fatal("se expuso estado mutable")
	}
	if err := s.RegistrarResultado(s.TutorID(), "Otro resumen", ahora); !errors.Is(err, ErrResultadoDuplicado) {
		t.Fatalf("duplicado: %v", err)
	}
	if len(s.Asistencias()) != 0 || len(s.Ausencias()) != 0 {
		t.Fatal("el resultado no debe acreditar asistencia ni ausencia")
	}
	if err := s.DeclararAsistencia(s.EscolarID(), ahora); err != nil {
		t.Fatal(err)
	}
	if s.Resultado().Resumen() != "Se practicaron ecuaciones." {
		t.Fatal("la asistencia altero el resultado")
	}
}

func TestI11ResultadoAutorizacionYValidacion(t *testing.T) {
	s := nuevaSesionPrueba(t)
	for _, usuario := range []string{s.EscolarID(), "ajeno"} {
		if err := s.RegistrarResultado(usuario, "Resumen", s.Fin()); !errors.Is(err, ErrNoAutorizado) {
			t.Fatalf("autorizacion de %s: %v", usuario, err)
		}
	}
	for _, texto := range []string{"", " \n\t", strings.Repeat("ñ", MaxCaracteresResultado+1)} {
		if err := s.RegistrarResultado(s.TutorID(), texto, s.Fin()); !errors.Is(err, ErrResultadoInvalido) {
			t.Fatalf("validacion: %v", err)
		}
	}
	if err := s.RegistrarResultado(s.TutorID(), "Resumen", time.Time{}); !errors.Is(err, ErrSesionInvalida) {
		t.Fatalf("fecha: %v", err)
	}
	if s.Resultado() != nil {
		t.Fatal("los rechazos alteraron el resultado")
	}
	if err := s.RegistrarResultado(s.TutorID(), strings.Repeat("ñ", MaxCaracteresResultado), s.Fin()); err != nil {
		t.Fatalf("limite de caracteres Unicode: %v", err)
	}
}
