package acceso

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var identificadorPorRol = map[Rol]int{
	RolEscolar: 1, RolApoderado: 2, RolTutor: 3, RolCoordinador: 4, RolCoordinadorGeneral: 5,
}

const consultaUsuario = `SELECT u.id, u.email, u.nombre_uso, u.clave_hash, r.nombre,
	COALESCE(u.facultad_id::text, ''), COALESCE(f.nombre, ''),
	u.activo, u.correo_confirmado, u.creado_en, u.ultimo_ingreso
	FROM usuarios u JOIN roles r ON r.id = u.rol_id
	LEFT JOIN facultades f ON f.id = u.facultad_id `

func (repositorio RepositorioPostgres) CrearUsuario(ctx context.Context, usuario *Usuario) (*Usuario, error) {
	registrado := *usuario
	err := repositorio.DB.QueryRowContext(ctx,
		`INSERT INTO usuarios (email, nombre_uso, clave_hash, rol_id, correo_confirmado, facultad_id)
		 VALUES ($1,$2,$3,$4,$5,NULLIF($6,'')::smallint) RETURNING id, creado_en`,
		usuario.Email, usuario.NombreUso, usuario.ClaveHash, identificadorPorRol[usuario.Rol],
		usuario.EmailConfirmado, usuario.FacultadID).
		Scan(&registrado.ID, &registrado.CreadoEn)
	var errorPostgres *pq.Error
	if errors.As(err, &errorPostgres) && errorPostgres.Code == "23505" {
		return nil, ErrEmailEnUso
	}
	if err != nil {
		return nil, err
	}
	return &registrado, nil
}

func (repositorio RepositorioPostgres) UsuarioPorEmail(ctx context.Context, email string) (*Usuario, error) {
	return escanearUsuario(repositorio.DB.QueryRowContext(ctx, consultaUsuario+`WHERE u.email = $1`, email))
}

func (repositorio RepositorioPostgres) UsuarioPorID(ctx context.Context, usuarioID string) (*Usuario, error) {
	return escanearUsuario(repositorio.DB.QueryRowContext(ctx, consultaUsuario+`WHERE u.id = $1`, usuarioID))
}

func (repositorio RepositorioPostgres) UsuariosActivosPorRol(
	ctx context.Context,
	rol Rol,
	facultadID string,
) ([]*Usuario, error) {
	identificador, existe := identificadorPorRol[rol]
	if !existe {
		return nil, ErrNoEncontrado
	}
	filas, err := repositorio.DB.QueryContext(ctx,
		consultaUsuario+`WHERE u.rol_id = $1 AND u.activo = TRUE AND u.facultad_id = NULLIF($2,'')::smallint
		ORDER BY u.nombre_uso, u.id`,
		identificador, facultadID,
	)
	if err != nil {
		return nil, err
	}
	defer filas.Close()

	usuarios := make([]*Usuario, 0)
	for filas.Next() {
		usuario, err := escanearUsuario(filas)
		if err != nil {
			return nil, err
		}
		usuarios = append(usuarios, usuario)
	}
	return usuarios, filas.Err()
}

func (repositorio RepositorioPostgres) DesactivarUsuario(ctx context.Context, usuarioID string) error {
	resultado, err := repositorio.DB.ExecContext(ctx, `UPDATE usuarios SET activo = FALSE WHERE id = $1`, usuarioID)
	if err != nil {
		return err
	}
	filas, err := resultado.RowsAffected()
	if err != nil {
		return err
	}
	if filas == 0 {
		return ErrNoEncontrado
	}
	return nil
}

func (repositorio RepositorioPostgres) ActualizarClave(ctx context.Context, usuarioID, claveHash string) error {
	_, err := repositorio.DB.ExecContext(ctx, `UPDATE usuarios SET clave_hash=$2 WHERE id=$1`, usuarioID, claveHash)
	return err
}

func (repositorio RepositorioPostgres) MarcarCorreoConfirmado(ctx context.Context, usuarioID string) error {
	_, err := repositorio.DB.ExecContext(ctx, `UPDATE usuarios SET correo_confirmado=TRUE WHERE id=$1`, usuarioID)
	return err
}

func (repositorio RepositorioPostgres) RegistrarIngreso(ctx context.Context, usuarioID string, momento time.Time) error {
	_, err := repositorio.DB.ExecContext(ctx, `UPDATE usuarios SET ultimo_ingreso=$2 WHERE id=$1`, usuarioID, momento)
	return err
}

func (repositorio RepositorioPostgres) EliminarUsuariosSinConfirmar(ctx context.Context, momento time.Time) (int64, error) {
	resultado, err := repositorio.DB.ExecContext(ctx, `DELETE FROM usuarios u
		WHERE NOT u.correo_confirmado AND NOT EXISTS (
			SELECT 1 FROM tokens_confirmacion t
			WHERE t.usuario_id = u.id AND t.usado_en IS NULL AND t.expira_en > $1
		)`, momento)
	if err != nil {
		return 0, err
	}
	return resultado.RowsAffected()
}

type escaneadorUsuario interface {
	Scan(destino ...any) error
}

func escanearUsuario(fila escaneadorUsuario) (*Usuario, error) {
	var usuario Usuario
	var nombreRol string
	var ultimoIngreso sql.NullTime
	if err := fila.Scan(&usuario.ID, &usuario.Email, &usuario.NombreUso, &usuario.ClaveHash, &nombreRol,
		&usuario.FacultadID, &usuario.FacultadNombre,
		&usuario.Activo, &usuario.EmailConfirmado, &usuario.CreadoEn, &ultimoIngreso); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoEncontrado
		}
		return nil, err
	}
	usuario.Rol = Rol(nombreRol)
	if ultimoIngreso.Valid {
		momento := ultimoIngreso.Time
		usuario.UltimoIngreso = &momento
	}
	return &usuario, nil
}
