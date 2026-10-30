package acceso

import (
	"context"
	"testing"
)

func TestSembrarVinculacionSinteticaEsIdempotente(testingT *testing.T) {
	contexto := context.Background()
	servicio, _, _, _ := nuevo(testingT)
	if err := Sembrar(contexto, servicio, DatosSiembra{DominioTutor: "uni.edu.pe", Clave: claveOK}); err != nil {
		testingT.Fatal(err)
	}

	if err := SembrarVinculacion(contexto, servicio, DatosSiembraVinculacion{EmailApoderado: EmailApoderadoSintetico, EmailEscolar: EmailEscolarSintetico}); err != nil {
		testingT.Fatal(err)
	}
	apoderado, err := servicio.RepositorioUsuarios.UsuarioPorEmail(contexto, EmailApoderadoSintetico)
	if err != nil {
		testingT.Fatal(err)
	}
	escolar, err := servicio.RepositorioUsuarios.UsuarioPorEmail(contexto, EmailEscolarSintetico)
	if err != nil {
		testingT.Fatal(err)
	}

	vinculacion, err := servicio.RepositorioVinculaciones.VinculacionAceptadaDeEscolar(contexto, escolar.ID)
	if err != nil || vinculacion.ApoderadoID != apoderado.ID {
		testingT.Fatalf("vínculo sintético: %+v (%v)", vinculacion, err)
	}
	historial, err := servicio.Autorizacion.HistorialAutorizaciones(contexto, solicitudAutorizacion(apoderado, escolar.ID))
	if err != nil || len(historial) != 1 || !historial[0].Vigente() {
		testingT.Fatalf("autorización inicial: %+v (%v)", historial, err)
	}

	if err := SembrarVinculacion(contexto, servicio, DatosSiembraVinculacion{EmailApoderado: EmailApoderadoSintetico, EmailEscolar: EmailEscolarSintetico}); err != nil {
		testingT.Fatal(err)
	}
	historial, err = servicio.Autorizacion.HistorialAutorizaciones(contexto, solicitudAutorizacion(apoderado, escolar.ID))
	if err != nil || len(historial) != 1 {
		testingT.Fatalf("semilla idempotente: %+v (%v)", historial, err)
	}
}
