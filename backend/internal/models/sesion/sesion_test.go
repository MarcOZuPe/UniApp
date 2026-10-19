package sesion

import (
    "errors"
    "testing"
    "time"
)

func nuevaSesionPrueba() Sesion {
    inicio := time.Date(
        2026, 10, 20, 15, 0, 0, 0, time.UTC,
    )

    return Sesion{
        ID: "sesion-001",
        ReservaID: "reserva-sintetica-001",
        TutorID: "tutor-001",
        EscolarID: "escolar-001",
        Inicio: inicio,
        Fin: inicio.Add(time.Hour),
    }
}

func TestDeclararAsistencia(t *testing.T) {
    sesion := nuevaSesionPrueba()

    err := sesion.DeclararAsistencia(
        "escolar-001",
        sesion.Inicio,
    )

    if err != nil {
        t.Fatalf("No deberia fallar: %v", err)
    }

    if len(sesion.Asistencias) != 1 {
        t.Fatal("Se esperaba una declaracion")
    }

    if sesion.Asistencias[0].Rol != RolEscolar {
        t.Fatal("El rol registrado es incorrecto")
    }
}

func TestRechazarParticipanteAjeno(t *testing.T) {
    sesion := nuevaSesionPrueba()

    err := sesion.DeclararAsistencia(
        "usuario-ajeno",
        sesion.Inicio,
    )

    if !errors.Is(err, ErrNoAutorizado) {
        t.Fatalf("Se esperaba acceso denegado: %v", err)
    }

    if len(sesion.Asistencias) != 0 {
        t.Fatal("No debe registrar al usuario ajeno")
    }
}

func TestRechazarAsistenciaDuplicada(t *testing.T) {
    sesion := nuevaSesionPrueba()

    if err := sesion.DeclararAsistencia(
        "tutor-001", sesion.Inicio,
    ); err != nil {
        t.Fatal(err)
    }

    err := sesion.DeclararAsistencia(
        "tutor-001",
        sesion.Inicio.Add(time.Minute),
    )

    if !errors.Is(err, ErrAsistenciaDuplicada) {
        t.Fatalf("Se esperaba duplicado: %v", err)
    }

    if len(sesion.Asistencias) != 1 {
        t.Fatal("No debe duplicar declaraciones")
    }
}

func TestRechazarSesionInvalida(t *testing.T) {
    sesion := nuevaSesionPrueba()
    sesion.Fin = sesion.Inicio

    err := sesion.DeclararAsistencia(
        "tutor-001",
        sesion.Inicio,
    )

    if !errors.Is(err, ErrSesionInvalida) {
        t.Fatalf("Se esperaba sesion invalida: %v", err)
    }
}
