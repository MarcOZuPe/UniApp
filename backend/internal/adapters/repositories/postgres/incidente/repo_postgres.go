package incidente

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

const columnasIncidente = `i.id, i.escolar_id, e.nombre_uso, i.reportado_por,
	i.tutor_id, i.tutor_nombre, i.facultad_id::text, i.reserva_id, i.curso, i.inicio_sesion, i.tipo,
	i.descripcion, i.estado, i.resultado, i.comentario_revision, i.revisado_por,
	i.revisado_en, i.creado_en`

const columnasEvidencia = `id, nombre, tipo, tamano_bytes, clave_almacenamiento, creado_en`

type RepositorioPostgres struct {
	DB *sql.DB
}

func (r RepositorioPostgres) TutoriasDeEscolar(
	ctx context.Context,
	escolarID string,
	ahora time.Time,
) ([]Tutoria, error) {
	filas, err := r.DB.QueryContext(ctx, `
		SELECT r.id, r.escolar_id, u.nombre_uso, r.tutor_id, r.tutor_nombre,
		       COALESCE(tu.facultad_id::text, ''), r.curso, r.inicio, r.fin,
		       EXISTS (
		           SELECT 1
		           FROM sesion_sesiones s
		           JOIN sesion_asistencias a ON a.sesion_id = s.id
		           WHERE s.reserva_id = r.id::text
		             AND a.usuario_id = r.escolar_id::text
		             AND a.rol = 'escolar'
		       )
		FROM reservas r
		JOIN usuarios u ON u.id = r.escolar_id
		JOIN usuarios tu ON tu.id::text = r.tutor_id
		WHERE r.escolar_id = $1 AND r.estado = 'confirmada' AND r.fin <= $2
		ORDER BY r.inicio DESC, r.id`, escolarID, ahora.UTC())
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	tutorias := make([]Tutoria, 0)
	for filas.Next() {
		var tutoria Tutoria
		if err := filas.Scan(
			&tutoria.ReservaID, &tutoria.EscolarID, &tutoria.EscolarNombre,
			&tutoria.TutorID, &tutoria.TutorNombre, &tutoria.TutorFacultadID,
			&tutoria.Curso, &tutoria.Inicio, &tutoria.Fin, &tutoria.Asistio,
		); err != nil {
			return nil, err
		}
		tutorias = append(tutorias, tutoria)
	}
	return tutorias, filas.Err()
}

func (r RepositorioPostgres) TutoriaDeEscolar(
	ctx context.Context,
	escolarID, reservaID string,
	ahora time.Time,
) (*Tutoria, error) {
	var tutoria Tutoria
	err := r.DB.QueryRowContext(ctx, `
		SELECT r.id, r.escolar_id, u.nombre_uso, r.tutor_id, r.tutor_nombre,
		       COALESCE(tu.facultad_id::text, ''), r.curso, r.inicio, r.fin,
		       EXISTS (
		           SELECT 1
		           FROM sesion_sesiones s
		           JOIN sesion_asistencias a ON a.sesion_id = s.id
		           WHERE s.reserva_id = r.id::text
		             AND a.usuario_id = r.escolar_id::text
		             AND a.rol = 'escolar'
		       )
		FROM reservas r
		JOIN usuarios u ON u.id = r.escolar_id
		JOIN usuarios tu ON tu.id::text = r.tutor_id
		WHERE r.id = $1 AND r.escolar_id = $2 AND r.estado = 'confirmada' AND r.fin <= $3`,
		reservaID, escolarID, ahora.UTC(),
	).Scan(
		&tutoria.ReservaID, &tutoria.EscolarID, &tutoria.EscolarNombre,
		&tutoria.TutorID, &tutoria.TutorNombre, &tutoria.TutorFacultadID,
		&tutoria.Curso, &tutoria.Inicio, &tutoria.Fin, &tutoria.Asistio,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTutoriaNoEncontrada
	}
	if err != nil {
		return nil, err
	}
	return &tutoria, nil
}

