package sesion

import (
    "errors"
    "time"
)

var (
    ErrSesionInvalida = errors.New("sesion invalida")
    ErrNoAutorizado = errors.New("participante no autorizado")
    ErrAsistenciaDuplicada = errors.New("asistencia ya declarada")
)

type Rol string

const (
    RolTutor Rol = "tutor"
    RolEscolar Rol = "escolar"
)

type Asistencia struct {
    UsuarioID string
    Rol Rol
    DeclaradaEn time.Time
}

type Sesion struct {
    ID string
    ReservaID string
    TutorID string
    EscolarID string
    Inicio time.Time
    Fin time.Time
    Asistencias []Asistencia
}

func (s *Sesion) Validar() error {
    if s.ID == "" || s.ReservaID == "" ||
        s.TutorID == "" || s.EscolarID == "" ||
        s.TutorID == s.EscolarID ||
        s.Inicio.IsZero() || s.Fin.IsZero() ||
        !s.Fin.After(s.Inicio) {
        return ErrSesionInvalida
    }

    return nil
}

func (s *Sesion) DeclararAsistencia(
    usuarioID string,
    ahora time.Time,
) error {
    if err := s.Validar(); err != nil {
        return err
    }

    if ahora.IsZero() {
        return ErrSesionInvalida
    }

    var rol Rol

    switch usuarioID {
    case s.TutorID:
        rol = RolTutor
    case s.EscolarID:
        rol = RolEscolar
    default:
        return ErrNoAutorizado
    }

    for _, asistencia := range s.Asistencias {
        if asistencia.UsuarioID == usuarioID {
            return ErrAsistenciaDuplicada
        }
    }

    s.Asistencias = append(s.Asistencias, Asistencia{
        UsuarioID: usuarioID,
        Rol: rol,
        DeclaradaEn: ahora.UTC(),
    })

    return nil
}
