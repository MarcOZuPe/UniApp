package tutor

import (
	"context"
	"testing"
	memoria "uniapp/internal/adapters/repositories/memory/tutor"
)

func TestSembrarExpedienteSinteticoEsIdempotente(t *testing.T) {
	ctx := context.Background()
	almacen := nuevoAlmacenMemoria()
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria(), almacen)

	if err := SembrarExpedienteSintetico(ctx, servicio, "tutor-sintetico"); err != nil {
		t.Fatal(err)
	}
	primero, err := servicio.MiExpediente(ctx, "tutor-sintetico")
	if err != nil {
		t.Fatal(err)
	}
	if primero.Estado != EstadoPendiente || len(primero.Documentos) != 1 {
		t.Fatalf("expediente sintético inválido: %+v", primero)
	}
	if len(almacen.documentos) != 1 {
		t.Fatalf("documentos almacenados: %d", len(almacen.documentos))
	}

	if err := SembrarExpedienteSintetico(ctx, servicio, "tutor-sintetico"); err != nil {
		t.Fatal(err)
	}
	segundo, err := servicio.MiExpediente(ctx, "tutor-sintetico")
	if err != nil {
		t.Fatal(err)
	}
	if len(segundo.Documentos) != 1 || len(almacen.documentos) != 1 {
		t.Fatalf("la semilla duplicó documentos: expediente=%d almacén=%d", len(segundo.Documentos), len(almacen.documentos))
	}
}

func TestSembrarExpedienteSinteticoReparaPdfInvalido(t *testing.T) {
	ctx := context.Background()
	almacen := nuevoAlmacenMemoria()
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria(), almacen)

	datos := datosSinteticos()
	datos.Documentos[0].Contenido = []byte("%PDF-1.4\n% Documento sintético de UniApp\n")
	anterior, err := servicio.Presentar(ctx, "tutor-sintetico", datos)
	if err != nil {
		t.Fatal(err)
	}
	idAnterior := anterior.Documentos[0].ID

	if err := SembrarExpedienteSintetico(ctx, servicio, "tutor-sintetico"); err != nil {
		t.Fatal(err)
	}
	actual, err := servicio.MiExpediente(ctx, "tutor-sintetico")
	if err != nil {
		t.Fatal(err)
	}
	if len(actual.Documentos) != 1 || actual.Documentos[0].ID == idAnterior {
		t.Fatalf("la semilla no reemplazó el PDF inválido: %+v", actual.Documentos)
	}
	contenido, err := almacen.Leer(ctx, actual.Documentos[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if !pdfSinteticoValido(contenido) {
		t.Fatal("el PDF reparado no es válido")
	}
	if _, existe := almacen.documentos[idAnterior]; existe {
		t.Fatal("el PDF inválido anterior debe eliminarse")
	}
}

func TestPdfSinteticoEsValido(t *testing.T) {
	contenido := pdfSintetico()
	if !pdfSinteticoValido(contenido) {
		t.Fatal("el PDF sintético no tiene encabezado, xref o EOF válidos")
	}
	if pdfSinteticoValido([]byte("%PDF-1.4\n% Documento sintético de UniApp\n")) {
		t.Fatal("el PDF incompleto anterior no debe considerarse válido")
	}
}

func TestSembrarExpedienteSinteticoNoReemplazaUnExpedienteExistente(t *testing.T) {
	ctx := context.Background()
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria(), nuevoAlmacenMemoria())
	datos := datosValidos()
	datos.Presentacion = "Presentación propia del tutor"
	if _, err := servicio.Presentar(ctx, "tutor-con-expediente", datos); err != nil {
		t.Fatal(err)
	}

	if err := SembrarExpedienteSintetico(ctx, servicio, "tutor-con-expediente"); err != nil {
		t.Fatal(err)
	}
	expediente, err := servicio.MiExpediente(ctx, "tutor-con-expediente")
	if err != nil {
		t.Fatal(err)
	}
	if expediente.Presentacion != "Presentación propia del tutor" {
		t.Fatalf("la semilla reemplazó un expediente existente: %q", expediente.Presentacion)
	}
}