func (r RepositorioPostgres) EscolaresDeApoderado(ctx context.Context, apoderadoID string) ([]EscolarVinculado, error) {
	filas, err := r.DB.QueryContext(ctx, `
		SELECT u.id, u.nombre_uso
		FROM vinculaciones_apoderado v
		JOIN usuarios u ON u.id = v.escolar_id
		WHERE v.apoderado_id = $1 AND v.estado = 'aceptada'
		ORDER BY u.nombre_uso, u.id`, apoderadoID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	escolares := make([]EscolarVinculado, 0)
	for filas.Next() {
		var escolar EscolarVinculado
		if err := filas.Scan(&escolar.ID, &escolar.Nombre); err != nil {
			return nil, err
		}
		escolares = append(escolares, escolar)
	}
	return escolares, filas.Err()
}

func (r RepositorioPostgres) EsApoderadoDe(ctx context.Context, apoderadoID, escolarID string) (bool, error) {
	var vinculado bool
	err := r.DB.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM vinculaciones_apoderado
			WHERE apoderado_id = $1 AND escolar_id = $2 AND estado = 'aceptada'
		)`, apoderadoID, escolarID).Scan(&vinculado)
	return vinculado, err
}

func (r RepositorioPostgres) Crear(ctx context.Context, incidente *Incidente) (*Incidente, error) {
	if incidente == nil {
		return nil, ErrIncidenteInvalido
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	copia := *incidente
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO incidentes (
			escolar_id, reportado_por, tutor_id, tutor_nombre, facultad_id, reserva_id, curso,
			inicio_sesion, tipo, descripcion, estado, creado_en
		) VALUES ($1,$2,$3,$4,NULLIF($5,'')::smallint,$6,$7,$8,$9,$10,'pendiente',$11)
		RETURNING id, creado_en`,
		copia.EscolarID, copia.ReportadoPor, copia.TutorID, copia.TutorNombre,
		copia.FacultadID, copia.ReservaID, copia.Curso, copia.Inicio.UTC(), string(copia.Tipo),
		copia.Descripcion, copia.CreadoEn.UTC(),
	).Scan(&copia.ID, &copia.CreadoEn); err != nil {
		return nil, err
	}

	for indice := range copia.Evidencias {
		var creadaEn time.Time
		if err := tx.QueryRowContext(ctx, `
			INSERT INTO incidente_evidencias (
				incidente_id, nombre, tipo, tamano_bytes, clave_almacenamiento, creado_en
			) VALUES ($1,$2,$3,$4,$5,$6)
			RETURNING id, creado_en`,
			copia.ID, copia.Evidencias[indice].Nombre, copia.Evidencias[indice].Tipo,
			copia.Evidencias[indice].TamanoBytes, copia.Evidencias[indice].ClaveAlmacenamiento,
			copia.Evidencias[indice].CreadoEn.UTC(),
		).Scan(&copia.Evidencias[indice].ID, &creadaEn); err != nil {
			return nil, err
		}
		copia.Evidencias[indice].CreadoEn = creadaEn
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &copia, nil
}

func (r RepositorioPostgres) PorID(ctx context.Context, incidenteID string) (*Incidente, error) {
	incidente, err := escanearIncidente(r.DB.QueryRowContext(ctx, `
		SELECT `+columnasIncidente+`
		FROM incidentes i
		JOIN usuarios e ON e.id = i.escolar_id
		WHERE i.id = $1`, incidenteID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrIncidenteNoEncontrado
	}
	if err != nil {
		return nil, err
	}
	incidente.Evidencias, err = r.evidenciasDe(ctx, incidente.ID)
	if err != nil {
		return nil, err
	}
	return incidente, nil
}

func (r RepositorioPostgres) Listar(ctx context.Context, filtro Filtro) ([]Incidente, error) {
	consulta := `
		SELECT ` + columnasIncidente + `
		FROM incidentes i
		JOIN usuarios e ON e.id = i.escolar_id`
	argumentos := make([]any, 0, 1)
	switch {
	case filtro.FacultadID != "":
		consulta += " WHERE i.facultad_id = NULLIF($1,'')::smallint"
		argumentos = append(argumentos, filtro.FacultadID)
	case filtro.Todos:
	case filtro.EscolarID != "":
		consulta += " WHERE i.escolar_id = $1"
		argumentos = append(argumentos, filtro.EscolarID)
	case len(filtro.EscolaresID) > 0:
		consulta += " WHERE i.escolar_id = ANY($1)"
		argumentos = append(argumentos, pq.Array(filtro.EscolaresID))
	default:
		return []Incidente{}, nil
	}
	consulta += " ORDER BY i.creado_en DESC, i.id DESC"

	filas, err := r.DB.QueryContext(ctx, consulta, argumentos...)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	incidentes := make([]Incidente, 0)
	for filas.Next() {
		incidente, err := escanearIncidente(filas)
		if err != nil {
			return nil, err
		}
		incidentes = append(incidentes, *incidente)
	}
	if err := filas.Err(); err != nil {
		return nil, err
	}
	for indice := range incidentes {
		incidentes[indice].Evidencias, err = r.evidenciasDe(ctx, incidentes[indice].ID)
		if err != nil {
			return nil, err
		}
	}
	return incidentes, nil
}

func (r RepositorioPostgres) Revisar(
	ctx context.Context,
	incidenteID, revisorID string,
	resultado Resultado,
	comentario string,
	ahora time.Time,
) (*Incidente, error) {
	operacion, err := r.DB.ExecContext(ctx, `
		UPDATE incidentes
		SET estado = 'revisado', resultado = $2, comentario_revision = $3,
		    revisado_por = $4, revisado_en = $5
		WHERE id = $1 AND estado = 'pendiente'`,
		incidenteID, string(resultado), comentario, revisorID, ahora.UTC())
	if err != nil {
		return nil, err
	}
	afectadas, err := operacion.RowsAffected()
	if err != nil {
		return nil, err
	}
	if afectadas == 0 {
		var existente string
		if err := r.DB.QueryRowContext(ctx, "SELECT estado FROM incidentes WHERE id = $1", incidenteID).Scan(&existente); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrIncidenteNoEncontrado
			}
			return nil, err
		}
		return nil, ErrYaRevisado
	}
	return r.PorID(ctx, incidenteID)
}

