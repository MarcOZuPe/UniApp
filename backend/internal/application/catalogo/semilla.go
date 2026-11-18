package catalogo

import (
	"context"
	"strings"
)

const FuenteSinteticaInicial = "sintetico: fixture inicial I07"

// SembrarAprobacionSintetica deja un curso aprobado para que el catálogo público
// tenga una muestra visible desde el inicio. Solo actúa sobre la importación
// sintética inicial y cuando ningún curso fue revisado todavía; así no revierte
// decisiones de coordinación ni afecta importaciones reales.
func SembrarAprobacionSintetica(ctx context.Context, servicio *Servicio, coordinadorID string) error {
	if servicio == nil || strings.TrimSpace(coordinadorID) == "" {
		return nil
	}
	importacion, err := servicio.repositorio.ImportacionActual(ctx)
	if err != nil {
		return err
	}
	if importacion.Fuente != FuenteSinteticaInicial {
		return nil
	}
	for _, curso := range importacion.Cursos {
		if curso.Estado != EstadoPendiente {
			return nil
		}
	}
	for _, curso := range importacion.Cursos {
		if curso.Nombre == "Matemática básica" && curso.Grado == "1.° de secundaria" {
			_, err := servicio.RevisarCurso(ctx, curso.ID, coordinadorID, EstadoAprobado, "")
			return err
		}
	}
	return nil
}
