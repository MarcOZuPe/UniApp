package acceso

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

const consultaVinculacion = `SELECT v.id, v.apoderado_id, v.escolar_id, v.estado, v.token_hash,
	v.expira_en, v.creada_en, v.aceptada_en,
	a.nombre_uso, a.email, e.nombre_uso, e.email,
	auth.id, auth.vinculacion_id, auth.apoderado_id, auth.escolar_id, auth.estado, auth.creada_en
	FROM vinculaciones_apoderado v
	JOIN usuarios a ON a.id = v.apoderado_id
	JOIN usuarios e ON e.id = v.escolar_id
	LEFT JOIN LATERAL (
		SELECT au.id, au.vinculacion_id, au.apoderado_id, au.escolar_id, au.estado, au.creada_en
		FROM autorizaciones_apoderado au
		WHERE au.vinculacion_id = v.id
		ORDER BY au.creada_en DESC, au.id DESC
		LIMIT 1
	) auth ON TRUE `

func (repositorio RepositorioPostgres) GuardarSolicitudVinculacion(ctx context.Context, vinculacion Vinculacion) (*Vinculacion, error) {
	vinculacionCreada := vinculacion
	err := repositorio.DB.QueryRowContext(ctx,
		`INSERT INTO vinculaciones_apoderado (apoderado_id, escolar_id, estado, token_hash, expira_en)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id, creada_en`,
		vinculacion.ApoderadoID, vinculacion.EscolarID, vinculacion.Estado, vinculacion.TokenHash, vinculacion.ExpiraEn).
		Scan(&vinculacionCreada.ID, &vinculacionCreada.CreadaEn)
	if err != nil {
		return nil, err
	}
	return &vinculacionCreada, nil
}

func (repositorio RepositorioPostgres) SolicitudPendienteDeEscolar(ctx context.Context, escolarID string, momento time.Time) (*Vinculacion, error) {
	return escanearVinculacion(repositorio.DB.QueryRowContext(ctx,
		consultaVinculacion+`WHERE v.escolar_id=$1 AND v.estado='pendiente' AND v.expira_en > $2
		ORDER BY v.creada_en DESC LIMIT 1`, escolarID, momento))
}

func (repositorio RepositorioPostgres) SolicitudVinculacionPorToken(ctx context.Context, consulta ConsultaVinculacionPorToken) (*Vinculacion, error) {
	return escanearVinculacion(repositorio.DB.QueryRowContext(ctx,
		consultaVinculacion+`WHERE v.token_hash=$1 AND v.escolar_id=$2 AND v.estado='pendiente' AND v.expira_en > $3
		ORDER BY v.creada_en DESC LIMIT 1`, consulta.TokenHash, consulta.EscolarID, consulta.Ahora))
}

func (repositorio RepositorioPostgres) VinculacionAceptadaDeEscolar(ctx context.Context, escolarID string) (*Vinculacion, error) {
	vinculacion, err := escanearVinculacion(repositorio.DB.QueryRowContext(ctx,
		consultaVinculacion+`WHERE v.escolar_id=$1 AND v.estado='aceptada'
		ORDER BY v.aceptada_en DESC LIMIT 1`, escolarID))
	if err != nil {
		var errorPostgres *pq.Error
		if errors.As(err, &errorPostgres) && errorPostgres.Code == "22P02" {
			return nil, ErrNoEncontrado
		}
		return nil, err
	}
	return vinculacion, nil
}

