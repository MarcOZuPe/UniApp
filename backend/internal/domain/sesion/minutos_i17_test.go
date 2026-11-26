package sesion

import (
	"errors"
	"strings"
	"testing"
	"time"
)

const coordinadorI17 = "coordinador-001"

// sesionConAsistenciaI17 devuelve una sesion en la que tutor y escolar
// declararon asistencia, y el momento justo despues de su fin.
func sesionConAsistenciaI17(t *testing.T) (*Sesion, time.Time) {
	t.Helper()
	s := nuevaSesionPrueba(t)
	for _, id := range []string{"tutor-001", "escolar-001"} {
		if err := s.DeclararAsistencia(id, s.Inicio()); err != nil {
			t.Fatal(err)
		}
	}
	return s, s.Fin().Add(time.Minute)
}

func registroPropuestoI17(t *testing.T, minutos int) (*RegistroMinutos, time.Time) {
	t.Helper()
	s, ahora := sesionConAsistenciaI17(t)
	r, err := ProponerMinutos(s, "tutor-001", minutos, ahora)
	if err != nil {
		t.Fatal(err)
	}
	return r, ahora
}

func esperarErrorI17(t *testing.T, err, esperado error) {
	t.Helper()
	if !errors.Is(err, esperado) {
		t.Fatalf("se esperaba %v, se obtuvo %v", esperado, err)
	}
}

// "45 minutos suman 45": se suma lo dictado, no la hora completa.
func TestI17CuarentaYCincoMinutosSuman45(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	if r.MinutosComputables() != 0 {
		t.Fatal("lo propuesto no debe sumar antes de verificarse")
	}
	if err := r.Aprobar(coordinadorI17, ahora); err != nil {
		t.Fatal(err)
	}
	if r.Estado() != MinutosVerificados || r.MinutosComputables() != 45 {
		t.Fatalf("esperado verificado con 45, se obtuvo %s con %d", r.Estado(), r.MinutosComputables())
	}
}

// "No autoaprueba": ni el tutor ni el escolar pueden verificar.
func TestI17NoAutoaprueba(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	esperarErrorI17(t, r.Aprobar("tutor-001", ahora), ErrAutoaprobacion)
	esperarErrorI17(t, r.Aprobar("escolar-001", ahora), ErrAutoaprobacion)
	esperarErrorI17(t, r.Ajustar("tutor-001", 60, "subir", ahora), ErrAutoaprobacion)
	if r.Estado() != MinutosPropuestos || len(r.Auditoria()) != 1 {
		t.Fatal("un intento rechazado no debe cambiar el registro")
	}
}

// "Una disputa congela": mientras esta en disputa suma 0.
func TestI17DisputaCongela(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 50)
	if err := r.Aprobar(coordinadorI17, ahora); err != nil {
		t.Fatal(err)
	}
	if err := r.Disputar("escolar-001", "La sesion duro 30 minutos", ahora); err != nil {
		t.Fatal(err)
	}
	if r.Estado() != MinutosEnDisputa || r.MinutosComputables() != 0 {
		t.Fatal("la disputa debe congelar los minutos en 0")
	}
	esperarErrorI17(t, r.Aprobar(coordinadorI17, ahora), ErrTransicionMinutos)
	esperarErrorI17(t, r.Ajustar(coordinadorI17, 40, "x", ahora), ErrTransicionMinutos)
	if err := r.Resolver(coordinadorI17, 30, "Grabacion revisada", ahora); err != nil {
		t.Fatal(err)
	}
	if r.Estado() != MinutosVerificados || r.MinutosComputables() != 30 {
		t.Fatal("la resolucion debe fijar los minutos finales")
	}
}

