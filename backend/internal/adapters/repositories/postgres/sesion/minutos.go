package sesion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// RepositorioMinutosPostgres implementa RepositorioMinutos (I17).
type RepositorioMinutosPostgres struct {
	DB *sql.DB
}

const columnasMinutos = `m.sesion_id, s.reserva_id, s.tutor_id, s.escolar_id, s.fin,
	m.estado, m.minutos_propuestos, m.minutos_verificados, m.disputado_por`

func (r RepositorioMinutosPostgres) CrearMinutos(ctx context.Context, registro *RegistroMinutos) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO sesion_minutos
		(sesion_id, estado, minutos_propuestos, minutos_verificados, disputado_por)
		VALUES ($1, $2, $3, $4, $5)`,
		registro.SesionID(), string(registro.Estado()), registro.MinutosPropuestos(),
		nuloSi(registro.MinutosVerificados()), registro.DisputadoPor())
	if err != nil {
		var pg *pq.Error
		if errors.As(err, &pg) && pg.Code == "23505" {
			return ErrMinutosDuplicados
		}
		return err
	}
	if err := insertarAuditoria(ctx, tx, registro.SesionID(), registro.Auditoria()); err != nil {
		return err
	}
	return tx.Commit()
}

func (r RepositorioMinutosPostgres) MinutosPorSesion(ctx context.Context, sesionID string) (*RegistroMinutos, error) {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	registro, err := cargarMinutos(ctx, tx, sesionID, false)
	if err != nil {
		return nil, err
	}
	return registro, tx.Commit()
}

// ActualizarMinutos bloquea la fila (FOR UPDATE), aplica el cambio en el
// dominio y persiste el estado y SOLO las entradas de auditoria nuevas.
func (r RepositorioMinutosPostgres) ActualizarMinutos(
	ctx context.Context, sesionID string, cambio func(*RegistroMinutos) error,
) (*RegistroMinutos, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	registro, err := cargarMinutos(ctx, tx, sesionID, true)
	if err != nil {
		return nil, err
	}
	previas := len(registro.Auditoria())
	if err := cambio(registro); err != nil {
		return nil, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE sesion_minutos
		SET estado = $2, minutos_verificados = $3, disputado_por = $4
		WHERE sesion_id = $1`,
		sesionID, string(registro.Estado()), nuloSi(registro.MinutosVerificados()), registro.DisputadoPor())
	if err != nil {
		return nil, err
	}
	if err := insertarAuditoria(ctx, tx, sesionID, registro.Auditoria()[previas:]); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return registro, nil
}

func (r RepositorioMinutosPostgres) MinutosPendientes(ctx context.Context) ([]*RegistroMinutos, error) {
	return r.listar(ctx, `m.estado IN ('propuesto', 'en_disputa')`, nil)
}

func (r RepositorioMinutosPostgres) MinutosDeTutor(ctx context.Context, tutorID string) ([]*RegistroMinutos, error) {
	return r.listar(ctx, `s.tutor_id = $1`, []any{tutorID})
}

func (r RepositorioMinutosPostgres) listar(ctx context.Context, filtro string, argumentos []any) ([]*RegistroMinutos, error) {
	tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	filas, err := tx.QueryContext(ctx, `SELECT m.sesion_id
		FROM sesion_minutos m JOIN sesion_sesiones s ON s.id = m.sesion_id
		WHERE `+filtro+` ORDER BY s.fin, m.sesion_id`, argumentos...)
	if err != nil {
		return nil, err
	}
	var ids []string
	for filas.Next() {
		var id string
		if err := filas.Scan(&id); err != nil {
			filas.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return nil, err
	}
	registros := make([]*RegistroMinutos, 0, len(ids))
	for _, id := range ids {
		registro, err := cargarMinutos(ctx, tx, id, false)
		if err != nil {
			return nil, err
		}
		registros = append(registros, registro)
	}
	return registros, tx.Commit()
}

func cargarMinutos(ctx context.Context, tx *sql.Tx, sesionID string, bloquear bool) (*RegistroMinutos, error) {
	consulta := `SELECT ` + columnasMinutos + `
		FROM sesion_minutos m JOIN sesion_sesiones s ON s.id = m.sesion_id
		WHERE m.sesion_id = $1`
	if bloquear {
		consulta += ` FOR UPDATE OF m`
	}
	var d DatosRegistroMinutos
	var estado string
	var verificados sql.NullInt64
	err := tx.QueryRowContext(ctx, consulta, sesionID).Scan(
		&d.SesionID, &d.ReservaID, &d.TutorID, &d.EscolarID, &d.FinSesion,
		&estado, &d.Propuestos, &verificados, &d.DisputadoPor)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrMinutosNoEncontrados
	}
	if err != nil {
		return nil, err
	}
	d.Estado = EstadoMinutos(estado)
	d.Verificados = enteroDe(verificados)

	filas, err := tx.QueryContext(ctx, `SELECT accion, actor_id, valor_anterior, valor_nuevo, motivo, registrada_en
		FROM sesion_minutos_auditoria WHERE sesion_id = $1 ORDER BY id`, sesionID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	for filas.Next() {
		var e DatosEntradaAuditoria
		var accion string
		var anterior, nuevo sql.NullInt64
		if err := filas.Scan(&accion, &e.ActorID, &anterior, &nuevo, &e.Motivo, &e.RegistradaEn); err != nil {
			return nil, err
		}
		e.Accion = AccionMinutos(accion)
		e.ValorAnterior, e.ValorNuevo = enteroDe(anterior), enteroDe(nuevo)
		d.Auditoria = append(d.Auditoria, e)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}
	registro, err := RestaurarRegistroMinutos(d)
	if err != nil {
		return nil, fmt.Errorf("minutos almacenados invalidos: %w", err)
	}
	return registro, nil
}

func insertarAuditoria(ctx context.Context, tx *sql.Tx, sesionID string, entradas []EntradaAuditoria) error {
	for _, e := range entradas {
		_, err := tx.ExecContext(ctx, `INSERT INTO sesion_minutos_auditoria
			(sesion_id, accion, actor_id, valor_anterior, valor_nuevo, motivo, registrada_en)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			sesionID, string(e.Accion()), e.ActorID(), nuloSi(e.ValorAnterior()), nuloSi(e.ValorNuevo()),
			e.Motivo(), e.RegistradaEn().UTC().Truncate(time.Microsecond))
		if err != nil {
			return err
		}
	}
	return nil
}

func nuloSi(valor int, presente bool) any {
	if !presente {
		return nil
	}
	return valor
}

func enteroDe(valor sql.NullInt64) *int {
	if !valor.Valid {
		return nil
	}
	v := int(valor.Int64)
	return &v
}
