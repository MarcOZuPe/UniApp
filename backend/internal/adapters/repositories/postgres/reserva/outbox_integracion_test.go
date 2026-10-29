package reserva_test

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"
	reservaRepo "uniapp/internal/adapters/repositories/postgres/reserva"
	sesionRepo "uniapp/internal/adapters/repositories/postgres/sesion"
	sesionApp "uniapp/internal/application/sesion"
	acceso "uniapp/internal/domain/acceso"
	reserva "uniapp/internal/domain/reserva"
	sesionDominio "uniapp/internal/domain/sesion"
)

type correoI10 struct {
	falla    bool
	mensajes []acceso.Mensaje
}

func (m *correoI10) Enviar(_ context.Context, v acceso.Mensaje) error {
	m.mensajes = append(m.mensajes, v)
	if m.falla {
		return errors.New("proveedor caído con detalle privado")
	}
	return nil
}

func prepararI10(t *testing.T) (*sql.DB, *reserva.Reserva, time.Time) {
	t.Helper()
	db := baseOutboxAislada(t)
	ahora := time.Now().UTC().Truncate(time.Second)
	escolar := crearEscolar(t, db, "escolar-i10@example.test")
	var tutor string
	if err := db.QueryRow(`INSERT INTO usuarios(email,nombre_uso,clave_hash,rol_id) VALUES ('tutor-i10@example.test','Tutor','hash',3) RETURNING id`).Scan(&tutor); err != nil {
		t.Fatal(err)
	}
	r := reservaPara(escolar, "slot-i10", ahora.Add(72*time.Hour))
	r.TutorID = tutor
	r.CreadaEn = ahora
	r.ClaveIdempotencia = "clave-i10"
	return db, r, ahora
}
func contarI10(t *testing.T, db *sql.DB, consulta string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(consulta).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}
func vaciarPendientesI10(t *testing.T, p sesionApp.ProcesadorOutbox) {
	t.Helper()
	for i := 0; i < 30; i++ {
		hay, err := p.ProcesarUno(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if !hay {
			return
		}
	}
	t.Fatal("procesador no termina")
}

func TestI10TransaccionIdempotenciaYAgenda(t *testing.T) {
	db, r, ahora := prepararI10(t)
	repo := reservaRepo.RepositorioPostgres{DB: db, AnticipacionRecordatorio: 2 * time.Hour}
	guardada, nueva, err := repo.ReservarTransaccional(context.Background(), r, reserva.LimitesPorDefecto)
	if err != nil || !nueva {
		t.Fatalf("crear: %v", err)
	}
	repetida, nueva, err := repo.ReservarTransaccional(context.Background(), r, reserva.LimitesPorDefecto)
	if err != nil || nueva || repetida.ID != guardada.ID {
		t.Fatalf("idempotencia: %v", err)
	}
	if n := contarI10(t, db, "SELECT count(*) FROM sesion_outbox"); n != 8 {
		t.Fatalf("entregas: %d", n)
	}
	var disponible time.Time
	if err := db.QueryRow("SELECT disponible_en FROM sesion_outbox WHERE tipo='recordatorio' LIMIT 1").Scan(&disponible); err != nil || !disponible.Equal(ahora.Add(70*time.Hour)) {
		t.Fatalf("programación: %v %v", disponible, err)
	}
	for _, id := range []string{r.TutorID, "otro"} {
		items, err := repo.PorTutor(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if id == r.TutorID {
			want = 1
		}
		if len(items) != want {
			t.Fatal("filtro tutor incorrecto")
		}
	}
	propios, err := repo.PorEscolar(context.Background(), r.EscolarID)
	if err != nil || len(propios) != 1 {
		t.Fatal("agenda escolar incorrecta")
	}
	ajenos, err := repo.PorEscolar(context.Background(), r.TutorID)
	if err != nil || len(ajenos) != 0 {
		t.Fatal("filtro escolar incorrecto")
	}
}

func TestI10FalloOutboxRevierteReserva(t *testing.T) {
	db, r, _ := prepararI10(t)
	// Inyección de fallo solo en la base desechable de este test.
	if _, err := db.Exec("ALTER TABLE sesion_outbox ADD CONSTRAINT fallo_prueba CHECK (false)"); err != nil {
		t.Fatal(err)
	}
	_, _, err := (reservaRepo.RepositorioPostgres{DB: db}).ReservarTransaccional(context.Background(), r, reserva.LimitesPorDefecto)
	if err == nil {
		t.Fatal("se esperaba fallo del outbox")
	}
	if contarI10(t, db, "SELECT count(*) FROM reservas") != 0 || contarI10(t, db, "SELECT count(*) FROM sesion_outbox") != 0 {
		t.Fatal("rollback incompleto")
	}
}

func TestI10FalloProveedorReintentosYRecordatorios(t *testing.T) {
	db, r, ahora := prepararI10(t)
	repo := reservaRepo.RepositorioPostgres{DB: db}
	guardada, _, err := repo.ReservarTransaccional(context.Background(), r, reserva.LimitesPorDefecto)
	if err != nil {
		t.Fatal(err)
	}
	out := sesionRepo.RepositorioOutboxPostgres{DB: db}
	mailer := &correoI10{falla: true}
	reloj := ahora
	p := sesionApp.ProcesadorOutbox{Repositorio: out, Mailer: mailer, Ahora: func() time.Time { return reloj }}
	vaciarPendientesI10(t, p)
	if len(mailer.mensajes) != 2 {
		t.Fatalf("correos intentados: %d", len(mailer.mensajes))
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE canal='interno' AND estado='entregado'") != 2 {
		t.Fatal("faltan confirmaciones internas")
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE canal='correo' AND estado='pendiente' AND intentos=1") != 2 {
		t.Fatal("no conservó reintentos")
	}
	leida, err := repo.PorID(context.Background(), guardada.ID)
	if err != nil || leida.Estado != reserva.EstadoConfirmada {
		t.Fatal("se perdió reserva confirmada")
	}
	for _, id := range []string{r.EscolarID, r.TutorID, "ajeno"} {
		b, err := out.NotificacionesDe(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		want := 1
		if id == "ajeno" {
			want = 0
		}
		if len(b) != want {
			t.Fatal("bandeja mezcla destinatarios")
		}
	}
	vaciarPendientesI10(t, p)
	if len(mailer.mensajes) != 2 {
		t.Fatal("reintentó antes de plazo")
	}
	for intento := 1; intento < sesionDominio.MaxIntentosOutbox; intento++ {
		reloj = reloj.Add(sesionDominio.EsperaOutbox(intento))
		vaciarPendientesI10(t, p)
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE estado='agotado'") != 2 {
		t.Fatal("no limitó intentos")
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE canal='interno' AND estado='entregado'") != 2 {
		t.Fatal("duplicó internos")
	}
	// Reiniciar el procesador no pierde recordatorios programados.
	mailer.falla = false
	reloj = ahora.Add(48*time.Hour - time.Second)
	nuevo := sesionApp.ProcesadorOutbox{Repositorio: sesionRepo.RepositorioOutboxPostgres{DB: db}, Mailer: mailer, Ahora: func() time.Time { return reloj }}
	vaciarPendientesI10(t, nuevo)
	if len(mailer.mensajes) != 10 {
		t.Fatal("recordatorio anticipado")
	}
	reloj = reloj.Add(time.Second)
	vaciarPendientesI10(t, nuevo)
	if len(mailer.mensajes) != 12 {
		t.Fatal("faltan correos recordatorio")
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE tipo='recordatorio' AND estado='entregado'") != 4 {
		t.Fatal("faltan recordatorios")
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox") != 8 {
		t.Fatal("duplicó eventos")
	}
	for _, m := range mailer.mensajes {
		if m.Para != "escolar-i10@example.test" && m.Para != "tutor-i10@example.test" {
			t.Fatal("destinatario incorrecto")
		}
	}
}

func TestI10ReintentoExitosoSinDuplicarInterno(t *testing.T) {
	db, r, ahora := prepararI10(t)
	_, _, err := (reservaRepo.RepositorioPostgres{DB: db}).ReservarTransaccional(context.Background(), r, reserva.LimitesPorDefecto)
	if err != nil {
		t.Fatal(err)
	}
	mailer := &correoI10{falla: true}
	reloj := ahora
	p := sesionApp.ProcesadorOutbox{Repositorio: sesionRepo.RepositorioOutboxPostgres{DB: db}, Mailer: mailer, Ahora: func() time.Time { return reloj }}
	vaciarPendientesI10(t, p)
	mailer.falla = false
	reloj = reloj.Add(time.Minute)
	vaciarPendientesI10(t, p)
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE tipo='confirmacion' AND estado='entregado'") != 4 {
		t.Fatal("reintento no completó canales")
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE canal='interno' AND intentos=1 AND estado='entregado'") != 2 {
		t.Fatal("reintentó canal ya completado")
	}
}

func TestI10ReclamosConcurrentesYRecuperacion(t *testing.T) {
	db, r, ahora := prepararI10(t)
	_, _, err := (reservaRepo.RepositorioPostgres{DB: db}).ReservarTransaccional(context.Background(), r, reserva.LimitesPorDefecto)
	if err != nil {
		t.Fatal(err)
	}
	out := sesionRepo.RepositorioOutboxPostgres{DB: db}
	var wg sync.WaitGroup
	entregas := make(chan *sesionDominio.Entrega, 12)
	fallos := make(chan error, 12)
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e, err := out.Reclamar(context.Background(), ahora)
			if err != nil {
				fallos <- err
			}
			if e != nil {
				entregas <- e
			}
		}()
	}
	wg.Wait()
	close(entregas)
	close(fallos)
	for err := range fallos {
		t.Fatal(err)
	}
	ids := map[int64]bool{}
	var viejo sesionDominio.Entrega
	for e := range entregas {
		if ids[e.ID] {
			t.Fatal("reclamo duplicado")
		}
		ids[e.ID] = true
		viejo = *e
	}
	if len(ids) != 4 {
		t.Fatalf("reclamos: %d", len(ids))
	}
	if e, err := out.Reclamar(context.Background(), ahora.Add(time.Second)); err != nil || e != nil {
		t.Fatal("reclamó lease vigente")
	}
	reloj := ahora.Add(sesionDominio.PlazoReclamoOutbox + time.Second)
	for i := 0; i < 4; i++ {
		e, err := out.Reclamar(context.Background(), reloj)
		if err != nil || e == nil || e.Intentos != 2 {
			t.Fatal("no recuperó pendientes")
		}
	}
	if err := out.Resolver(context.Background(), viejo, reloj, ""); !errors.Is(err, sesionDominio.ErrReclamoVencido) {
		t.Fatal("worker antiguo pudo sobrescribir")
	}
	if _, err := db.Exec("UPDATE sesion_outbox SET intentos=$1 WHERE estado='procesando'", sesionDominio.MaxIntentosOutbox); err != nil {
		t.Fatal(err)
	}
	if _, err := out.Reclamar(context.Background(), reloj.Add(sesionDominio.PlazoReclamoOutbox)); err != nil {
		t.Fatal(err)
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE estado='agotado'") != 4 {
		t.Fatal("último intento caído quedó bloqueado")
	}
}

