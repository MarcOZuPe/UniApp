package sesion

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

var claveI18 = []byte("clave-de-prueba-i18-suficiente")

func seudonimizadorI18(t *testing.T) Seudonimizador {
	t.Helper()
	s, err := NuevoSeudonimizador(claveI18)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func filtroI18(t *testing.T, tutorID string) FiltroReporte {
	t.Helper()
	desde := time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)
	f, err := NuevoFiltroReporte(desde, desde.AddDate(0, 0, 30), tutorID)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func sesionI18(t *testing.T, id, tutor, escolar string, inicio time.Time) *Sesion {
	t.Helper()
	s, err := NuevaSesion(id, "reserva-"+id, tutor, escolar, inicio, inicio.Add(DuracionSesion))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// conMinutos deja la sesion con asistencia completa y minutos propuestos.
func conMinutos(t *testing.T, s *Sesion, minutos int) *RegistroMinutos {
	t.Helper()
	for _, id := range []string{s.TutorID(), s.EscolarID()} {
		if err := s.DeclararAsistencia(id, s.Inicio()); err != nil {
			t.Fatal(err)
		}
	}
	r, err := ProponerMinutos(s, s.TutorID(), minutos, s.Fin().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestI18FiltroValidaPeriodo(t *testing.T) {
	base := time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)
	casos := map[string][2]time.Time{
		"desde cero":     {{}, base},
		"hasta antes":    {base, base.Add(-time.Hour)},
		"iguales":        {base, base},
		"mas de 31 dias": {base, base.AddDate(0, 0, MaxDiasReporte).Add(time.Second)},
	}
	for nombre, caso := range casos {
		if _, err := NuevoFiltroReporte(caso[0], caso[1], ""); !errors.Is(err, ErrFiltroReporteInvalido) {
			t.Errorf("%s: esperaba ErrFiltroReporteInvalido, obtuve %v", nombre, err)
		}
	}
	if _, err := NuevoFiltroReporte(base, base.AddDate(0, 0, MaxDiasReporte), " tutor "); err != nil {
		t.Fatalf("31 dias exactos debe aceptarse: %v", err)
	}
}

func TestI18FiltroIncluyeInicioSemiabierto(t *testing.T) {
	f := filtroI18(t, "tutor-a")
	dentro := sesionI18(t, "s1", "tutor-a", "esc", f.Desde())
	enElLimite := sesionI18(t, "s2", "tutor-a", "esc", f.Hasta())
	otroTutor := sesionI18(t, "s3", "tutor-b", "esc", f.Desde())
	if !f.Incluye(dentro) || f.Incluye(enElLimite) || f.Incluye(otroTutor) || f.Incluye(nil) {
		t.Fatal("el filtro debe ser [desde, hasta) y respetar el tutor")
	}
}

func TestI18FormatoSoloCSVoPDF(t *testing.T) {
	for _, v := range []string{"csv", " PDF "} {
		if _, err := ValidarFormatoReporte(v); err != nil {
			t.Errorf("%q deberia aceptarse: %v", v, err)
		}
	}
	for _, v := range []string{"", "xlsx", "json"} {
		if _, err := ValidarFormatoReporte(v); !errors.Is(err, ErrFormatoReporteInvalido) {
			t.Errorf("%q deberia rechazarse", v)
		}
	}
}

func TestI18SeudonimoEstableEIrreversible(t *testing.T) {
	if _, err := NuevoSeudonimizador([]byte("corta")); !errors.Is(err, ErrClaveSeudonimoInvalida) {
		t.Fatal("una clave corta debe rechazarse")
	}
	s := seudonimizadorI18(t)
	a1, a2, b := s.Seudonimo("escolar-1"), s.Seudonimo("escolar-1"), s.Seudonimo("escolar-2")
	if a1 != a2 || a1 == b {
		t.Fatal("el seudonimo debe ser estable y distinto por escolar")
	}
	if !strings.HasPrefix(a1, PrefijoSeudonimo) || len(a1) != len(PrefijoSeudonimo)+10 || strings.Contains(a1, "escolar") {
		t.Fatalf("formato inesperado: %q", a1)
	}
	otra, _ := NuevoSeudonimizador([]byte("otra-clave-distinta-de-prueba"))
	if otra.Seudonimo("escolar-1") == a1 {
		t.Fatal("con otra clave el seudonimo debe cambiar")
	}
}

func TestI18ReporteSoloVerificadoSumaYNoExponeEscolar(t *testing.T) {
	f := filtroI18(t, "")
	ahora := f.Desde().AddDate(0, 0, 20)
	inicio := f.Desde().Add(10 * time.Hour)

	verificada := sesionI18(t, "s-verif", "tutor-a", "escolar-1", inicio)
	rVerif := conMinutos(t, verificada, 45)
	if err := rVerif.Aprobar("coord", ahora); err != nil {
		t.Fatal(err)
	}

	ajustada := sesionI18(t, "s-ajuste", "tutor-a", "escolar-2", inicio.Add(24*time.Hour))
	rAjuste := conMinutos(t, ajustada, 60)
	if err := rAjuste.Ajustar("coord", 50, "termino antes", ahora); err != nil {
		t.Fatal(err)
	}

	propuesta := sesionI18(t, "s-prop", "tutor-b", "escolar-1", inicio.Add(48*time.Hour))
	rProp := conMinutos(t, propuesta, 30)

	sinMinutos := sesionI18(t, "s-sin", "tutor-b", "escolar-3", inicio.Add(-time.Hour))
	fuera := sesionI18(t, "s-fuera", "tutor-a", "escolar-1", f.Hasta().Add(time.Hour))

	reporte, err := ConstruirReporte("rep-1", f, "coord", ahora, []EntradaReporte{
		{Sesion: propuesta, Minutos: rProp},
		{Sesion: verificada, Minutos: rVerif},
		{Sesion: fuera},
		{Sesion: ajustada, Minutos: rAjuste},
		{Sesion: sinMinutos},
	}, map[string]string{"tutor-a": "Ana Quispe"}, seudonimizadorI18(t))
	if err != nil {
		t.Fatal(err)
	}

	if len(reporte.Filas) != 4 {
		t.Fatalf("esperaba 4 filas (sin la de fuera del periodo), hay %d", len(reporte.Filas))
	}
	orden := []string{"s-sin", "s-verif", "s-ajuste", "s-prop"}
	for i, id := range orden {
		if reporte.Filas[i].SesionID != id {
			t.Fatalf("fila %d: esperaba %s, obtuve %s", i, id, reporte.Filas[i].SesionID)
		}
	}
	for _, fila := range reporte.Filas {
		if strings.Contains(fila.EscolarSeudonimo, "escolar") {
			t.Fatalf("la fila expone el ID del escolar: %+v", fila)
		}
	}
	sin, verif, ajuste, prop := reporte.Filas[0], reporte.Filas[1], reporte.Filas[2], reporte.Filas[3]
	if sin.EstadoMinutos != EstadoSinRegistro || sin.MinutosComputables != 0 || sin.UltimaAccion != nil || sin.Asistencia != AsistenciaSinDeclarar {
		t.Fatalf("fila sin minutos inesperada: %+v", sin)
	}
	if verif.MinutosComputables != 45 || verif.Asistencia != AsistenciaCompleta || verif.TutorNombre != "Ana Quispe" || verif.UltimaAccion.Accion != AccionAprobar {
		t.Fatalf("45 verificados deben sumar 45: %+v", verif)
	}
	if ajuste.MinutosComputables != 50 || ajuste.Ajustes != 1 || ajuste.MinutosPropuestos != 60 {
		t.Fatalf("el ajuste debe contarse y sumar el valor ajustado: %+v", ajuste)
	}
	if prop.MinutosComputables != 0 || prop.TutorNombre != "tutor-b" {
		t.Fatalf("lo propuesto no suma y sin nombre se usa el ID: %+v", prop)
	}
	if verif.EscolarSeudonimo != prop.EscolarSeudonimo {
		t.Fatal("el mismo escolar debe tener el mismo seudonimo en todo el reporte")
	}

	if len(reporte.Totales) != 2 {
		t.Fatalf("esperaba 2 tutores, hay %d", len(reporte.Totales))
	}
	ana, b := reporte.Totales[0], reporte.Totales[1]
	if ana.TutorID != "tutor-a" || ana.MinutosVerificados != 95 || ana.Sesiones != 2 {
		t.Fatalf("total de Ana inesperado: %+v", ana)
	}
	if b.MinutosVerificados != 0 || b.PorEstado[MinutosPropuestos] != 1 || b.PorEstado[EstadoSinRegistro] != 1 {
		t.Fatalf("total del tutor b inesperado: %+v", b)
	}
	if reporte.Aviso != AvisoSinValidezOficial {
		t.Fatal("el reporte debe llevar el aviso de que no tiene validez oficial")
	}
}

func TestI18ReporteDisputaYAusencia(t *testing.T) {
	f := filtroI18(t, "")
	inicio := f.Desde().Add(10 * time.Hour)
	ahora := f.Desde().AddDate(0, 0, 3)

	disputada := sesionI18(t, "s-disp", "tutor-a", "escolar-1", inicio)
	r := conMinutos(t, disputada, 60)
	if err := r.Disputar("escolar-1", "termino a los 40", ahora); err != nil {
		t.Fatal(err)
	}
	ausente := sesionI18(t, "s-aus", "tutor-a", "escolar-2", inicio.Add(time.Hour*3))
	if err := ausente.ReportarAusencia("escolar-2", ausente.Inicio()); err != nil {
		t.Fatal(err)
	}
	reporte, err := ConstruirReporte("rep-2", f, "coord", ahora,
		[]EntradaReporte{{Sesion: disputada, Minutos: r}, {Sesion: ausente}}, nil, seudonimizadorI18(t))
	if err != nil {
		t.Fatal(err)
	}
	if reporte.Filas[0].EstadoMinutos != MinutosEnDisputa || reporte.Filas[0].MinutosComputables != 0 {
		t.Fatalf("una disputa congela (suma 0): %+v", reporte.Filas[0])
	}
	if reporte.Filas[1].Asistencia != AsistenciaConAusencia {
		t.Fatalf("la ausencia debe verse en el reporte: %+v", reporte.Filas[1])
	}
	if reporte.Totales[0].MinutosVerificados != 0 {
		t.Fatal("nada verificado: el total debe ser 0")
	}
}

func TestI18ReporteRechazaDatosIncoherentes(t *testing.T) {
	f := filtroI18(t, "")
	s := seudonimizadorI18(t)
	ahora := f.Desde().AddDate(0, 0, 2)
	if _, err := ConstruirReporte("", f, "coord", ahora, nil, nil, s); !errors.Is(err, ErrReporteInvalido) {
		t.Fatal("sin ID debe rechazarse")
	}
	if _, err := ConstruirReporte("r", f, " ", ahora, nil, nil, s); !errors.Is(err, ErrReporteInvalido) {
		t.Fatal("sin autor debe rechazarse")
	}
	if _, err := ConstruirReporte("r", f, "coord", ahora, nil, nil, Seudonimizador{}); !errors.Is(err, ErrReporteInvalido) {
		t.Fatal("sin clave de seudonimo debe rechazarse")
	}
	a := sesionI18(t, "s-a", "tutor", "esc", f.Desde().Add(time.Hour))
	b := sesionI18(t, "s-b", "tutor", "esc", f.Desde().Add(2*time.Hour))
	rB := conMinutos(t, b, 30)
	if _, err := ConstruirReporte("r", f, "coord", ahora, []EntradaReporte{{Sesion: a, Minutos: rB}}, nil, s); !errors.Is(err, ErrReporteInvalido) {
		t.Fatal("minutos de otra sesion deben rechazarse")
	}
	vacio, err := ConstruirReporte("r", f, "coord", ahora, nil, nil, s)
	if err != nil || len(vacio.Filas) != 0 || len(vacio.Totales) != 0 {
		t.Fatalf("un periodo sin sesiones es un reporte valido y vacio: %v", err)
	}
}

func TestI18RegistroReporteGuardaHuella(t *testing.T) {
	f := filtroI18(t, "tutor-a")
	ahora := f.Desde().AddDate(0, 0, 1)
	reporte, err := ConstruirReporte("rep-3", f, "coord", ahora, nil, nil, seudonimizadorI18(t))
	if err != nil {
		t.Fatal(err)
	}
	contenido := []byte("contenido del archivo")
	registro, err := NuevoRegistroReporte(reporte, FormatoCSV, contenido)
	if err != nil {
		t.Fatal(err)
	}
	esperada := sha256.Sum256(contenido)
	if registro.SHA256 != hex.EncodeToString(esperada[:]) || registro.TutorID != "tutor-a" || registro.GeneradoPor != "coord" || registro.Filas != 0 {
		t.Fatalf("registro inesperado: %+v", registro)
	}
	if _, err := NuevoRegistroReporte(reporte, FormatoCSV, nil); !errors.Is(err, ErrReporteInvalido) {
		t.Fatal("un archivo vacio no se registra")
	}
	if _, err := NuevoRegistroReporte(reporte, "xlsx", contenido); !errors.Is(err, ErrFormatoReporteInvalido) {
		t.Fatal("un formato desconocido no se registra")
	}
}
