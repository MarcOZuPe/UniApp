package tutor

import (
	"context"
	"errors"
	"testing"
	"time"

	memoria "uniapp/internal/adapters/repositories/memory/tutor"
)

type almacenPrueba struct{ datos map[string][]byte }

func (a *almacenPrueba) Guardar(_ context.Context, clave string, contenido []byte) error {
	if a.datos == nil {
		a.datos = map[string][]byte{}
	}
	a.datos[clave] = append([]byte(nil), contenido...)
	return nil
}
func (a *almacenPrueba) Leer(_ context.Context, clave string) ([]byte, error) {
	contenido, existe := a.datos[clave]
	if !existe {
		return nil, ErrNoEncontrado
	}
	return contenido, nil
}
func (a *almacenPrueba) Eliminar(_ context.Context, clave string) error {
	delete(a.datos, clave)
	return nil
}

func datosTutor() DatosExpediente {
	return DatosExpediente{
		Presentacion: "Presentación válida", Cursos: "Álgebra", Experiencia: "Experiencia válida",
		Documentos: []Archivo{{Nombre: "titulo.pdf", Contenido: []byte("%PDF-1.4 contenido")}},
	}
}

func TestServicioTutorFlujoCompleto(t *testing.T) {
	ctx := context.Background()
	almacen := &almacenPrueba{}
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria(), almacen)

	expediente, err := servicio.Presentar(ctx, "tutor-1", datosTutor())
	if err != nil {
		t.Fatalf("presentar falló: %v", err)
	}
	if expediente.TutorID != "tutor-1" || expediente.Estado != EstadoPendiente || len(expediente.Documentos) != 1 {
		t.Fatalf("expediente inesperado: %+v", expediente)
	}
	if len(almacen.datos) != 1 {
		t.Fatal("el documento debería quedar almacenado")
	}
	documentoID := expediente.Documentos[0].ID
	if _, err := servicio.MiExpediente(ctx, "tutor-1"); err != nil {
		t.Fatal(err)
	}
	pendientes, err := servicio.Pendientes(ctx, "1")
	if err != nil || len(pendientes) != 1 {
		t.Fatalf("pendientes inesperados: %+v err=%v", pendientes, err)
	}
	if _, err := servicio.Revisar(ctx, "tutor-1", "coord-1", "1", EstadoAprobado, "", time.Time{}); err == nil {
		t.Fatal("aprobar sin vencimiento debe fallar")
	}
	aprobado, err := servicio.Revisar(ctx, "tutor-1", "coord-1", "1", EstadoAprobado, "", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil || aprobado.Estado != EstadoAprobado {
		t.Fatalf("aprobar falló: %+v err=%v", aprobado, err)
	}
	if _, err := servicio.Revisar(ctx, "tutor-1", "coord-1", "2", EstadoAprobado, "", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, ErrSinPermiso) {
		t.Fatalf("otra facultad debe rechazarse: %v", err)
	}
	aprobados, err := servicio.Aprobados(ctx, "1")
	if err != nil || len(aprobados) != 1 {
		t.Fatalf("aprobados inesperados: %+v err=%v", aprobados, err)
	}
	publico, err := servicio.PerfilPublico(ctx, "tutor-1")
	if err != nil || len(publico.Documentos) != 0 {
		t.Fatalf("perfil público inesperado: %+v err=%v", publico, err)
	}
	perfiles, err := servicio.PerfilesPublicos(ctx)
	if err != nil || len(perfiles) != 1 {
		t.Fatalf("perfiles públicos inesperados: %+v err=%v", perfiles, err)
	}
	contenido, documento, err := servicio.DescargarDocumento(ctx, documentoID, "tutor-1", "", false)
	if err != nil || string(contenido) != "%PDF-1.4 contenido" || documento.ID != documentoID {
		t.Fatalf("descarga inesperada: %q %+v err=%v", contenido, documento, err)
	}
	if _, _, err := servicio.DescargarDocumento(ctx, documentoID, "otro", "", false); !errors.Is(err, ErrSinPermiso) {
		t.Fatalf("otro usuario no debe descargar: %v", err)
	}
	if _, _, err := servicio.DescargarDocumento(ctx, documentoID, "coord-1", "2", true); !errors.Is(err, ErrSinPermiso) {
		t.Fatalf("coordinador de otra facultad no debe descargar: %v", err)
	}
}

func TestServicioTutorRechazaSinExpediente(t *testing.T) {
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria(), &almacenPrueba{})
	ctx := context.Background()
	if _, err := servicio.MiExpediente(ctx, "tutor-1"); err == nil {
		t.Fatal("un tutor sin expediente debe fallar")
	}
	if _, err := servicio.PerfilPublico(ctx, "tutor-1"); err == nil {
		t.Fatal("un perfil inexistente no debe ser público")
	}
}
