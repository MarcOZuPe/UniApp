package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type AlmacenDocumentosFS struct {
	Directorio string
}

func (a AlmacenDocumentosFS) Guardar(ctx context.Context, id string, contenido []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	ruta, err := a.ruta(id)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(a.Directorio, 0700); err != nil {
		return fmt.Errorf("crear directorio privado de documentos: %w", err)
	}
	if err := os.Chmod(a.Directorio, 0700); err != nil {
		return fmt.Errorf("proteger directorio privado de documentos: %w", err)
	}
	archivo, err := os.OpenFile(ruta, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return fmt.Errorf("crear documento privado: %w", err)
	}
	if _, err := archivo.Write(contenido); err != nil {
		_ = archivo.Close()
		_ = os.Remove(ruta)
		return fmt.Errorf("guardar documento privado: %w", err)
	}
	if err := archivo.Sync(); err != nil {
		_ = archivo.Close()
		_ = os.Remove(ruta)
		return fmt.Errorf("sincronizar documento privado: %w", err)
	}
	if err := archivo.Close(); err != nil {
		_ = os.Remove(ruta)
		return fmt.Errorf("cerrar documento privado: %w", err)
	}
	return nil
}

func (a AlmacenDocumentosFS) Leer(ctx context.Context, id string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ruta, err := a.ruta(id)
	if err != nil {
		return nil, err
	}
	contenido, err := os.ReadFile(ruta)
	if err != nil {
		return nil, fmt.Errorf("leer documento privado: %w", err)
	}
	return contenido, nil
}

func (a AlmacenDocumentosFS) Eliminar(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	ruta, err := a.ruta(id)
	if err != nil {
		return err
	}
	if err := os.Remove(ruta); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("eliminar documento privado: %w", err)
	}
	return nil
}

func (a AlmacenDocumentosFS) ruta(id string) (string, error) {
	if a.Directorio == "" || len(id) != 36 {
		return "", fmt.Errorf("identificador o directorio de documento inválido")
	}
	for indice, caracter := range id {
		if indice == 8 || indice == 13 || indice == 18 || indice == 23 {
			if caracter != '-' {
				return "", fmt.Errorf("identificador de documento inválido")
			}
			continue
		}
		if !strings.ContainsRune("0123456789abcdef", caracter) {
			return "", fmt.Errorf("identificador de documento inválido")
		}
	}
	return filepath.Join(a.Directorio, id+".pdf"), nil
}