// "Una cancelacion suma cero", en cualquier estado e idempotente.
func TestI17CancelacionSumaCero(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 60)
	if err := r.Aprobar(coordinadorI17, ahora); err != nil {
		t.Fatal(err)
	}
	if err := r.Anular(ahora); err != nil {
		t.Fatal(err)
	}
	if err := r.Anular(ahora); err != nil {
		t.Fatal(err)
	}
	if r.Estado() != MinutosAnulados || r.MinutosComputables() != 0 {
		t.Fatal("una cancelacion debe sumar cero")
	}
	if n := len(r.Auditoria()); n != 3 {
		t.Fatalf("anular dos veces no debe duplicar la auditoria: %d entradas", n)
	}
	esperarErrorI17(t, r.Ajustar(coordinadorI17, 10, "x", ahora), ErrSesionCancelada)
	esperarErrorI17(t, r.Disputar("escolar-001", "x", ahora), ErrSesionCancelada)
}

// "Ajustes con auditoria": quien, cuando, antes, despues y por que.
func TestI17AjusteAuditado(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	esperarErrorI17(t, r.Ajustar(coordinadorI17, 40, "   ", ahora), ErrMotivoInvalido)
	esperarErrorI17(t, r.Ajustar(coordinadorI17, 40, strings.Repeat("a", 501), ahora), ErrMotivoInvalido)
	if err := r.Aprobar(coordinadorI17, ahora); err != nil {
		t.Fatal(err)
	}
	if err := r.Ajustar(coordinadorI17, 40, "  Inicio con 5 minutos de retraso  ", ahora.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	auditoria := r.Auditoria()
	if len(auditoria) != 3 {
		t.Fatalf("esperadas 3 entradas, hay %d", len(auditoria))
	}
	ajuste := auditoria[2]
	anterior, okA := ajuste.ValorAnterior()
	nuevo, okN := ajuste.ValorNuevo()
	if ajuste.Accion() != AccionAjustar || ajuste.ActorID() != coordinadorI17 ||
		!okA || anterior != 45 || !okN || nuevo != 40 ||
		ajuste.Motivo() != "Inicio con 5 minutos de retraso" || !ajuste.RegistradaEn().Equal(ahora.Add(time.Hour)) {
		t.Fatalf("auditoria del ajuste incorrecta: %+v", ajuste)
	}
	// La copia no permite reescribir la historia.
	auditoria[0] = EntradaAuditoria{}
	if r.Auditoria()[0].Accion() != AccionProponer {
		t.Fatal("la auditoria interna fue modificada desde fuera")
	}
}

func TestI17ProponerExigeReglas(t *testing.T) {
	s := nuevaSesionPrueba(t)
	despues := s.Fin().Add(time.Minute)

	// Sin asistencia de ambos: un enlace abierto no acredita.
	esperarErrorI17(t, errDe(ProponerMinutos(s, "tutor-001", 45, despues)), ErrAsistenciaIncompleta)
	if err := s.DeclararAsistencia("tutor-001", s.Inicio()); err != nil {
		t.Fatal(err)
	}
	esperarErrorI17(t, errDe(ProponerMinutos(s, "tutor-001", 45, despues)), ErrAsistenciaIncompleta)
	if err := s.DeclararAsistencia("escolar-001", s.Inicio()); err != nil {
		t.Fatal(err)
	}

	esperarErrorI17(t, errDe(ProponerMinutos(s, "escolar-001", 45, despues)), ErrNoAutorizado)
	esperarErrorI17(t, errDe(ProponerMinutos(s, "ajeno", 45, despues)), ErrNoAutorizado)
	esperarErrorI17(t, errDe(ProponerMinutos(s, "tutor-001", 45, s.Fin().Add(-time.Minute))), ErrSesionNoFinalizada)
	esperarErrorI17(t, errDe(ProponerMinutos(s, "tutor-001", 45, s.Fin().Add(PlazoResultado+time.Second))), ErrFueraDePlazo)
	for _, m := range []int{0, -5, 61} {
		esperarErrorI17(t, errDe(ProponerMinutos(s, "tutor-001", m, despues)), ErrMinutosInvalidos)
	}
	r, err := ProponerMinutos(s, "tutor-001", 60, s.Fin())
	if err != nil {
		t.Fatalf("en el fin exacto debe poder proponer: %v", err)
	}
	if r.Estado() != MinutosPropuestos || r.MinutosPropuestos() != 60 || r.TutorID() != "tutor-001" {
		t.Fatal("registro propuesto incorrecto")
	}
}

func TestI17ReglasDeDisputa(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 50)
	// El tutor no disputa lo que el mismo propuso sin ajuste.
	esperarErrorI17(t, r.Disputar("tutor-001", "x", ahora), ErrDisputaNoPermitida)
	esperarErrorI17(t, r.Disputar("ajeno", "x", ahora), ErrNoAutorizado)
	esperarErrorI17(t, r.Disputar("escolar-001", "", ahora), ErrMotivoInvalido)
	esperarErrorI17(t, r.Disputar("escolar-001", "x", r.FinSesion().Add(PlazoResultado+time.Second)), ErrFueraDePlazo)

	// Un ajuste a la baja si lo puede disputar el tutor.
	if err := r.Ajustar(coordinadorI17, 30, "Asistencia parcial", ahora); err != nil {
		t.Fatal(err)
	}
	if err := r.Disputar("tutor-001", "Dicte los 50 minutos", ahora); err != nil {
		t.Fatal(err)
	}
	if r.DisputadoPor() != "tutor-001" {
		t.Fatal("debe registrar quien disputa")
	}
	if err := r.Resolver(coordinadorI17, 45, "Se acuerda 45", ahora); err != nil {
		t.Fatal(err)
	}
	// La resolucion es final.
	esperarErrorI17(t, r.Disputar("escolar-001", "otra vez", ahora), ErrDisputaNoPermitida)
	esperarErrorI17(t, r.Disputar("tutor-001", "otra vez", ahora), ErrDisputaNoPermitida)
	esperarErrorI17(t, r.Resolver(coordinadorI17, 50, "x", ahora), ErrTransicionMinutos)
}