func TestI10ReservasConcurrentesNoDuplicanEventos(t *testing.T) {
	for _, mismoIntento := range []bool{true, false} {
		t.Run(map[bool]string{true: "idempotencia", false: "slot_ocupado"}[mismoIntento], func(t *testing.T) {
			db, r, _ := prepararI10(t)
			repo := reservaRepo.RepositorioPostgres{DB: db}
			var wg sync.WaitGroup
			var mu sync.Mutex
			creadas := 0
			ids := map[string]bool{}
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					copia := *r
					if !mismoIntento {
						copia.ClaveIdempotencia = ""
					}
					guardada, nueva, err := repo.ReservarTransaccional(context.Background(), &copia, reserva.LimitesPorDefecto)
					mu.Lock()
					defer mu.Unlock()
					if err != nil {
						if mismoIntento || !errors.Is(err, reserva.ErrSlotOcupado) {
							t.Errorf("reserva concurrente: %v", err)
						}
						return
					}
					if nueva {
						creadas++
					}
					ids[guardada.ID] = true
				}(i)
			}
			wg.Wait()
			if creadas != 1 || len(ids) != 1 {
				t.Fatalf("creadas=%d ids=%d", creadas, len(ids))
			}
			if contarI10(t, db, "SELECT count(*) FROM reservas") != 1 || contarI10(t, db, "SELECT count(*) FROM sesion_outbox") != 8 {
				t.Fatal("duplicó reserva o eventos")
			}
		})
	}
}

