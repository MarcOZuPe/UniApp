package sesion

import (
	"testing"
	"time"
)

func sesionConAsistencias(t *testing.T) (*Sesion, time.Time, time.Time) {
	t.Helper()
	inicio := time.Now().UTC().Add(-3 * time.Hour)
	sesion, err := NuevaSesion("s-min", "r-min", "tutor-1", "escolar-1", inicio, inicio.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err := sesion.DeclararAsistencia("tutor-1", inicio.Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := sesion.DeclararAsistencia("escolar-1", inicio.Add(30*time.Minute)); err != nil {
		t.Fatal(err)
	}
	return sesion, inicio, inicio.Add(time.Hour)
}

func TestProponerYVerificarMinutos(t *testing.T) {
	sesion, inicio, fin := sesionConAsistencias(t)
	ahora := fin.Add(time.Hour)
	if _, err := ProponerMinutos(sesion, "tutor-1", 60, inicio); err == nil {
		t.Fatal("no se puede proponer antes de finalizar")
	}
	if _, err := ProponerMinutos(sesion, "tutor-1", 0, ahora); err != ErrMinutosInvalidos {
		t.Fatalf("se esperaba minutos inválidos: %v", err)
	}
	registro, err := ProponerMinutos(sesion, "tutor-1", 60, ahora)
	if err != nil {
		t.Fatalf("proponer falló: %v", err)
	}
	if registro.Estado() != MinutosPropuestos || registro.MinutosPropuestos() != 60 || len(registro.Auditoria()) != 1 {
		t.Fatalf("registro inesperado: %+v", registro)
	}
	if registro.MinutosComputables() != 0 {
		t.Fatal("lo propuesto no suma")
	}
	if err := registro.Aprobar("tutor-1", ahora); err != ErrAutoaprobacion {
		t.Fatalf("el tutor no puede aprobar: %v", err)
	}
	if err := registro.Aprobar("coord-1", ahora); err != nil {
		t.Fatalf("aprobar falló: %v", err)
	}
	if registro.Estado() != MinutosVerificados || registro.MinutosComputables() != 60 {
		t.Fatalf("aprobación inesperada: %+v", registro)
	}
	if verificados, ok := registro.MinutosVerificados(); !ok || verificados != 60 {
		t.Fatalf("minutos verificados inesperados: %d %v", verificados, ok)
	}
	if err := registro.Ajustar("coord-1", 45, "ajuste", ahora); err != nil {
		t.Fatalf("ajustar falló: %v", err)
	}
	if registro.MinutosComputables() != 45 {
		t.Fatal("el ajuste no se aplicó")
	}
	if err := registro.Ajustar("coord-1", 61, "ajuste", ahora); err != ErrMinutosInvalidos {
		t.Fatalf("se esperaba minutos inválidos: %v", err)
	}
	if err := registro.Ajustar("coord-1", 30, "", ahora); err != ErrMotivoInvalido {
		t.Fatalf("se esperaba motivo inválido: %v", err)
	}
}

func TestDisputaResolucionYAnulacion(t *testing.T) {
	sesion, _, fin := sesionConAsistencias(t)
	ahora := fin.Add(time.Hour)
	registro, err := ProponerMinutos(sesion, "tutor-1", 60, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if err := registro.Disputar("otro", "motivo", ahora); err != ErrNoAutorizado {
		t.Fatalf("se esperaba no autorizado: %v", err)
	}
	if err := registro.Disputar("escolar-1", "", ahora); err != ErrMotivoInvalido {
		t.Fatalf("se esperaba motivo inválido: %v", err)
	}
	if err := registro.Disputar("escolar-1", "no estuve de acuerdo", ahora); err != nil {
		t.Fatalf("disputar falló: %v", err)
	}
	if registro.Estado() != MinutosEnDisputa || registro.MinutosComputables() != 0 {
		t.Fatalf("disputa inesperada: %+v", registro)
	}
	if err := registro.Disputar("escolar-1", "otra", ahora); err != ErrDisputaNoPermitida {
		t.Fatalf("no se debe disputar dos veces: %v", err)
	}
	if err := registro.Resolver("coord-1", 0, "no procede", ahora); err != nil {
		t.Fatalf("resolver falló: %v", err)
	}
	if registro.Estado() != MinutosVerificados || registro.MinutosComputables() != 0 || registro.DisputadoPor() != "" {
		t.Fatalf("resolución inesperada: %+v", registro)
	}
	if err := registro.Anular(ahora); err != nil {
		t.Fatal(err)
	}
	if registro.Estado() != MinutosAnulados || registro.MinutosComputables() != 0 {
		t.Fatalf("anulación inesperada: %+v", registro)
	}
	if err := registro.Anular(ahora); err != nil {
		t.Fatalf("anular dos veces debe ser idempotente: %v", err)
	}
	if len(registro.Auditoria()) != 4 {
		t.Fatalf("auditoría inesperada: %d", len(registro.Auditoria()))
	}
	if _, err := ProponerMinutos(sesion, "tutor-1", 60, fin.Add(8*24*time.Hour)); err == nil {
		t.Fatal("fuera del plazo no se debe proponer")
	}
}

func TestRestaurarRegistroMinutos(t *testing.T) {
	ahora := time.Now().UTC()
	datos := DatosRegistroMinutos{
		SesionID: "s", ReservaID: "r", TutorID: "t", EscolarID: "e", FinSesion: ahora,
		Estado: MinutosPropuestos, Propuestos: 60,
		Auditoria: []DatosEntradaAuditoria{{Accion: AccionProponer, ActorID: "t", ValorNuevo: entero(60), RegistradaEn: ahora}},
	}
	registro, err := RestaurarRegistroMinutos(datos)
	if err != nil || registro.Estado() != MinutosPropuestos {
		t.Fatalf("restaurar inesperado: %+v err=%v", registro, err)
	}
	if _, err := RestaurarRegistroMinutos(DatosRegistroMinutos{}); err == nil {
		t.Fatal("datos vacíos deben rechazarse")
	}
	datos.Estado = MinutosVerificados
	if _, err := RestaurarRegistroMinutos(datos); err == nil {
		t.Fatal("un verificado sin valor debe rechazarse")
	}
	if _, err := RestaurarRegistroMinutos(DatosRegistroMinutos{
		SesionID: "s", ReservaID: "r", TutorID: "t", EscolarID: "e", FinSesion: ahora,
		Estado: MinutosPropuestos, Propuestos: 60,
	}); err == nil {
		t.Fatal("una auditoría vacía debe rechazarse")
	}
}
