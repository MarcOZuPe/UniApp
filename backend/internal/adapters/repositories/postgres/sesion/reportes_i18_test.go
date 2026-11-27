package sesion

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestI18PostgresSesionesEntreYConstancia(t *testing.T) {
	db := baseAisladaMinutos(t) // base nueva con todas las migraciones, incluida 020
	ctx := context.Background()
	sesiones := RepositorioPostgres{DB: db}
	reportes := RepositorioReportesPostgres{DB: db}

	desde := time.Date(2026, 11, 1, 5, 0, 0, 0, time.UTC)
	crear := func(id, tutor string, inicio time.Time) {
		s, err := NuevaSesion(id, "reserva-"+id, tutor, "escolar-"+id, inicio, inicio.Add(DuracionSesion))
		if err != nil {
			t.Fatal(err)
		}
		if err := sesiones.CrearSesion(ctx, s); err != nil {
			t.Fatal(err)
		}
	}
	crear("s-b", "tutor-a", desde.Add(48*time.Hour))
	crear("s-a", "tutor-a", desde.Add(24*time.Hour))
	crear("s-otro", "tutor-b", desde.Add(30*time.Hour))
	crear("s-borde", "tutor-a", desde.AddDate(0, 0, 30)) // inicio == hasta: fuera
	if _, err := sesiones.DeclararAsistencia(ctx, "s-a", "tutor-a", desde.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}

	filtro, err := NuevoFiltroReporte(desde, desde.AddDate(0, 0, 30), "tutor-a")
	if err != nil {
		t.Fatal(err)
	}
	encontradas, err := reportes.SesionesEntre(ctx, filtro)
	if err != nil {
		t.Fatal(err)
	}
	if len(encontradas) != 2 || encontradas[0].ID() != "s-a" || encontradas[1].ID() != "s-b" {
		t.Fatalf("esperaba s-a y s-b en orden: %v", encontradas)
	}
	if len(encontradas[0].Asistencias()) != 1 {
		t.Fatal("las sesiones deben cargarse completas (con asistencias)")
	}

	sinTutor, _ := NuevoFiltroReporte(desde, desde.AddDate(0, 0, 30), "")
	todas, err := reportes.SesionesEntre(ctx, sinTutor)
	if err != nil || len(todas) != 3 {
		t.Fatalf("sin filtro de tutor esperaba 3: %d %v", len(todas), err)
	}

	registro := RegistroReporte{
		ID: "rep-pg", GeneradoPor: "coord", GeneradoEn: desde.AddDate(0, 0, 31),
		Desde: filtro.Desde(), Hasta: filtro.Hasta(), TutorID: "tutor-a",
		Formato: FormatoPDF, Filas: 2, SHA256: strings.Repeat("a", 64),
	}
	if err := reportes.RegistrarReporte(ctx, registro); err != nil {
		t.Fatal(err)
	}
	if err := reportes.RegistrarReporte(ctx, registro); !errors.Is(err, ErrReporteInvalido) {
		t.Fatalf("un ID repetido debe rechazarse: %v", err)
	}
	recientes, err := reportes.ReportesRecientes(ctx, 10)
	if err != nil || len(recientes) != 1 || recientes[0] != registro {
		t.Fatalf("constancia leida distinta: %+v %v", recientes, err)
	}

	if _, err := db.ExecContext(ctx, `UPDATE sesion_reportes SET filas = 99 WHERE id = 'rep-pg'`); err == nil {
		t.Fatal("la constancia debe ser inmutable (UPDATE)")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM sesion_reportes WHERE id = 'rep-pg'`); err == nil {
		t.Fatal("la constancia debe ser inmutable (DELETE)")
	}
}
