package catalogo

import (
	"strings"
	"testing"
	"time"
)

func TestNuevaImportacionValidaYNormaliza(t *testing.T) {
	ahora := time.Now().UTC()
	importacion, err := NuevaImportacion(DatosImportacion{
		Fuente:   "  padrón   escolar  ",
		Cursos:   []CursoEntrada{{Nombre: "  Álgebra  ", Grado: " 5to  "}},
		Colegios: []ColegioEntrada{{Nombre: "  Colegio Uno ", Distrito: " Lima  "}},
	}, ahora)
	if err != nil {
		t.Fatalf("importación válida rechazada: %v", err)
	}
	if importacion.Fuente != "padrón   escolar" || importacion.Cursos[0].Nombre != "Álgebra" ||
		importacion.Cursos[0].Estado != EstadoPendiente || importacion.Colegios[0].Distrito != "Lima" {
		t.Fatalf("importación no normalizada: %+v", importacion)
	}
}

func TestNuevaImportacionRechazaDatosInvalidos(t *testing.T) {
	casos := []DatosImportacion{
		{Fuente: "", Cursos: nil, Colegios: nil},
		{Fuente: "origen", Cursos: []CursoEntrada{{Nombre: "", Grado: "5to"}}, Colegios: []ColegioEntrada{{Nombre: "C", Distrito: "D"}}},
		{Fuente: "origen", Cursos: []CursoEntrada{{Nombre: "Álgebra", Grado: "5to"}, {Nombre: "Álgebra", Grado: "5to"}}, Colegios: []ColegioEntrada{{Nombre: "C", Distrito: "D"}}},
		{Fuente: "origen", Cursos: []CursoEntrada{{Nombre: "Álgebra", Grado: "5to"}}, Colegios: []ColegioEntrada{{Nombre: "", Distrito: "D"}}},
	}
	for _, datos := range casos {
		if _, err := NuevaImportacion(datos, time.Now()); err == nil {
			t.Fatalf("se esperaba error para %+v", datos)
		}
	}
	if _, err := NuevaImportacion(DatosImportacion{
		Fuente: "origen", Cursos: []CursoEntrada{{Nombre: "Álgebra", Grado: "5to"}},
		Colegios: []ColegioEntrada{{Nombre: "C", Distrito: "D"}},
	}, time.Time{}); err == nil {
		t.Fatal("una fecha cero debe rechazarse")
	}
}

func TestNuevaSolicitudColegio(t *testing.T) {
	solicitud, err := NuevaSolicitudColegio(SolicitudColegioEntrada{
		Nombre: "  Colegio   Nuevo ", Distrito: "  Lima  ",
	}, time.Now())
	if err != nil || solicitud.Nombre != "Colegio Nuevo" || solicitud.Estado != EstadoSolicitudPendiente {
		t.Fatalf("solicitud inesperada: %+v err=%v", solicitud, err)
	}
	if _, err := NuevaSolicitudColegio(SolicitudColegioEntrada{}, time.Now()); err == nil {
		t.Fatal("una solicitud vacía debe rechazarse")
	}
	if _, err := NuevaSolicitudColegio(SolicitudColegioEntrada{Nombre: "C", Distrito: "D"}, time.Time{}); err == nil {
		t.Fatal("una fecha cero debe rechazarse")
	}
}

func TestCursoRevisar(t *testing.T) {
	curso := &Curso{ID: 1, Estado: EstadoPendiente}
	if err := curso.Revisar(EstadoAprobado, "coord-1", "", time.Now()); err != nil {
		t.Fatalf("aprobar debería funcionar: %v", err)
	}
	if curso.Estado != EstadoAprobado || curso.RevisadoPor != "coord-1" || curso.RevisadoEn == nil {
		t.Fatalf("curso no actualizado: %+v", curso)
	}
	if err := curso.Revisar(EstadoAprobado, "coord-1", "", time.Now()); err == nil {
		t.Fatal("un curso ya revisado no debe revisarse otra vez")
	}

	rechazado := &Curso{ID: 2, Estado: EstadoPendiente}
	if err := rechazado.Revisar(EstadoRechazado, "coord-1", "", time.Now()); err == nil {
		t.Fatal("rechazar exige motivo")
	}
	if err := rechazado.Revisar(EstadoRechazado, "coord-1", strings.Repeat("m", 501), time.Now()); err == nil {
		t.Fatal("un motivo demasiado largo debe rechazarse")
	}
	if err := rechazado.Revisar(EstadoRechazado, "coord-1", "no cumple", time.Now()); err != nil {
		t.Fatalf("rechazo válido falló: %v", err)
	}
	if err := rechazado.Revisar(EstadoAprobado, "coord-1", "", time.Now()); err == nil {
		t.Fatal("un curso rechazado no debe aprobarse")
	}
	if err := (&Curso{}).Revisar(EstadoAprobado, "coord-1", "", time.Now()); err == nil {
		t.Fatal("un curso sin ID debe rechazarse")
	}
}

func TestCopiarImportacionEsProfunda(t *testing.T) {
	if CopiarImportacion(nil) != nil {
		t.Fatal("copiar nil debe devolver nil")
	}
	revisado := time.Now()
	original := &Importacion{
		Cursos:   []Curso{{ID: 1, RevisadoEn: &revisado}},
		Colegios: []Colegio{{Nombre: "C"}},
	}
	copia := CopiarImportacion(original)
	copia.Cursos[0].ID = 99
	*copia.Cursos[0].RevisadoEn = revisado.Add(time.Hour)
	copia.Colegios[0].Nombre = "Otro"
	if original.Cursos[0].ID != 1 || !original.Cursos[0].RevisadoEn.Equal(revisado) || original.Colegios[0].Nombre != "C" {
		t.Fatal("la copia no debe compartir datos con el original")
	}
	if ClaveNormalizada("  Álgebra   Básica ") != "álgebra básica" {
		t.Fatal("la clave normalizada no compacta espacios")
	}
}