func TestI17RestaurarConservaInvariantes(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	if err := r.Ajustar(coordinadorI17, 40, "retraso", ahora); err != nil {
		t.Fatal(err)
	}
	datos := DatosRegistroMinutos{
		SesionID: r.SesionID(), ReservaID: r.ReservaID(), TutorID: r.TutorID(), EscolarID: r.EscolarID(),
		FinSesion: r.FinSesion(), Estado: r.Estado(), Propuestos: r.MinutosPropuestos(),
	}
	if v, ok := r.MinutosVerificados(); ok {
		datos.Verificados = &v
	}
	for _, e := range r.Auditoria() {
		a := DatosEntradaAuditoria{Accion: e.Accion(), ActorID: e.ActorID(), Motivo: e.Motivo(), RegistradaEn: e.RegistradaEn()}
		if v, ok := e.ValorAnterior(); ok {
			a.ValorAnterior = &v
		}
		if v, ok := e.ValorNuevo(); ok {
			a.ValorNuevo = &v
		}
		datos.Auditoria = append(datos.Auditoria, a)
	}
	restaurado, err := RestaurarRegistroMinutos(datos)
	if err != nil {
		t.Fatal(err)
	}
	if restaurado.MinutosComputables() != 40 || len(restaurado.Auditoria()) != 2 {
		t.Fatal("restauracion incorrecta")
	}

	malos := []func(*DatosRegistroMinutos){
		func(d *DatosRegistroMinutos) { d.Estado = "inventado" },
		func(d *DatosRegistroMinutos) { d.Propuestos = 0 },
		func(d *DatosRegistroMinutos) { d.Verificados = nil },
		func(d *DatosRegistroMinutos) { d.Auditoria = nil },
		func(d *DatosRegistroMinutos) { d.SesionID = "" },
	}
	for i, alterar := range malos {
		copia := datos
		copia.Auditoria = append([]DatosEntradaAuditoria(nil), datos.Auditoria...)
		alterar(&copia)
		if _, err := RestaurarRegistroMinutos(copia); err == nil {
			t.Fatalf("caso %d: se esperaba rechazo", i)
		}
	}
}

func TestI17ClonarEsIndependiente(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	copia := r.Clonar()
	if err := copia.Aprobar(coordinadorI17, ahora); err != nil {
		t.Fatal(err)
	}
	if r.Estado() != MinutosPropuestos || len(r.Auditoria()) != 1 {
		t.Fatal("modificar el clon altero el original")
	}
}

func errDe(_ *RegistroMinutos, err error) error { return err }
