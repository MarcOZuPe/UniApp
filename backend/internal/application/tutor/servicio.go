package tutor

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type Servicio struct {
	repositorio Repositorio
	almacen     AlmacenDocumentos
	ahora       func() time.Time
}

func NuevoServicio(repositorio Repositorio, almacen AlmacenDocumentos) *Servicio {
	return &Servicio{
		repositorio: repositorio,
		almacen:     almacen,
		ahora:       func() time.Time { return time.Now().UTC() },
	}
}

func (s *Servicio) Presentar(ctx context.Context, tutorID string, datos DatosExpediente) (*Expediente, error) {
	expediente, err := NuevoExpediente(tutorID, datos, s.ahora())
	if err != nil {
		return nil, err
	}

	guardados := make([]string, 0, len(datos.Documentos))
	for indice, archivo := range datos.Documentos {
		id, err := nuevoID()
		if err != nil {
			return nil, errors.Join(err, s.eliminarDocumentos(context.Background(), guardados))
		}
		if err := s.almacen.Guardar(ctx, id, archivo.Contenido); err != nil {
			return nil, errors.Join(err, s.eliminarDocumentos(context.Background(), append(guardados, id)))
		}
		guardados = append(guardados, id)
		expediente.Documentos[indice].ID = id
	}

	if err := s.repositorio.Guardar(ctx, expediente); err != nil {
		return nil, errors.Join(err, s.eliminarDocumentos(context.Background(), guardados))
	}
	return Copiar(expediente), nil
}

func (s *Servicio) MiExpediente(ctx context.Context, tutorID string) (*Expediente, error) {
	return s.repositorio.PorTutor(ctx, tutorID)
}

func (s *Servicio) Pendientes(ctx context.Context) ([]Expediente, error) {
	return s.repositorio.Pendientes(ctx)
}

func (s *Servicio) Revisar(ctx context.Context, tutorID, revisorID string, estado Estado, motivo string, venceEn time.Time) (*Expediente, error) {
	expediente, err := s.repositorio.PorTutor(ctx, tutorID)
	if err != nil {
		return nil, err
	}
	if err := expediente.Revisar(estado, revisorID, motivo, venceEn, s.ahora()); err != nil {
		return nil, err
	}
	if err := s.repositorio.Revisar(ctx, expediente); err != nil {
		return nil, err
	}
	return expediente, nil
}

func (s *Servicio) PerfilPublico(ctx context.Context, tutorID string) (*Expediente, error) {
	expediente, err := s.repositorio.PorTutor(ctx, tutorID)
	if err != nil {
		return nil, err
	}
	if !expediente.Publicable(s.ahora()) {
		return nil, ErrNoPublicable
	}
	expediente.Documentos = nil
	expediente.Motivo = ""
	expediente.RevisadoPor = ""
	expediente.RevisadoEn = nil
	return expediente, nil
}

func (s *Servicio) PerfilesPublicos(ctx context.Context) ([]PerfilListado, error) {
	return s.repositorio.PerfilesPublicos(ctx, s.ahora())
}

func (s *Servicio) DescargarDocumento(ctx context.Context, documentoID, usuarioID string, coordinador bool) ([]byte, Documento, error) {
	documento, propietarioID, err := s.repositorio.Documento(ctx, documentoID)
	if err != nil {
		return nil, Documento{}, err
	}
	if !coordinador && propietarioID != usuarioID {
		return nil, Documento{}, ErrSinPermiso
	}
	contenido, err := s.almacen.Leer(ctx, documentoID)
	if err != nil {
		return nil, Documento{}, err
	}
	return contenido, documento, nil
}

func (s *Servicio) eliminarDocumentos(ctx context.Context, ids []string) error {
	var errores []error
	for _, id := range ids {
		if err := s.almacen.Eliminar(ctx, id); err != nil {
			errores = append(errores, err)
		}
	}
	return errors.Join(errores...)
}

func nuevoID() (string, error) {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	hexID := hex.EncodeToString(bytes[:])
	return strings.Join([]string{hexID[:8], hexID[8:12], hexID[12:16], hexID[16:20], hexID[20:]}, "-"), nil
}
