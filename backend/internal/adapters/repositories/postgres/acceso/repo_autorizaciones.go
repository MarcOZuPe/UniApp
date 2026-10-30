package acceso

import (
	"context"
	"database/sql"
	"errors"
)

const consultaAutorizacion = `SELECT a.id, a.vinculacion_id, a.apoderado_id, a.escolar_id, a.estado, a.creada_en
	FROM autorizaciones_apoderado a
	JOIN vinculaciones_apoderado v ON v.id = a.vinculacion_id AND v.estado = 'aceptada' `

func (repositorio RepositorioPostgres) AutorizacionVigenteDeEscolar(ctx context.Context, escolarID string) (*Autorizacion, error) {
	return escanearAutorizacion(repositorio.DB.QueryRowContext(ctx,
		consultaAutorizacion+`WHERE a.escolar_id=$1
		ORDER BY a.creada_en DESC, a.id DESC LIMIT 1`, escolarID))
}

func (repositorio RepositorioPostgres) AutorizacionesDeVinculacion(ctx context.Context, vinculacionID string) ([]Autorizacion, error) {
	filas, err := repositorio.DB.QueryContext(ctx,
		consultaAutorizacion+`WHERE a.vinculacion_id=$1
		ORDER BY a.creada_en DESC, a.id DESC`, vinculacionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = filas.Close() }()

	autorizaciones := make([]Autorizacion, 0)
	for filas.Next() {
		var autorizacion Autorizacion
		var estadoTexto string
		if err := filas.Scan(
			&autorizacion.ID, &autorizacion.VinculacionID, &autorizacion.ApoderadoID,
			&autorizacion.EscolarID, &estadoTexto, &autorizacion.CreadaEn,
		); err != nil {
			return nil, err
		}
		autorizacion.Estado = EstadoAutorizacion(estadoTexto)
		autorizaciones = append(autorizaciones, autorizacion)
	}
	return autorizaciones, filas.Err()
}

func (repositorio RepositorioPostgres) CambiarAutorizacion(ctx context.Context, cambio CambioAutorizacion) (*Autorizacion, error) {
	transaccion, err := repositorio.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = transaccion.Rollback() }()

	var apoderadoDelVinculo, escolarDelVinculo string
	err = transaccion.QueryRowContext(ctx,
		`SELECT apoderado_id, escolar_id FROM vinculaciones_apoderado
		 WHERE id=$1 AND estado='aceptada' FOR UPDATE`,
		cambio.VinculacionID).Scan(&apoderadoDelVinculo, &escolarDelVinculo)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	if apoderadoDelVinculo != cambio.ApoderadoID || escolarDelVinculo != cambio.EscolarID {
		return nil, ErrNoEncontrado
	}

	vigente, err := escanearAutorizacion(transaccion.QueryRowContext(ctx,
		`SELECT id, vinculacion_id, apoderado_id, escolar_id, estado, creada_en
		 FROM autorizaciones_apoderado
		 WHERE vinculacion_id=$1 ORDER BY creada_en DESC, id DESC LIMIT 1`,
		cambio.VinculacionID))
	if err != nil && !errors.Is(err, ErrNoEncontrado) {
		return nil, err
	}
	if err == nil && vigente.Estado == cambio.Estado {
		if err := transaccion.Commit(); err != nil {
			return nil, err
		}
		return vigente, nil
	}

	autorizacionCambiada := Autorizacion{
		VinculacionID: cambio.VinculacionID,
		ApoderadoID:   cambio.ApoderadoID,
		EscolarID:     cambio.EscolarID,
		Estado:        cambio.Estado,
		CreadaEn:      cambio.Momento.UTC(),
	}
	if err := transaccion.QueryRowContext(ctx,
		`INSERT INTO autorizaciones_apoderado (vinculacion_id, apoderado_id, escolar_id, estado, creada_en)
		 VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		autorizacionCambiada.VinculacionID, autorizacionCambiada.ApoderadoID, autorizacionCambiada.EscolarID, autorizacionCambiada.Estado, autorizacionCambiada.CreadaEn).
		Scan(&autorizacionCambiada.ID); err != nil {
		return nil, err
	}
	if err := transaccion.Commit(); err != nil {
		return nil, err
	}
	return &autorizacionCambiada, nil
}

func escanearAutorizacion(fila escaneadorFila) (*Autorizacion, error) {
	var autorizacion Autorizacion
	var estadoTexto string
	if err := fila.Scan(
		&autorizacion.ID, &autorizacion.VinculacionID, &autorizacion.ApoderadoID,
		&autorizacion.EscolarID, &estadoTexto, &autorizacion.CreadaEn,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoEncontrado
		}
		return nil, err
	}
	autorizacion.Estado = EstadoAutorizacion(estadoTexto)
	return &autorizacion, nil
}
