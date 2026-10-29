package reserva_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	repo "uniapp/internal/adapters/repositories/postgres/reserva"
	modelo "uniapp/internal/domain/reserva"
	"uniapp/migrations"
)

// baseAislada crea una base nueva para este paquete. Evita que otras pruebas que
// truncan usuarios interfieran con las carreras, y comprueba todas las migraciones
// desde cero.
func baseAislada(t *testing.T) *sql.DB {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL no definida")
	}
	administrador, err := sql.Open("postgres", base)
	if err != nil {
		t.Fatal(err)
	}
	nombre := fmt.Sprintf("uniapp_reserva_%d", time.Now().UnixNano())
	if _, err := administrador.Exec("CREATE DATABASE " + nombre); err != nil {
		t.Fatalf("crear base de prueba: %v", err)
	}
	t.Cleanup(func() {
		if _, err := administrador.Exec("DROP DATABASE IF EXISTS " + nombre); err != nil {
			t.Logf("no se pudo borrar la base de prueba: %v", err)
		}
		_ = administrador.Close()
	})

	destino, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	destino.Path = "/" + nombre
	baseDatos, err := migrations.Abrir(destino.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = baseDatos.Close() })
	if err := migrations.Migrar(context.Background(), baseDatos, migrations.FS); err != nil {
		t.Fatalf("migraciones: %v", err)
	}
	return baseDatos
}

func crearEscolar(t *testing.T, baseDatos *sql.DB, email string) string {
	t.Helper()
	var id string
	err := baseDatos.QueryRow(`INSERT INTO usuarios (email, nombre_uso, clave_hash, rol_id) VALUES ($1, 'Prueba', 'hash', 1) RETURNING id`, email).Scan(&id)
	if err != nil {
		t.Fatalf("crear escolar: %v", err)
	}
	return id
}

func reservaPara(escolarID, slotID string, inicio time.Time) *modelo.Reserva {
	return &modelo.Reserva{
		EscolarID: escolarID, TutorID: modelo.TutorSinteticoID, TutorNombre: modelo.TutorSinteticoNombre,
		SlotID: slotID, Curso: "Álgebra", Grado: "5to de secundaria",
		Inicio: inicio, Fin: inicio.Add(modelo.DuracionSesion),
		Tema: "Repaso de factorización", Objetivo: "Resolver ejercicios de factorización de trinomios sin ayuda.",
		Estado: modelo.EstadoConfirmada, CreadaEn: time.Now().UTC(),
	}
}

func TestCarreraPorElMismoSlot(t *testing.T) {
	baseDatos := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	escolar := crearEscolar(t, baseDatos, "carrera@example.test")
	inicio := time.Now().UTC().Add(72 * time.Hour)

	const intentos = 10
	var exitos, ocupados int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < intentos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, creada, err := repositorio.ReservarTransaccional(context.Background(), reservaPara(escolar, "slot-carrera", inicio), modelo.LimitesPorDefecto)
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil && creada:
				exitos++
			case errors.Is(err, modelo.ErrSlotOcupado):
				ocupados++
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}()
	}
	wg.Wait()
	if exitos != 1 || ocupados != intentos-1 {
		t.Fatalf("exactamente una reserva debe ganar: exitos=%d ocupados=%d", exitos, ocupados)
	}
}

func TestIdempotenciaConcurrenteDevuelveLaMismaReserva(t *testing.T) {
	baseDatos := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	escolar := crearEscolar(t, baseDatos, "idempotencia@example.test")
	inicio := time.Now().UTC().Add(96 * time.Hour)

	const intentos = 10
	var creadas int
	ids := map[string]bool{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for i := 0; i < intentos; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			reserva := reservaPara(escolar, "slot-idempotente", inicio)
			reserva.ClaveIdempotencia = "clave-unica-del-formulario"
			guardada, creada, err := repositorio.ReservarTransaccional(context.Background(), reserva, modelo.LimitesPorDefecto)
			if err != nil {
				t.Errorf("error inesperado: %v", err)
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if creada {
				creadas++
			}
			ids[guardada.ID] = true
		}()
	}
	wg.Wait()
	if creadas != 1 || len(ids) != 1 {
		t.Fatalf("una sola reserva y un solo identificador: creadas=%d ids=%d", creadas, len(ids))
	}
}

func TestSolapamientoConcurrenteDelEscolar(t *testing.T) {
	baseDatos := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	escolar := crearEscolar(t, baseDatos, "solape@example.test")
	inicio := time.Now().UTC().Add(72 * time.Hour)

	resultados := make(chan error, 2)
	// Cursos distintos para que el único motivo posible sea el cruce de horarios.
	for i, desplazamiento := range []time.Duration{0, 30 * time.Minute} {
		go func(indice int, desde time.Time) {
			reserva := reservaPara(escolar, fmt.Sprintf("slot-solape-%d", indice), desde)
			reserva.Curso = []string{"Álgebra", "Geometría"}[indice]
			_, _, err := repositorio.ReservarTransaccional(context.Background(), reserva, modelo.LimitesPorDefecto)
			resultados <- err
		}(i, inicio.Add(desplazamiento))
	}
	var exitos, solapados int
	for i := 0; i < 2; i++ {
		switch err := <-resultados; {
		case err == nil:
			exitos++
		case errors.Is(err, modelo.ErrSolapamiento):
			solapados++
		default:
			t.Fatalf("error inesperado: %v", err)
		}
	}
	if exitos != 1 || solapados != 1 {
		t.Fatalf("una sesión cruzada debe rechazarse: exitos=%d solapados=%d", exitos, solapados)
	}
}

