package incidente

import (
	"context"
	"testing"
	"time"
)

func TestSembrarTutoriaSinteticaEsIdempotente(t *testing.T) {
	db := baseAisladaIncidente(t)
	ctx := context.Background()
	escolarID := crearUsuarioIncidente(t, db, "escolar.prueba@example.test", 1)
	tutorID := crearUsuarioIncidente(t, db, "tutor.prueba@uni.edu.pe", 3)

	if err := SembrarTutoriaSintetica(ctx, db, "uni.edu.pe"); err != nil {
		t.Fatal(err)
	}
	if err := SembrarTutoriaSintetica(ctx, db, "uni.edu.pe"); err != nil {
		t.Fatalf("segunda siembra: %v", err)
	}

	var reservas, sesiones, asistencias int
	if err := db.QueryRow(`SELECT count(*) FROM reservas WHERE slot_id = $1`, slotTutoriaSintetica).Scan(&reservas); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM sesion_sesiones WHERE id = $1`, sesionTutoriaSintetica).Scan(&sesiones); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM sesion_asistencias WHERE sesion_id = $1`, sesionTutoriaSintetica).Scan(&asistencias); err != nil {
		t.Fatal(err)
	}
	if reservas != 1 || sesiones != 1 || asistencias != 1 {
		t.Fatalf("semilla duplicada: reservas=%d sesiones=%d asistencias=%d", reservas, sesiones, asistencias)
	}

	tutorias, err := (RepositorioPostgres{DB: db}).TutoriasDeEscolar(ctx, escolarID, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if len(tutorias) != 1 || tutoriaDeSintetica(tutorias, tutorID) == nil || !tutorias[0].Asistio {
		t.Fatalf("la tutoría sintética no está disponible para I13: %+v", tutorias)
	}
}

func tutoriaDeSintetica(tutorias []Tutoria, tutorID string) *Tutoria {
	for indice := range tutorias {
		if tutorias[indice].TutorID == tutorID {
			return &tutorias[indice]
		}
	}
	return nil
}
