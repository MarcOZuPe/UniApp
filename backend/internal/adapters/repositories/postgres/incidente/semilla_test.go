package incidente

import (
	"context"
	"testing"
	"time"
)

func TestSembrarTutoriasSinteticasEsIdempotente(t *testing.T) {
	db := baseAisladaIncidente(t)
	ctx := context.Background()
	escolarID := crearUsuarioIncidente(t, db, emailEscolarSintetico, 1)
	escolarDosID := crearUsuarioIncidente(t, db, emailEscolarSinteticoDos, 1)
	tutorID := crearUsuarioIncidente(t, db, "tutor.prueba@uni.edu.pe", 3)
	tutorDosID := crearUsuarioIncidente(t, db, "tutor2.prueba@uni.edu.pe", 3)

	if err := SembrarTutoriasSinteticas(ctx, db, "uni.edu.pe"); err != nil {
		t.Fatal(err)
	}
	if err := SembrarTutoriasSinteticas(ctx, db, "uni.edu.pe"); err != nil {
		t.Fatalf("segunda siembra: %v", err)
	}

	var reservas, sesiones, asistencias int
	if err := db.QueryRow(`
		SELECT count(*) FROM reservas WHERE slot_id IN ($1, $2)`,
		slotTutoriaSintetica, slotTutoriaSinteticaDos,
	).Scan(&reservas); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
		SELECT count(*) FROM sesion_sesiones WHERE id IN ($1, $2)`,
		sesionTutoriaSintetica, sesionTutoriaSinteticaDos,
	).Scan(&sesiones); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
		SELECT count(*) FROM sesion_asistencias WHERE sesion_id IN ($1, $2)`,
		sesionTutoriaSintetica, sesionTutoriaSinteticaDos,
	).Scan(&asistencias); err != nil {
		t.Fatal(err)
	}
	if reservas != 2 || sesiones != 2 || asistencias != 2 {
		t.Fatalf("semilla duplicada o incompleta: reservas=%d sesiones=%d asistencias=%d", reservas, sesiones, asistencias)
	}

	repo := RepositorioPostgres{DB: db}
	ahora := time.Now().UTC()
	tutorias := tutoriasDe(t, repo, ctx, escolarID, ahora)
	if len(tutorias) != 1 || tutoriaDeSintetica(tutorias, tutorID) == nil ||
		tutorias[0].Curso != "Álgebra" || !tutorias[0].Asistio {
		t.Fatalf("tutoría sintética 1 no disponible: %+v", tutorias)
	}
	tutoriasDos := tutoriasDe(t, repo, ctx, escolarDosID, ahora)
	if len(tutoriasDos) != 1 || tutoriaDeSintetica(tutoriasDos, tutorDosID) == nil ||
		tutoriasDos[0].Curso != "Biología" || !tutoriasDos[0].Asistio {
		t.Fatalf("tutoría sintética 2 no disponible: %+v", tutoriasDos)
	}
}

func tutoriasDe(
	t *testing.T,
	repo RepositorioPostgres,
	ctx context.Context,
	escolarID string,
	ahora time.Time,
) []Tutoria {
	t.Helper()
	tutorias, err := repo.TutoriasDeEscolar(ctx, escolarID, ahora)
	if err != nil {
		t.Fatal(err)
	}
	return tutorias
}

func tutoriaDeSintetica(tutorias []Tutoria, tutorID string) *Tutoria {
	for indice := range tutorias {
		if tutorias[indice].TutorID == tutorID {
			return &tutorias[indice]
		}
	}
	return nil
}