func TestCuotaConcurrenteDelEscolar(t *testing.T) {
	baseDatos := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	escolar := crearEscolar(t, baseDatos, "cuota@example.test")
	inicio := time.Now().UTC().Add(72 * time.Hour)

	cursos := []string{"Álgebra", "Aritmética", "Geometría"}
	resultados := make(chan error, len(cursos))
	for i, curso := range cursos {
		go func(indice int, nombreCurso string) {
			reserva := reservaPara(escolar, fmt.Sprintf("slot-cuota-%d", indice), inicio.Add(time.Duration(indice)*24*time.Hour))
			reserva.Curso = nombreCurso
			_, _, err := repositorio.ReservarTransaccional(context.Background(), reserva, modelo.LimitesPorDefecto)
			resultados <- err
		}(i, curso)
	}
	var exitos, rechazadas int
	for i := 0; i < len(cursos); i++ {
		switch err := <-resultados; {
		case err == nil:
			exitos++
		case errors.Is(err, modelo.ErrCuotaEscolar):
			rechazadas++
		default:
			t.Fatalf("error inesperado: %v", err)
		}
	}
	if exitos != modelo.MaxFuturasPorEscolar || rechazadas != len(cursos)-modelo.MaxFuturasPorEscolar {
		t.Fatalf("dos reservas futuras como máximo: exitos=%d rechazadas=%d", exitos, rechazadas)
	}
}

func TestAdjuntosEnPostgres(t *testing.T) {
	baseDatos := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: baseDatos}
	escolar := crearEscolar(t, baseDatos, "adjuntos@example.test")
	inicio := time.Now().UTC().Add(72 * time.Hour)
	reserva, _, err := repositorio.ReservarTransaccional(context.Background(), reservaPara(escolar, "slot-adjuntos", inicio), modelo.LimitesPorDefecto)
	if err != nil {
		t.Fatal(err)
	}
	adjunto, err := modelo.NuevoAdjunto("practica.pdf", "application/pdf", 1024, "huella", "0123456789abcdef0123456789abcdef", modelo.AdjuntoEnCuarentena, "pendiente", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	guardado, err := repositorio.AgregarAdjunto(context.Background(), reserva.ID, adjunto)
	if err != nil {
		t.Fatal(err)
	}
	if err := repositorio.ActualizarAdjunto(context.Background(), guardado.ID, modelo.AdjuntoValidado, ""); err != nil {
		t.Fatal(err)
	}
	leido, err := repositorio.AdjuntoPorID(context.Background(), reserva.ID, guardado.ID)
	if err != nil || leido.Estado != modelo.AdjuntoValidado {
		t.Fatalf("adjunto actualizado: %v %+v", err, leido)
	}
	if _, err := repositorio.AdjuntoPorID(context.Background(), reserva.ID, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, modelo.ErrAdjuntoNoEncontrado) {
		t.Fatalf("adjunto inexistente: %v", err)
	}
}

func TestReservaConfirmadaCreaSesionI05(t *testing.T) {
	db := baseAislada(t)
	repositorio := repo.RepositorioPostgres{DB: db}
	ctx := context.Background()

	escolarID := crearEscolar(
		t,
		db,
		"sesion-i05@example.test",
	)

	inicio := time.Now().UTC().Add(72 * time.Hour)

	nuevaReserva := reservaPara(
		escolarID,
		"slot-sesion-i05",
		inicio,
	)

	nuevaReserva.ClaveIdempotencia = "clave-i05-unica"

	// Primera solicitud: debe crear la reserva.
	guardada, creada, err := repositorio.ReservarTransaccional(
		ctx,
		nuevaReserva,
		modelo.LimitesPorDefecto,
	)

	if err != nil {
		t.Fatalf("crear reserva: %v", err)
	}

	if !creada {
		t.Fatal("la primera solicitud debe crear la reserva")
	}

	// La sesión debe existir y estar vinculada a la reserva.
	var sesionID, reservaID, tutorID, sesionEscolarID string
	var sesionInicio, sesionFin time.Time

	err = db.QueryRowContext(ctx, `
        SELECT id, reserva_id, tutor_id, escolar_id, inicio, fin
        FROM sesion_sesiones
        WHERE reserva_id = $1
    `, guardada.ID).Scan(
		&sesionID,
		&reservaID,
		&tutorID,
		&sesionEscolarID,
		&sesionInicio,
		&sesionFin,
	)

	if err != nil {
		t.Fatalf("la sesión debería existir: %v", err)
	}

	if sesionID != "sesion-"+guardada.ID {
		t.Fatalf("ID de sesión inesperado: %s", sesionID)
	}

	if reservaID != guardada.ID ||
		tutorID != guardada.TutorID ||
		sesionEscolarID != guardada.EscolarID {
		t.Fatal("los participantes o la reserva no coinciden")
	}

	if !sesionInicio.Equal(guardada.Inicio) ||
		!sesionFin.Equal(guardada.Fin) {
		t.Fatal("el horario de sesión no coincide con la reserva")
	}

	// Segunda solicitud: no debe crear otra reserva ni sesión.
	repetida, creadaOtraVez, err := repositorio.ReservarTransaccional(
		ctx,
		nuevaReserva,
		modelo.LimitesPorDefecto,
	)

	if err != nil {
		t.Fatalf("repetir reserva: %v", err)
	}

	if creadaOtraVez || repetida.ID != guardada.ID {
		t.Fatal("la idempotencia de la reserva falló")
	}

	var cantidad int

	err = db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM sesion_sesiones
        WHERE reserva_id = $1
    `, guardada.ID).Scan(&cantidad)

	if err != nil {
		t.Fatalf("contar sesiones: %v", err)
	}

	if cantidad != 1 {
		t.Fatalf("se esperaba una sesión, existen %d", cantidad)
	}
}
