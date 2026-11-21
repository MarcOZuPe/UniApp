package sesion

import (
	"errors"
	"testing"
	"time"
)

func TestI11VentanasSeCalculanDesdeLaSesion(t *testing.T) {
	s := nuevaSesionPrueba(t)

	declaracion := s.VentanaDeclaracion()
	if !declaracion.Desde().Equal(s.Inicio().Add(-AnticipacionDeclaracion)) ||
		!declaracion.Hasta().Equal(s.Fin().Add(PlazoDeclaracion)) {
		t.Fatalf("ventana de declaracion incorrecta: %v - %v", declaracion.Desde(), declaracion.Hasta())
	}

	resultado := s.VentanaResultado()
	if !resultado.Desde().Equal(s.Inicio()) || !resultado.Hasta().Equal(s.Fin().Add(PlazoResultado)) {
		t.Fatalf("ventana de resultado incorrecta: %v - %v", resultado.Desde(), resultado.Hasta())
	}
}

func TestI11DeclaracionesRespetanLaVentana(t *testing.T) {
	base := nuevaSesionPrueba(t)
	ventana := base.VentanaDeclaracion()

	casos := []struct {
		nombre   string
		momento  time.Time
		esperado error
	}{
		{"antes de la ventana", ventana.Desde().Add(-time.Second), ErrFueraDePlazo},
		{"limite inicial", ventana.Desde(), nil},
		{"durante la sesion", base.Inicio().Add(30 * time.Minute), nil},
		{"limite final", ventana.Hasta(), nil},
		{"despues de la ventana", ventana.Hasta().Add(time.Second), ErrFueraDePlazo},
	}

	comandos := map[string]func(*Sesion, time.Time) error{
		"asistencia": func(s *Sesion, momento time.Time) error { return s.DeclararAsistencia(s.EscolarID(), momento) },
		"ausencia":   func(s *Sesion, momento time.Time) error { return s.ReportarAusencia(s.EscolarID(), momento) },
	}

	for nombreComando, comando := range comandos {
		for _, caso := range casos {
			t.Run(nombreComando+"/"+caso.nombre, func(t *testing.T) {
				s := nuevaSesionPrueba(t)
				if err := comando(s, caso.momento); !errors.Is(err, caso.esperado) {
					t.Fatalf("esperado %v, obtenido %v", caso.esperado, err)
				}
			})
		}
	}
}

func TestI11ResultadoRespetaLaVentana(t *testing.T) {
	base := nuevaSesionPrueba(t)
	ventana := base.VentanaResultado()

	casos := []struct {
		nombre   string
		momento  time.Time
		esperado error
	}{
		{"antes del inicio", ventana.Desde().Add(-time.Second), ErrFueraDePlazo},
		{"al iniciar", ventana.Desde(), nil},
		{"limite final", ventana.Hasta(), nil},
		{"despues del plazo", ventana.Hasta().Add(time.Second), ErrFueraDePlazo},
	}

	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			s := nuevaSesionPrueba(t)
			err := s.RegistrarResultado(s.TutorID(), "Resumen", caso.momento)
			if !errors.Is(err, caso.esperado) {
				t.Fatalf("esperado %v, obtenido %v", caso.esperado, err)
			}
		})
	}
}

func TestI11AutorizacionPrevaleceSobrePlazo(t *testing.T) {
	s := nuevaSesionPrueba(t)
	fuera := s.VentanaResultado().Hasta().Add(time.Hour)

	if err := s.DeclararAsistencia("ajeno", fuera); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("asistencia de ajeno: %v", err)
	}
	if err := s.ReportarAusencia("ajeno", fuera); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("ausencia de ajeno: %v", err)
	}
	if err := s.RegistrarResultado(s.EscolarID(), "Resumen", fuera); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("resultado de escolar: %v", err)
	}
}

func TestI11RestaurarNoAplicaPoliticaTemporal(t *testing.T) {
	s := nuevaSesionPrueba(t)
	historico := s.Fin().Add(30 * 24 * time.Hour)

	if err := s.RestaurarAsistencia(s.TutorID(), historico); err != nil {
		t.Fatalf("asistencia historica: %v", err)
	}
	if err := s.RestaurarAusencia(s.EscolarID(), historico); err != nil {
		t.Fatalf("ausencia historica: %v", err)
	}
	if err := s.RestaurarResultado(s.TutorID(), "Resumen", historico); err != nil {
		t.Fatalf("resultado historico: %v", err)
	}
}

func TestI11RestaurarMantieneInvariantes(t *testing.T) {
	s := nuevaSesionPrueba(t)
	momento := s.Fin()

	if err := s.RestaurarAsistencia("ajeno", momento); !errors.Is(err, ErrNoAutorizado) {
		t.Fatalf("ajeno: %v", err)
	}
	if err := s.RestaurarAsistencia(s.TutorID(), momento); err != nil {
		t.Fatal(err)
	}
	if err := s.RestaurarAsistencia(s.TutorID(), momento); !errors.Is(err, ErrAsistenciaDuplicada) {
		t.Fatalf("duplicado: %v", err)
	}
	if err := s.RestaurarAusencia(s.TutorID(), momento); !errors.Is(err, ErrDeclaracionContradictoria) {
		t.Fatalf("contradiccion: %v", err)
	}
	if err := s.RestaurarResultado(s.TutorID(), "   ", momento); !errors.Is(err, ErrResultadoInvalido) {
		t.Fatalf("resumen vacio: %v", err)
	}
}
