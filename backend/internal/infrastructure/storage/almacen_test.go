package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAlmacenDocumentosFS(t *testing.T) {
	directorio := t.TempDir()
	almacen := AlmacenDocumentosFS{Directorio: directorio}
	ctx := context.Background()
	id := "123e4567-e89b-12d3-a456-426614174000"
	if err := almacen.Guardar(ctx, id, []byte("%PDF-1.4")); err != nil {
		t.Fatalf("guardar falló: %v", err)
	}
	contenido, err := almacen.Leer(ctx, id)
	if err != nil || string(contenido) != "%PDF-1.4" {
		t.Fatalf("leer inesperado: %q err=%v", contenido, err)
	}
	if err := almacen.Guardar(ctx, id, []byte("otra")); err == nil {
		t.Fatal("no se debe sobrescribir un documento existente")
	}
	if err := almacen.Eliminar(ctx, id); err != nil {
		t.Fatalf("eliminar falló: %v", err)
	}
	if err := almacen.Eliminar(ctx, id); err != nil {
		t.Fatalf("eliminar de nuevo no debe fallar: %v", err)
	}
	if _, err := almacen.Leer(ctx, id); err == nil {
		t.Fatal("un documento eliminado no debe leerse")
	}
	for _, invalido := range []string{"", "no-uuid", "123e4567-e89b-12d3-a456-42661417400z"} {
		if _, err := almacen.Leer(ctx, invalido); err == nil {
			t.Fatalf("el ID %q debe rechazarse", invalido)
		}
	}
	cancelado, cancelar := context.WithCancel(ctx)
	cancelar()
	if err := almacen.Guardar(cancelado, id, []byte("x")); err == nil {
		t.Fatal("un contexto cancelado debe rechazarse")
	}
}

func TestAlmacenAdjuntosYEvidencias(t *testing.T) {
	almacen := AlmacenAdjuntosFS{Directorio: t.TempDir()}
	ctx := context.Background()
	clave := "0123456789abcdef0123456789abcdef"
	if err := almacen.Guardar(ctx, clave, []byte("adjunto")); err != nil {
		t.Fatalf("guardar adjunto falló: %v", err)
	}
	contenido, err := almacen.Leer(ctx, clave)
	if err != nil || string(contenido) != "adjunto" {
		t.Fatalf("leer adjunto inesperado: %q err=%v", contenido, err)
	}
	if err := almacen.Eliminar(ctx, clave); err != nil {
		t.Fatal(err)
	}
	if err := almacen.Eliminar(ctx, clave); err != nil {
		t.Fatalf("eliminar dos veces no debe fallar: %v", err)
	}
	if err := almacen.Guardar(ctx, "clave-mala", []byte("x")); err == nil {
		t.Fatal("una clave inválida debe rechazarse")
	}
	if _, err := (AlmacenAdjuntosFS{}).Leer(ctx, clave); err == nil {
		t.Fatal("un directorio vacío debe rechazarse")
	}

	evidencias := AlmacenEvidenciasFS{Directorio: t.TempDir()}
	if err := evidencias.Guardar(ctx, clave, []byte("evidencia")); err != nil {
		t.Fatalf("guardar evidencia falló: %v", err)
	}
	if contenido, err := evidencias.Leer(ctx, clave); err != nil || string(contenido) != "evidencia" {
		t.Fatalf("leer evidencia inesperado: %q err=%v", contenido, err)
	}
	if err := evidencias.Eliminar(ctx, clave); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(directorioInexistente(), clave)); err == nil {
		t.Fatal("la evidencia no debe existir fuera del directorio")
	}
}

func directorioInexistente() string { return "/tmp/no-existe-uniapp" }