func (repositorio RepositorioPostgres) ConsumirSolicitudVinculacion(ctx context.Context, consulta ConsultaVinculacionPorToken) (*Vinculacion, error) {
	consultaSQL := `WITH actualizada AS (
		UPDATE vinculaciones_apoderado
		SET estado='aceptada', aceptada_en=$3
		WHERE token_hash=$1 AND escolar_id=$2 AND estado='pendiente' AND expira_en > $3
		RETURNING id, apoderado_id, escolar_id, estado, token_hash, expira_en, creada_en, aceptada_en
	), inicial AS (
		INSERT INTO autorizaciones_apoderado (vinculacion_id, apoderado_id, escolar_id, estado, creada_en)
		SELECT id, apoderado_id, escolar_id, 'autorizado', $3 FROM actualizada
		RETURNING id, vinculacion_id, apoderado_id, escolar_id, estado, creada_en
	)
	SELECT v.id, v.apoderado_id, v.escolar_id, v.estado, v.token_hash,
		v.expira_en, v.creada_en, v.aceptada_en,
		a.nombre_uso, a.email, e.nombre_uso, e.email,
		auth.id, auth.vinculacion_id, auth.apoderado_id, auth.escolar_id, auth.estado, auth.creada_en
	FROM actualizada v
	JOIN usuarios a ON a.id = v.apoderado_id
	JOIN usuarios e ON e.id = v.escolar_id
	JOIN inicial auth ON auth.vinculacion_id = v.id`
	vinculacion, err := escanearVinculacion(repositorio.DB.QueryRowContext(ctx, consultaSQL, consulta.TokenHash, consulta.EscolarID, consulta.Ahora))
	if err == nil {
		return vinculacion, nil
	}

	var errorPostgres *pq.Error
	if errors.As(err, &errorPostgres) && errorPostgres.Code == "23505" {
		return nil, ErrYaVinculado
	}
	if errors.Is(err, ErrNoEncontrado) {
		return nil, ErrTokenInvalido
	}
	return nil, err
}

func (repositorio RepositorioPostgres) VinculacionesDeApoderado(ctx context.Context, apoderadoID string, momento time.Time) ([]Vinculacion, error) {
	filas, err := repositorio.DB.QueryContext(ctx, consultaVinculacion+`
		WHERE v.apoderado_id=$1 AND (v.estado='aceptada' OR v.expira_en > $2)
		ORDER BY CASE WHEN v.estado='aceptada' THEN 0 ELSE 1 END, v.creada_en DESC`, apoderadoID, momento)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	vinculaciones := make([]Vinculacion, 0)
	for filas.Next() {
		vinculacion, err := escanearVinculacion(filas)
		if err != nil {
			return nil, err
		}
		vinculaciones = append(vinculaciones, *vinculacion)
	}
	return vinculaciones, filas.Err()
}

func escanearVinculacion(fila escaneadorFila) (*Vinculacion, error) {
	var vinculacion Vinculacion
	var estadoTexto string
	var aceptadaEn sql.NullTime
	var autorizacionID, autorizacionVinculacionID, autorizacionApoderadoID, autorizacionEscolarID, autorizacionEstado sql.NullString
	var autorizacionCreadaEn sql.NullTime
	if err := fila.Scan(
		&vinculacion.ID, &vinculacion.ApoderadoID, &vinculacion.EscolarID, &estadoTexto,
		&vinculacion.TokenHash, &vinculacion.ExpiraEn, &vinculacion.CreadaEn, &aceptadaEn,
		&vinculacion.ApoderadoNombre, &vinculacion.ApoderadoEmail,
		&vinculacion.EscolarNombre, &vinculacion.EscolarEmail,
		&autorizacionID, &autorizacionVinculacionID, &autorizacionApoderadoID, &autorizacionEscolarID,
		&autorizacionEstado, &autorizacionCreadaEn,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoEncontrado
		}
		return nil, err
	}
	vinculacion.Estado = EstadoVinculacion(estadoTexto)
	if aceptadaEn.Valid {
		momento := aceptadaEn.Time
		vinculacion.AceptadaEn = &momento
	}
	if autorizacionID.Valid {
		vinculacion.Autorizacion = &Autorizacion{
			ID:            autorizacionID.String,
			VinculacionID: autorizacionVinculacionID.String,
			ApoderadoID:   autorizacionApoderadoID.String,
			EscolarID:     autorizacionEscolarID.String,
			Estado:        EstadoAutorizacion(autorizacionEstado.String),
			CreadaEn:      autorizacionCreadaEn.Time,
		}
	}
	return &vinculacion, nil
}
