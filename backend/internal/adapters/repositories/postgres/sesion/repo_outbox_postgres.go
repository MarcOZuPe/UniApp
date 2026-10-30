package sesion

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type RepositorioOutboxPostgres struct{ DB *sql.DB }

func (repositorio RepositorioOutboxPostgres) Reclamar(ctx context.Context, momento time.Time) (*Entrega, error) {
	if _, err := repositorio.DB.ExecContext(ctx, `UPDATE sesion_outbox SET estado='agotado', ultimo_error='intentos_agotados', bloqueado_hasta=NULL
 WHERE estado='procesando' AND bloqueado_hasta <= $1 AND intentos >= $2`, momento, MaxIntentosOutbox); err != nil {
		return nil, err
	}

	var entrega Entrega
	err := repositorio.DB.QueryRowContext(ctx, `WITH siguiente AS (
 SELECT id FROM sesion_outbox WHERE intentos < $2 AND
 ((estado='pendiente' AND disponible_en <= $1) OR (estado='procesando' AND bloqueado_hasta <= $1))
 ORDER BY disponible_en,id FOR UPDATE SKIP LOCKED LIMIT 1
 ) UPDATE sesion_outbox o SET estado='procesando', intentos=o.intentos+1, bloqueado_hasta=$3
 FROM siguiente s WHERE o.id=s.id
 RETURNING o.id,o.reserva_id::text,o.destinatario_id,o.tipo,o.canal,o.asunto,o.cuerpo,o.intentos`,
		momento, MaxIntentosOutbox, momento.Add(PlazoReclamoOutbox)).
		Scan(&entrega.ID, &entrega.ReservaID, &entrega.DestinatarioID, &entrega.Tipo,
			&entrega.Canal, &entrega.Asunto, &entrega.Cuerpo, &entrega.Intentos)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entrega, nil
}

func (repositorio RepositorioOutboxPostgres) Resolver(ctx context.Context, entrega Entrega, momento time.Time, codigo string) error {
	estadoEntrega := "entregado"
	var entregadoEn any = momento
	if codigo != "" {
		estadoEntrega = "pendiente"
		entregadoEn = nil
		if entrega.Intentos >= MaxIntentosOutbox {
			estadoEntrega = "agotado"
		}
	}

	resultado, err := repositorio.DB.ExecContext(ctx, `UPDATE sesion_outbox SET estado=$3,ultimo_error=$4,
 disponible_en=$5,entregado_en=$6,bloqueado_hasta=NULL
 WHERE id=$1 AND intentos=$2 AND estado='procesando' AND bloqueado_hasta > $7`,
		entrega.ID, entrega.Intentos, estadoEntrega, codigo,
		momento.Add(EsperaOutbox(entrega.Intentos)), entregadoEn, momento)
	if err != nil {
		return err
	}

	filasAfectadas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if filasAfectadas != 1 {
		return ErrReclamoVencido
	}
	return nil
}

func (repositorio RepositorioOutboxPostgres) CorreoDe(ctx context.Context, usuarioID string) (string, error) {
	var correo string
	err := repositorio.DB.QueryRowContext(ctx, "SELECT email FROM usuarios WHERE id::text=$1", usuarioID).Scan(&correo)
	return correo, err
}

func (repositorio RepositorioOutboxPostgres) NotificacionesDe(ctx context.Context, usuarioID string) ([]Entrega, error) {
	filas, err := repositorio.DB.QueryContext(ctx, `SELECT id,reserva_id::text,tipo,asunto,cuerpo,entregado_en FROM sesion_outbox
 WHERE destinatario_id=$1 AND canal='interno' AND estado='entregado' ORDER BY entregado_en DESC,id DESC LIMIT 100`, usuarioID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	entregas := []Entrega{}
	for filas.Next() {
		var entrega Entrega
		if err := filas.Scan(&entrega.ID, &entrega.ReservaID, &entrega.Tipo,
			&entrega.Asunto, &entrega.Cuerpo, &entrega.EntregadoEn); err != nil {
			return nil, err
		}
		entregas = append(entregas, entrega)
	}
	return entregas, filas.Err()
}
