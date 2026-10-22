package adapters

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AlmacenAdjuntosFS guarda los adjuntos del flujo D fuera de la base de datos, con
// permisos privados y un nombre opaco. El modelo solo conoce el puerto Almacen.
type AlmacenAdjuntosFS struct {
	Directorio string
}

func (a AlmacenAdjuntosFS) Guardar(ctx context.Context, clave string, contenido []byte) error {
	ruta, err := a.ruta(clave)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(a.Directorio, 0700); err != nil {
		return fmt.Errorf("crear directorio de adjuntos: %w", err)
	}
	archivo, err := os.OpenFile(ruta, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("crear adjunto: %w", err)
	}
	if _, err := archivo.Write(contenido); err != nil {
		_ = archivo.Close()
		_ = os.Remove(ruta)
		return fmt.Errorf("guardar adjunto: %w", err)
	}
	if err := archivo.Close(); err != nil {
		_ = os.Remove(ruta)
		return fmt.Errorf("cerrar adjunto: %w", err)
	}
	return nil
}

func (a AlmacenAdjuntosFS) Leer(ctx context.Context, clave string) ([]byte, error) {
	ruta, err := a.ruta(clave)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("leer adjunto: %w", err)
	}
	return contenido, nil
}

func (a AlmacenAdjuntosFS) Eliminar(ctx context.Context, clave string) error {
	ruta, err := a.ruta(clave)
	if err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.Remove(ruta); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("eliminar adjunto: %w", err)
	}
	return nil
}

// ruta exige una clave hexadecimal de 32 caracteres: así el nombre del archivo no puede
// escapar del directorio privado.
func (a AlmacenAdjuntosFS) ruta(clave string) (string, error) {
	if a.Directorio == "" || len(clave) != 32 {
		return "", fmt.Errorf("clave o directorio de adjunto inválido")
	}
	for _, caracter := range clave {
		if !strings.ContainsRune("0123456789abcdef", caracter) {
			return "", fmt.Errorf("clave de adjunto inválida")
		}
	}
	return filepath.Join(a.Directorio, clave), nil
}