func TestI10TutorSinteticoNoSeAsignaAOtroUsuario(t *testing.T) {
	db, r, ahora := prepararI10(t)
	tutorReal := r.TutorID
	r.TutorID = reserva.TutorSinteticoID
	repo := reservaRepo.RepositorioPostgres{DB: db}
	if _, _, err := repo.ReservarTransaccional(context.Background(), r, reserva.LimitesPorDefecto); err != nil {
		t.Fatal(err)
	}
	agenda, err := repo.PorTutor(context.Background(), tutorReal)
	if err != nil || len(agenda) != 0 {
		t.Fatal("reasignó reserva sintética")
	}
	out := sesionRepo.RepositorioOutboxPostgres{DB: db}
	mailer := &correoI10{}
	vaciarPendientesI10(t, sesionApp.ProcesadorOutbox{Repositorio: out, Mailer: mailer, Ahora: func() time.Time { return ahora }})
	if len(mailer.mensajes) != 1 || mailer.mensajes[0].Para != "escolar-i10@example.test" {
		t.Fatal("envió a tutor equivocado")
	}
	if contarI10(t, db, "SELECT count(*) FROM sesion_outbox WHERE ultimo_error='destinatario_no_disponible'") != 1 {
		t.Fatal("no registró dependencia de identidad real")
	}
}
