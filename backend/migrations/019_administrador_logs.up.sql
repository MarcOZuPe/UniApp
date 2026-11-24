-- Administrador del sistema y bitácora de seguridad/administración.
INSERT INTO roles (id, nombre) VALUES (6, 'administrador');

ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_facultad_por_rol;
ALTER TABLE usuarios ADD CONSTRAINT usuarios_facultad_por_rol CHECK (
    (rol_id IN (3, 4, 5) AND facultad_id IS NOT NULL)
    OR
    (rol_id IN (1, 2, 6) AND facultad_id IS NULL)
);

ALTER TABLE auditoria_eventos DROP CONSTRAINT IF EXISTS auditoria_eventos_actor_rol_check;
ALTER TABLE auditoria_eventos ADD CONSTRAINT auditoria_eventos_actor_rol_check
    CHECK (actor_rol IN ('escolar', 'apoderado', 'tutor', 'coordinador', 'coordinador_general', 'administrador'));

CREATE TABLE logs_sistema (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    creado_en   TIMESTAMPTZ NOT NULL DEFAULT now(),
    categoria   TEXT NOT NULL CHECK (categoria IN ('seguridad', 'administracion', 'sistema')),
    evento      TEXT NOT NULL CHECK (char_length(trim(evento)) BETWEEN 1 AND 120),
    actor_id    UUID,
    actor_rol   TEXT NOT NULL CHECK (actor_rol IN (
        'escolar', 'apoderado', 'tutor', 'coordinador', 'coordinador_general', 'administrador'
    )),
    ip          TEXT NOT NULL DEFAULT '' CHECK (char_length(ip) <= 128),
    detalle     JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT logs_sistema_detalle_objeto CHECK (jsonb_typeof(detalle) = 'object')
);

CREATE INDEX logs_sistema_fecha_idx ON logs_sistema (creado_en DESC, id DESC);
CREATE INDEX logs_sistema_categoria_idx ON logs_sistema (categoria, creado_en DESC, id DESC);

CREATE FUNCTION logs_sistema_impedir_modificacion() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'la bitácora del sistema es inmutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER logs_sistema_sin_update_delete
    BEFORE UPDATE OR DELETE ON logs_sistema
    FOR EACH ROW EXECUTE FUNCTION logs_sistema_impedir_modificacion();

CREATE TRIGGER logs_sistema_sin_truncate
    BEFORE TRUNCATE ON logs_sistema
    FOR EACH STATEMENT EXECUTE FUNCTION logs_sistema_impedir_modificacion();
