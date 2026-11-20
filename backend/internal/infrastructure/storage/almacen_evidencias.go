package storage

import (
	"context"
	"fmt"
	"os"
)

// AlmacenEvidenciasFS reutiliza el almacén privado de adjuntos con un directorio propio.
// Las evidencias de incidentes son sensibles: se guardan con nombre opaco y permisos
// 0700/0600, y solo se sirven a través del controlador autorizado.
type AlmacenEvidenciasFS struct {
	Directorio string
}

func (a AlmacenEvidenciasFS) Guardar(ctx context.Context, clave string, contenido []byte) error {
	if err := os.MkdirAll(a.Directorio, 0700); err != nil {
		return fmt.Errorf("crear directorio de evidencias: %w", err)
	}
	if err := os.Chmod(a.Directorio, 0700); err != nil {
		return fmt.Errorf("proteger directorio de evidencias: %w", err)
	}
	return (AlmacenAdjuntosFS{Directorio: a.Directorio}).Guardar(ctx, clave, contenido)
}

func (a AlmacenEvidenciasFS) Leer(ctx context.Context, clave string) ([]byte, error) {
	return (AlmacenAdjuntosFS{Directorio: a.Directorio}).Leer(ctx, clave)
}

func (a AlmacenEvidenciasFS) Eliminar(ctx context.Context, clave string) error {
	return (AlmacenAdjuntosFS{Directorio: a.Directorio}).Eliminar(ctx, clave)
}
