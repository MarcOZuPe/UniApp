-- I14 · Panel de coordinación y auditoría de decisiones/cambios de reglas.
-- La auditoría es append-only: no hay endpoints de borrado y la base rechaza
-- UPDATE, DELETE y TRUNCATE sobre sus filas. El detalle se limita a datos de
-- negocio; nunca guarda tokens, claves, correos ni contenido de evidencias.
CREATE TABLE auditoria_eventos (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ocurrido_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    actor_id    UUID REFERENCES usuarios(id),
    actor_rol   TEXT NOT NULL CHECK (actor_rol IN ('escolar', 'apoderado', 'tutor', 'coordinador')),
    accion      TEXT NOT NULL CHECK (char_length(trim(accion)) BETWEEN 1 AND 80),
    entidad     TEXT NOT NULL CHECK (char_length(trim(entidad)) BETWEEN 1 AND 80),
    entidad_id  TEXT NOT NULL CHECK (char_length(trim(entidad_id)) BETWEEN 1 AND 160),
    detalle     JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT auditoria_detalle_objeto CHECK (jsonb_typeof(detalle) = 'object')
);

CREATE INDEX auditoria_eventos_fecha_idx
    ON auditoria_eventos (ocurrido_en DESC, id DESC);
CREATE INDEX auditoria_eventos_entidad_idx
    ON auditoria_eventos (entidad, entidad_id, ocurrido_en DESC, id DESC);

CREATE FUNCTION auditoria_impedir_modificacion() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'la auditoría es inmutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auditoria_eventos_sin_update_delete
    BEFORE UPDATE OR DELETE ON auditoria_eventos
    FOR EACH ROW EXECUTE FUNCTION auditoria_impedir_modificacion();

CREATE TRIGGER auditoria_eventos_sin_truncate
    BEFORE TRUNCATE ON auditoria_eventos
    FOR EACH STATEMENT EXECUTE FUNCTION auditoria_impedir_modificacion();

CREATE FUNCTION auditoria_registrar_incidente() RETURNS trigger AS $$
BEGIN
    IF NEW.estado = 'revisado' AND OLD.estado IS DISTINCT FROM NEW.estado THEN
        INSERT INTO auditoria_eventos (actor_id, actor_rol, accion, entidad, entidad_id, detalle)
        SELECT NEW.revisado_por, r.nombre, 'incidente.revisado', 'incidente', NEW.id::text,
               jsonb_build_object(
                   'estado_anterior', OLD.estado,
                   'estado', NEW.estado,
                   'resultado', NEW.resultado,
                   'tutor_id', NEW.tutor_id
               )
        FROM usuarios u JOIN roles r ON r.id = u.rol_id
        WHERE u.id = NEW.revisado_por;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auditoria_incidente_revision
    AFTER UPDATE ON incidentes
    FOR EACH ROW EXECUTE FUNCTION auditoria_registrar_incidente();

CREATE FUNCTION auditoria_registrar_expediente() RETURNS trigger AS $$
BEGIN
    IF NEW.estado IS DISTINCT FROM OLD.estado AND NEW.revisado_por IS NOT NULL THEN
        INSERT INTO auditoria_eventos (actor_id, actor_rol, accion, entidad, entidad_id, detalle)
        SELECT NEW.revisado_por, r.nombre, 'expediente.revisado', 'tutor_expediente', NEW.tutor_id::text,
               jsonb_build_object(
                   'estado_anterior', OLD.estado,
                   'estado', NEW.estado,
                   'vence_en', NEW.vence_en
               )
        FROM usuarios u JOIN roles r ON r.id = u.rol_id
        WHERE u.id = NEW.revisado_por;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auditoria_expediente_revision
    AFTER UPDATE ON tutor_expedientes
    FOR EACH ROW EXECUTE FUNCTION auditoria_registrar_expediente();

CREATE FUNCTION auditoria_registrar_curso() RETURNS trigger AS $$
BEGIN
    IF NEW.estado IS DISTINCT FROM OLD.estado AND NEW.revisado_por IS NOT NULL THEN
        INSERT INTO auditoria_eventos (actor_id, actor_rol, accion, entidad, entidad_id, detalle)
        SELECT NEW.revisado_por, r.nombre, 'curso.revisado', 'catalogo_curso', NEW.id::text,
               jsonb_build_object(
                   'estado_anterior', OLD.estado,
                   'estado', NEW.estado,
                   'curso', NEW.nombre,
                   'grado', NEW.grado
               )
        FROM usuarios u JOIN roles r ON r.id = u.rol_id
        WHERE u.id = NEW.revisado_por;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auditoria_curso_revision
    AFTER UPDATE ON catalogo_cursos
    FOR EACH ROW EXECUTE FUNCTION auditoria_registrar_curso();

CREATE FUNCTION auditoria_registrar_autorizacion() RETURNS trigger AS $$
BEGIN
    INSERT INTO auditoria_eventos (actor_id, actor_rol, accion, entidad, entidad_id, detalle)
    SELECT NEW.apoderado_id, r.nombre, 'autorizacion.cambiada', 'autorizacion', NEW.id::text,
           jsonb_build_object(
               'escolar_id', NEW.escolar_id,
               'estado', NEW.estado
           )
    FROM usuarios u JOIN roles r ON r.id = u.rol_id
    WHERE u.id = NEW.apoderado_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auditoria_autorizacion_cambio
    AFTER INSERT ON autorizaciones_apoderado
    FOR EACH ROW EXECUTE FUNCTION auditoria_registrar_autorizacion();

CREATE FUNCTION auditoria_registrar_cupo() RETURNS trigger AS $$
DECLARE
    minutos_anteriores SMALLINT;
BEGIN
    IF TG_OP = 'UPDATE' THEN
        minutos_anteriores := OLD.minutos;
        IF OLD.minutos IS NOT DISTINCT FROM NEW.minutos THEN
            RETURN NEW;
        END IF;
    END IF;

    INSERT INTO auditoria_eventos (actor_id, actor_rol, accion, entidad, entidad_id, detalle)
    SELECT NEW.tutor_id, r.nombre, 'cupo.actualizado', 'cupo_semanal', NEW.tutor_id::text,
           jsonb_build_object(
               'minutos_anteriores', minutos_anteriores,
               'minutos', NEW.minutos
           )
    FROM usuarios u JOIN roles r ON r.id = u.rol_id
    WHERE u.id = NEW.tutor_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER auditoria_cupo_insercion
    AFTER INSERT ON cupos_semanales
    FOR EACH ROW EXECUTE FUNCTION auditoria_registrar_cupo();

CREATE TRIGGER auditoria_cupo_actualizacion
    AFTER UPDATE ON cupos_semanales
    FOR EACH ROW EXECUTE FUNCTION auditoria_registrar_cupo();
