package sesion

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type RepositorioOutboxPostgres struct{ DB *sql.DB }

func (r RepositorioOutboxPostgres) Reclamar(ctx context.Context, ahora time.Time) (*Entrega, error) {
	// Recuperar también procesos que murieron durante su último intento.
	if _, err := r.DB.ExecContext(ctx, `UPDATE sesion_outbox SET estado='agotado', ultimo_error='intentos_agotados', bloqueado_hasta=NULL
 WHERE estado='procesando' AND bloqueado_hasta <= $1 AND intentos >= $2`, ahora, MaxIntentosOutbox); err != nil {
		return nil, err
	}
	var e Entrega
	err := r.DB.QueryRowContext(ctx, `WITH siguiente AS (
 SELECT id FROM sesion_outbox WHERE intentos < $2 AND
 ((estado='pendiente' AND disponible_en <= $1) OR (estado='procesando' AND bloqueado_hasta <= $1))
 ORDER BY disponible_en,id FOR UPDATE SKIP LOCKED LIMIT 1
 ) UPDATE sesion_outbox o SET estado='procesando', intentos=o.intentos+1, bloqueado_hasta=$3
 FROM siguiente s WHERE o.id=s.id
 RETURNING o.id,o.reserva_id::text,o.destinatario_id,o.tipo,o.canal,o.asunto,o.cuerpo,o.intentos`,
		ahora, MaxIntentosOutbox, ahora.Add(PlazoReclamoOutbox)).Scan(&e.ID, &e.ReservaID, &e.DestinatarioID, &e.Tipo, &e.Canal, &e.Asunto, &e.Cuerpo, &e.Intentos)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r RepositorioOutboxPostgres) Resolver(ctx context.Context, e Entrega, ahora time.Time, codigo string) error {
	estado := "entregado"
	var entregado any = ahora
	if codigo != "" {
		estado = "pendiente"
		entregado = nil
		if e.Intentos >= MaxIntentosOutbox {
			estado = "agotado"
		}
	}
	// intentos actúa como versión del reclamo: un worker antiguo no puede completarlo.
	result, err := r.DB.ExecContext(ctx, `UPDATE sesion_outbox SET estado=$3,ultimo_error=$4,
 disponible_en=$5,entregado_en=$6,bloqueado_hasta=NULL
 WHERE id=$1 AND intentos=$2 AND estado='procesando' AND bloqueado_hasta > $7`,
		e.ID, e.Intentos, estado, codigo, ahora.Add(EsperaOutbox(e.Intentos)), entregado, ahora)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrReclamoVencido
	}
	return nil
}

func (r RepositorioOutboxPostgres) CorreoDe(ctx context.Context, id string) (string, error) {
	var correo string
	err := r.DB.QueryRowContext(ctx, "SELECT email FROM usuarios WHERE id::text=$1", id).Scan(&correo)
	return correo, err
}

func (r RepositorioOutboxPostgres) NotificacionesDe(ctx context.Context, id string) ([]Entrega, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,reserva_id::text,tipo,asunto,cuerpo,entregado_en FROM sesion_outbox
 WHERE destinatario_id=$1 AND canal='interno' AND estado='entregado' ORDER BY entregado_en DESC,id DESC LIMIT 100`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resultado := []Entrega{}
	for rows.Next() {
		var e Entrega
		if err := rows.Scan(&e.ID, &e.ReservaID, &e.Tipo, &e.Asunto, &e.Cuerpo, &e.EntregadoEn); err != nil {
			return nil, err
		}
		resultado = append(resultado, e)
	}
	return resultado, rows.Err()
}
