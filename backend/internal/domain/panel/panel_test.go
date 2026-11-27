package panel

import "testing"

func TestFiltroAuditoriaNormalizado(t *testing.T) {
	casos := []struct {
		limite   int
		esperado int
	}{
		{0, LimiteAuditoriaPorDefecto},
		{-1, LimiteAuditoriaPorDefecto},
		{25, 25},
		{LimiteAuditoriaMaximo + 1, LimiteAuditoriaMaximo},
		{LimiteAuditoriaMaximo, LimiteAuditoriaMaximo},
	}
	for _, caso := range casos {
		normalizado := (FiltroAuditoria{Limite: caso.limite, FacultadID: "fac-1"}).Normalizado()
		if normalizado.Limite != caso.esperado || normalizado.FacultadID != "fac-1" {
			t.Fatalf("filtro inesperado para %d: %+v", caso.limite, normalizado)
		}
	}
}
