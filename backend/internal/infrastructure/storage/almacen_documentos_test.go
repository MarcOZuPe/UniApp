package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAlmacenDocumentosFSGuardaFueraDelWebrootYValidaID(t *testing.T) {
	directorio := filepath.Join(t.TempDir(), "privado")
	almacen := AlmacenDocumentosFS{Directorio: directorio}
	id := "123e4567-e89b-42d3-a456-426614174000"
	contenido := []byte("%PDF-1.4\ncontenido")

	if err := almacen.Guardar(context.Background(), id, contenido); err != nil {
		t.Fatal(err)
	}
	guardado, err := almacen.Leer(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if string(guardado) != string(contenido) {
		t.Fatalf("contenido leído distinto: %q", guardado)
	}
	if _, err := os.Stat(filepath.Join(directorio, id+".pdf")); err != nil {
		t.Fatalf("documento no quedó en el directorio privado: %v", err)
	}
	if _, err := almacen.Leer(context.Background(), ".."); err == nil {
		t.Fatal("debe rechazarse un identificador que podría salir del directorio")
	}
	if err := almacen.Eliminar(context.Background(), id); err != nil {
		t.Fatal(err)
	}
}