func (r RepositorioPostgres) EvidenciaDe(ctx context.Context, incidenteID, evidenciaID string) (Evidencia, error) {
	var evidencia Evidencia
	err := r.DB.QueryRowContext(ctx, `
		SELECT `+columnasEvidencia+`
		FROM incidente_evidencias
		WHERE incidente_id = $1 AND id = $2`, incidenteID, evidenciaID,
	).Scan(
		&evidencia.ID, &evidencia.Nombre, &evidencia.Tipo, &evidencia.TamanoBytes,
		&evidencia.ClaveAlmacenamiento, &evidencia.CreadoEn,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Evidencia{}, ErrEvidenciaNoEncontrada
	}
	return evidencia, err
}

func (r RepositorioPostgres) evidenciasDe(ctx context.Context, incidenteID string) ([]Evidencia, error) {
	filas, err := r.DB.QueryContext(ctx, `
		SELECT `+columnasEvidencia+`
		FROM incidente_evidencias
		WHERE incidente_id = $1
		ORDER BY creado_en, id`, incidenteID)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	evidencias := make([]Evidencia, 0)
	for filas.Next() {
		var evidencia Evidencia
		if err := filas.Scan(
			&evidencia.ID, &evidencia.Nombre, &evidencia.Tipo, &evidencia.TamanoBytes,
			&evidencia.ClaveAlmacenamiento, &evidencia.CreadoEn,
		); err != nil {
			return nil, err
		}
		evidencias = append(evidencias, evidencia)
	}
	return evidencias, filas.Err()
}

type escaneador interface {
	Scan(destino ...any) error
}

func escanearIncidente(fila escaneador) (*Incidente, error) {
	var (
		incidente   Incidente
		resultado   sql.NullString
		revisadoPor sql.NullString
		revisadoEn  sql.NullTime
		tipo        string
		estado      string
	)
	if err := fila.Scan(
		&incidente.ID, &incidente.EscolarID, &incidente.EscolarNombre,
		&incidente.ReportadoPor, &incidente.TutorID, &incidente.TutorNombre,
		&incidente.FacultadID, &incidente.ReservaID, &incidente.Curso, &incidente.Inicio, &tipo,
		&incidente.Descripcion, &estado, &resultado, &incidente.ComentarioRevision,
		&revisadoPor, &revisadoEn, &incidente.CreadoEn,
	); err != nil {
		return nil, err
	}
	incidente.Tipo = Tipo(tipo)
	incidente.Estado = Estado(estado)
	incidente.Resultado = Resultado(resultado.String)
	incidente.RevisadoPor = revisadoPor.String
	if revisadoEn.Valid {
		momento := revisadoEn.Time
		incidente.RevisadoEn = &momento
	}
	return &incidente, nil
}
