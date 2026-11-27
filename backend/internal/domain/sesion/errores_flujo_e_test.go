package sesion

import (
	"errors"
	"testing"
)

// Pruebas de errores del flujo E (I17) que fallan a propósito: cada una
// demuestra un bug real encontrado al revisar el código. Siguen el formato de
// las F-xx de #66 y la clasificación de "Software Errors" (IntroTesting).

// F-26 (error funcional / de flujo de control): después de Resolver una
// disputa, coordinación todavía puede Ajustar el valor, pero ese ajuste nuevo
// ya no se puede disputar porque fueResuelto() bloquea para siempre. El
// "final" de la resolución debería cubrir solo el valor resuelto, no los
// ajustes posteriores.
func TestF26AjusteTrasResolucionDebeSerDisputable(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	if err := r.Disputar("escolar-001", "empezamos tarde", ahora); err != nil {
		t.Fatal(err)
	}
	if err := r.Resolver(coordinadorI17, 40, "revisado con ambos", ahora); err != nil {
		t.Fatal(err)
	}
	if err := r.Ajustar(coordinadorI17, 10, "corrección posterior", ahora); err != nil {
		t.Fatalf("el ajuste posterior está permitido hoy: %v", err)
	}
	if err := r.Disputar("tutor-001", "el ajuste a 10 no tiene sustento", ahora); err != nil {
		t.Fatalf("F-26: un ajuste posterior a la resolución quedó sin derecho a disputa: %v", err)
	}
}

// F-27 (error de límites / validación de entrada): un motivo con el byte NUL
// es UTF-8 válido y pasa normalizarMotivo, pero PostgreSQL rechaza \x00 en
// columnas TEXT. El repositorio falla y la API responde 500 en lugar de 422.
func TestF27MotivoConCaracterNULDebeRechazarse(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	err := r.Ajustar(coordinadorI17, 30, "motivo\x00oculto", ahora)
	if !errors.Is(err, ErrMotivoInvalido) {
		t.Fatalf("F-27: un motivo con NUL debe rechazarse con ErrMotivoInvalido (422), se obtuvo %v", err)
	}
}

// F-29 (error de auditoría / datos): Ajustar al mismo valor ya verificado
// agrega una entrada "ajustar" sin ningún cambio. Ensucia la bitácora y hace
// parecer que hubo una corrección.
func TestF29AjusteSinCambioNoDebeAuditarse(t *testing.T) {
	r, ahora := registroPropuestoI17(t, 45)
	if err := r.Aprobar(coordinadorI17, ahora); err != nil {
		t.Fatal(err)
	}
	antes := len(r.Auditoria())
	err := r.Ajustar(coordinadorI17, 45, "mismo valor", ahora)
	if err == nil && len(r.Auditoria()) != antes {
		t.Fatalf("F-29: ajustar a 45 cuando ya vale 45 creó %d entradas de auditoría sin cambio", len(r.Auditoria())-antes)
	}
}

// F-30 (error de manejo de estado): RestaurarRegistroMinutos acepta datos
// persistidos inconsistentes. Una fila corrupta se reconstituye como si fuera
// válida y el agregado queda en un estado imposible.
func TestF30RestaurarDebeRechazarEstadosInconsistentes(t *testing.T) {
	base := func() DatosRegistroMinutos {
		r, _ := registroPropuestoI17(t, 45)
		a := r.Auditoria()[0]
		return DatosRegistroMinutos{
			SesionID: r.SesionID(), ReservaID: r.ReservaID(), TutorID: r.TutorID(), EscolarID: r.EscolarID(),
			FinSesion: r.FinSesion(), Estado: MinutosPropuestos, Propuestos: 45,
			Auditoria: []DatosEntradaAuditoria{{Accion: a.Accion(), ActorID: a.ActorID(), RegistradaEn: a.RegistradaEn()}},
		}
	}

	enDisputaSinAutor := base()
	enDisputaSinAutor.Estado = MinutosEnDisputa
	if _, err := RestaurarRegistroMinutos(enDisputaSinAutor); err == nil {
		t.Fatal("F-30: se restauró un registro en_disputa sin disputadoPor")
	}

	propuestoConVerificados := base()
	verificados := 30
	propuestoConVerificados.Verificados = &verificados
	if _, err := RestaurarRegistroMinutos(propuestoConVerificados); err == nil {
		t.Fatal("F-30: se restauró un registro propuesto que ya tiene minutos verificados")
	}
}
