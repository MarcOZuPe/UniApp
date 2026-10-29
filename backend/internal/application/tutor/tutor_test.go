package tutor

import (
	"context"
	"errors"
	"testing"
	"time"
	memoria "uniapp/internal/adapters/repositories/memory/tutor"
)

type almacenMemoria struct {
	documentos map[string][]byte
}

func nuevoAlmacenMemoria() *almacenMemoria {
	return &almacenMemoria{documentos: map[string][]byte{}}
}

func (a *almacenMemoria) Guardar(_ context.Context, id string, contenido []byte) error {
	a.documentos[id] = append([]byte(nil), contenido...)
	return nil
}

func (a *almacenMemoria) Leer(_ context.Context, id string) ([]byte, error) {
	contenido, ok := a.documentos[id]
	if !ok {
		return nil, ErrNoEncontrado
	}
	return append([]byte(nil), contenido...), nil
}

func (a *almacenMemoria) Eliminar(_ context.Context, id string) error {
	delete(a.documentos, id)
	return nil
}

func TestExpedientePendienteNoSePublicaHastaAprobacionVigente(t *testing.T) {
	ahora := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	repositorio := memoria.NuevoRepositorioMemoria()
	servicio := NuevoServicio(repositorio, nuevoAlmacenMemoria())
	servicio.ahora = func() time.Time { return ahora }

	expediente, err := servicio.Presentar(context.Background(), "tutor-1", datosValidos())
	if err != nil {
		t.Fatal(err)
	}
	if expediente.Estado != EstadoPendiente {
		t.Fatalf("estado inicial: %q", expediente.Estado)
	}
	if _, err := servicio.PerfilPublico(context.Background(), "tutor-1"); !errors.Is(err, ErrNoPublicable) {
		t.Fatalf("perfil pendiente no debe publicarse: %v", err)
	}

	vencimiento := ahora.Add(24 * time.Hour)
	aprobado, err := servicio.Revisar(context.Background(), "tutor-1", "coord-1", EstadoAprobado, "", vencimiento)
	if err != nil {
		t.Fatal(err)
	}
	if aprobado.VenceEn == nil || !aprobado.VenceEn.Equal(vencimiento) {
		t.Fatalf("vencimiento no persistido: %+v", aprobado.VenceEn)
	}
	publico, err := servicio.PerfilPublico(context.Background(), "tutor-1")
	if err != nil {
		t.Fatalf("perfil aprobado vigente debe publicarse: %v", err)
	}
	if len(publico.Documentos) != 0 {
		t.Fatal("el perfil público no debe exponer documentos")
	}
	perfiles, err := servicio.PerfilesPublicos(context.Background())
	if err != nil || len(perfiles) != 1 || perfiles[0].TutorID != "tutor-1" {
		t.Fatalf("el directorio debe incluir perfiles aprobados vigentes: %+v, error: %v", perfiles, err)
	}

	ahora = vencimiento
	if _, err := servicio.PerfilPublico(context.Background(), "tutor-1"); !errors.Is(err, ErrNoPublicable) {
		t.Fatalf("perfil vencido no debe publicarse: %v", err)
	}
	perfiles, err = servicio.PerfilesPublicos(context.Background())
	if err != nil || len(perfiles) != 0 {
		t.Fatalf("el directorio debe ocultar perfiles vencidos: %+v, error: %v", perfiles, err)
	}
}

func TestRechazoExigeMotivoYExpedienteSoloSeRevisaUnaVez(t *testing.T) {
	ahora := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	expediente, err := NuevoExpediente("tutor-1", datosValidos(), ahora)
	if err != nil {
		t.Fatal(err)
	}
	if err := expediente.Revisar(EstadoRechazado, "coord-1", "", time.Time{}, ahora); !errors.Is(err, ErrExpedienteInvalido) {
		t.Fatalf("rechazo sin motivo debe fallar: %v", err)
	}
	if err := expediente.Revisar(EstadoRechazado, "coord-1", "Falta sustento", time.Time{}, ahora); err != nil {
		t.Fatal(err)
	}
	if expediente.Motivo != "Falta sustento" || expediente.VenceEn != nil {
		t.Fatalf("estado rechazado inválido: %+v", expediente)
	}
	if err := expediente.Revisar(EstadoAprobado, "coord-1", "", ahora.Add(time.Hour), ahora); !errors.Is(err, ErrEstadoInvalido) {
		t.Fatalf("no debe permitirse una segunda revisión: %v", err)
	}
}

func TestDocumentoPrivadoSoloParaPropietarioOCoordinacion(t *testing.T) {
	almacen := nuevoAlmacenMemoria()
	servicio := NuevoServicio(memoria.NuevoRepositorioMemoria(), almacen)
	expediente, err := servicio.Presentar(context.Background(), "tutor-1", datosValidos())
	if err != nil {
		t.Fatal(err)
	}
	id := expediente.Documentos[0].ID
	if _, _, err := servicio.DescargarDocumento(context.Background(), id, "tutor-2", false); !errors.Is(err, ErrSinPermiso) {
		t.Fatalf("otro tutor no debe descargar el documento: %v", err)
	}
	if _, _, err := servicio.DescargarDocumento(context.Background(), id, "coord-1", true); err != nil {
		t.Fatalf("coordinación debe poder descargar el documento: %v", err)
	}
	if _, _, err := servicio.DescargarDocumento(context.Background(), id, "tutor-1", false); err != nil {
		t.Fatalf("el tutor dueño debe poder descargar el documento: %v", err)
	}
}

func TestPresentarRechazaArchivosQueNoSonPDF(t *testing.T) {
	datos := datosValidos()
	datos.Documentos[0].Contenido = []byte("no es un pdf")
	if _, err := NuevoExpediente("tutor-1", datos, time.Now()); err == nil {
		t.Fatal("debe rechazarse contenido que no tiene firma PDF")
	}
	datos = datosValidos()
	datos.Documentos[0].Nombre = "documento.txt"
	if _, err := NuevoExpediente("tutor-1", datos, time.Now()); err == nil {
		t.Fatal("debe rechazarse una extensión que no es PDF")
	}
}

func datosValidos() DatosExpediente {
	return DatosExpediente{
		Presentacion: "Presentación breve del tutor",
		Cursos:       "Álgebra y geometría",
		Experiencia:  "Experiencia de tutoría sintética",
		Documentos: []Archivo{{
			Nombre: "constancia.pdf", Contenido: []byte("%PDF-1.4\ncontenido sintético"),
		}},
	}
}
