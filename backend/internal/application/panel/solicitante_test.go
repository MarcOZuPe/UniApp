package panel

import (
	"testing"

	modeloAcceso "uniapp/internal/application/acceso"
	modelo "uniapp/internal/domain/panel"
)

func TestSolicitanteSoloAutorizaCoordinacionConFacultad(t *testing.T) {
	casos := []struct {
		nombre      string
		solicitante Solicitante
		autorizado  bool
	}{
		{"coordinador con facultad", Solicitante{Rol: modeloAcceso.RolCoordinador, FacultadID: "fac-1"}, true},
		{"coordinador general con facultad", Solicitante{Rol: modeloAcceso.RolCoordinadorGeneral, FacultadID: "fac-1"}, true},
		{"coordinador sin facultad", Solicitante{Rol: modeloAcceso.RolCoordinador}, false},
		{"rol ajeno", Solicitante{Rol: modeloAcceso.RolEscolar, FacultadID: "fac-1"}, false},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			if obtenido := caso.solicitante.Autorizado(); obtenido != caso.autorizado {
				t.Fatalf("Autorizado()=%v, se esperaba %v", obtenido, caso.autorizado)
			}
		})
	}
}

// B-21: una facultad con solo espacios no identifica ninguna facultad.
func TestB21FacultadEnBlancoNoDebeAutorizar(t *testing.T) {
	solicitante := Solicitante{Rol: modeloAcceso.RolCoordinador, FacultadID: "   "}
	if solicitante.Autorizado() {
		t.Fatal("B-21: una facultad en blanco no debe autorizar el panel")
	}
}

func TestFiltroAuditoriaNormalizaElLimite(t *testing.T) {
	casos := []struct {
		nombre   string
		limite   int
		esperado int
	}{
		{"cero usa el valor por defecto", 0, modelo.LimiteAuditoriaPorDefecto},
		{"negativo usa el valor por defecto", -5, modelo.LimiteAuditoriaPorDefecto},
		{"valor válido se conserva", 25, 25},
		{"exceso se recorta al máximo", modelo.LimiteAuditoriaMaximo + 1, modelo.LimiteAuditoriaMaximo},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			normalizado := modelo.FiltroAuditoria{Limite: caso.limite}.Normalizado()
			if normalizado.Limite != caso.esperado {
				t.Fatalf("Limite=%d, se esperaba %d", normalizado.Limite, caso.esperado)
			}
		})
	}
}
