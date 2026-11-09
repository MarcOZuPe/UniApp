package sesion

import (
	"errors"
	"testing"
	"time"
)

func TestI11AusenciaPropia(t *testing.T) {
	for _, caso := range []struct {
		usuario string
		rol     Rol
	}{
		{"tutor-001", RolTutor},
		{"escolar-001", RolEscolar},
	} {
		t.Run(string(caso.rol), func(t *testing.T) {
			s := nuevaSesionPrueba(t)
			ahora := s.Fin().In(time.FixedZone("Lima", -5*60*60))
			if err := s.ReportarAusencia(caso.usuario, ahora); err != nil {
				t.Fatal(err)
			}
			ausencias := s.Ausencias()
			if len(ausencias) != 1 {
				t.Fatal("se esperaba una ausencia")
			}
			a := ausencias[0]
			if a.UsuarioID() != caso.usuario || a.Rol() != caso.rol ||
				!a.ReportadaEn().Equal(ahora) || a.ReportadaEn().Location() != time.UTC {
				t.Fatalf("ausencia incorrecta: %+v", a)
			}
			ausencias[0] = Ausencia{}
			if s.Ausencias()[0].UsuarioID() != caso.usuario {
				t.Fatal("la consulta modifico el estado interno")
			}
			copia := s.Clonar()
			copia.ausencias[0] = Ausencia{}
			if s.Ausencias()[0].UsuarioID() != caso.usuario {
				t.Fatal("el clon comparte las ausencias")
			}
		})
	}
}

func TestI11AusenciaRechazaAjenoYFechaInvalida(t *testing.T) {
	s := nuevaSesionPrueba(t)
	if err := s.ReportarAusencia("ajeno", s.Fin()); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("se esperaba no autorizado: %v", err)
	}
	if err := s.ReportarAusencia(s.TutorID(), time.Time{}); !errors.Is(err, ErrSesionInvalida) {
		t.Fatalf("se esperaba fecha invalida: %v", err)
	}
	if len(s.Ausencias()) != 0 {
		t.Fatal("un rechazo no debe registrar ausencias")
	}
}

func TestI11AusenciaDuplicadaYContradicciones(t *testing.T) {
	s := nuevaSesionPrueba(t)
	if err := s.ReportarAusencia(s.TutorID(), s.Fin()); err != nil {
		t.Fatal(err)
	}
	if err := s.ReportarAusencia(s.TutorID(), s.Fin()); !errors.Is(err, ErrAusenciaDuplicada) {
		t.Fatalf("se esperaba duplicado: %v", err)
	}
	if err := s.DeclararAsistencia(s.TutorID(), s.Fin()); !errors.Is(err, ErrDeclaracionContradictoria) {
		t.Fatalf("se esperaba contradiccion tras ausencia: %v", err)
	}
	if err := s.DeclararAsistencia(s.EscolarID(), s.Fin()); err != nil {
		t.Fatal(err)
	}
	if err := s.ReportarAusencia(s.EscolarID(), s.Fin()); !errors.Is(err, ErrDeclaracionContradictoria) {
		t.Fatalf("se esperaba contradiccion tras asistencia: %v", err)
	}
	if len(s.Ausencias()) != 1 || len(s.Asistencias()) != 1 {
		t.Fatal("los rechazos no deben modificar declaraciones")
	}
}
