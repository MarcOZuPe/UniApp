package tutor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type RepositorioPostgres struct {
	DB *sql.DB
}

const columnasExpediente = `tutor_id, presentacion, cursos, experiencia, estado, presentado_en,
	vence_en, revisado_por, revisado_en, motivo_rechazo`

func (r RepositorioPostgres) Guardar(ctx context.Context, expediente *Expediente) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, `INSERT INTO tutor_expedientes
		(tutor_id, presentacion, cursos, experiencia, estado, presentado_en, vence_en, revisado_por, revisado_en, motivo_rechazo)
		VALUES ($1,$2,$3,$4,$5,$6,NULL,NULL,NULL,'')
		ON CONFLICT (tutor_id) DO UPDATE SET presentacion=EXCLUDED.presentacion, cursos=EXCLUDED.cursos,
		experiencia=EXCLUDED.experiencia, estado=EXCLUDED.estado, presentado_en=EXCLUDED.presentado_en,
		vence_en=NULL, revisado_por=NULL, revisado_en=NULL, motivo_rechazo='',
		actualizado_en=now()`,
		expediente.TutorID, expediente.Presentacion, expediente.Cursos, expediente.Experiencia,
		string(EstadoPendiente), expediente.PresentadoEn)
	if err != nil {
		return traducirErrorTutor(err)
	}
	for indice := range expediente.Documentos {
		documento := &expediente.Documentos[indice]
		err = tx.QueryRowContext(ctx, `INSERT INTO tutor_documentos
			(id, tutor_id, nombre, tipo, tamano_bytes, creado_en)
			VALUES ($1,$2,$3,$4,$5,$6) RETURNING creado_en`,
			documento.ID, expediente.TutorID, documento.Nombre, documento.Tipo,
			documento.Tamano, expediente.PresentadoEn).Scan(&documento.CreadoEn)
		if err != nil {
			return traducirErrorTutor(err)
		}
	}
	return tx.Commit()
}

func (r RepositorioPostgres) PorTutor(ctx context.Context, tutorID string) (*Expediente, error) {
	expediente, err := r.escanear(r.DB.QueryRowContext(ctx, `SELECT `+columnasExpediente+`
		FROM tutor_expedientes WHERE tutor_id=$1`, tutorID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	documentos, err := r.documentosDe(ctx, tutorID)
	if err != nil {
		return nil, err
	}
	expediente.Documentos = documentos
	return expediente, nil
}

func (r RepositorioPostgres) Pendientes(ctx context.Context) ([]Expediente, error) {
	filas, err := r.DB.QueryContext(ctx, `SELECT `+columnasExpediente+`
		FROM tutor_expedientes WHERE estado=$1 ORDER BY presentado_en`, string(EstadoPendiente))
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()
	var resultado []Expediente
	for filas.Next() {
		expediente, err := r.escanear(filas)
		if err != nil {
			return nil, err
		}
		expediente.Documentos, err = r.documentosDe(ctx, expediente.TutorID)
		if err != nil {
			return nil, err
		}
		resultado = append(resultado, *expediente)
	}
	return resultado, filas.Err()
}

func (r RepositorioPostgres) Revisar(ctx context.Context, expediente *Expediente) error {
	resultado, err := r.DB.ExecContext(ctx, `UPDATE tutor_expedientes SET estado=$2, vence_en=$3,
		revisado_por=$4, revisado_en=$5, motivo_rechazo=$6, actualizado_en=now()
		WHERE tutor_id=$1 AND estado=$7`,
		expediente.TutorID, string(expediente.Estado), expediente.VenceEn, expediente.RevisadoPor,
		expediente.RevisadoEn, expediente.Motivo, string(EstadoPendiente))
	if err != nil {
		return traducirErrorTutor(err)
	}
	filas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if filas == 0 {
		return ErrEstadoInvalido
	}
	return nil
}

func (r RepositorioPostgres) Documento(ctx context.Context, id string) (Documento, string, error) {
	var documento Documento
	var tutorID string
	err := r.DB.QueryRowContext(ctx, `SELECT id, tutor_id, nombre, tipo, tamano_bytes, creado_en
		FROM tutor_documentos WHERE id=$1`, id).Scan(
		&documento.ID, &tutorID, &documento.Nombre, &documento.Tipo, &documento.Tamano, &documento.CreadoEn)
	if errors.Is(err, sql.ErrNoRows) {
		return Documento{}, "", ErrNoEncontrado
	}
	return documento, tutorID, err
}

func (r RepositorioPostgres) documentosDe(ctx context.Context, tutorID string) ([]Documento, error) {
	filas, err := r.DB.QueryContext(ctx, `SELECT id, nombre, tipo, tamano_bytes, creado_en
		FROM tutor_documentos WHERE tutor_id=$1 ORDER BY creado_en, id`, tutorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()
	var documentos []Documento
	for filas.Next() {
		var documento Documento
		if err := filas.Scan(&documento.ID, &documento.Nombre, &documento.Tipo, &documento.Tamano, &documento.CreadoEn); err != nil {
			return nil, err
		}
		documentos = append(documentos, documento)
	}
	return documentos, filas.Err()
}

type escanerExpediente interface {
	Scan(...any) error
}

func (r RepositorioPostgres) escanear(fila escanerExpediente) (*Expediente, error) {
	var expediente Expediente
	var estado string
	var revisadoPor sql.NullString
	err := fila.Scan(&expediente.TutorID, &expediente.Presentacion, &expediente.Cursos,
		&expediente.Experiencia, &estado, &expediente.PresentadoEn, &expediente.VenceEn,
		&revisadoPor, &expediente.RevisadoEn, &expediente.Motivo)
	if err != nil {
		return nil, err
	}
	expediente.Estado = Estado(estado)
	if revisadoPor.Valid {
		expediente.RevisadoPor = revisadoPor.String
	}
	return &expediente, nil
}

func traducirErrorTutor(err error) error {
	var numeroError interface{ SQLState() string }
	if errors.As(err, &numeroError) && numeroError.SQLState() == "23503" {
		return fmt.Errorf("tutor asociado no válido: %w", ErrExpedienteInvalido)
	}
	return err
}
