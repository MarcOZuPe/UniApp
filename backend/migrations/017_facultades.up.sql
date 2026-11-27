-- I18: separación de tutores y coordinación por facultad.
CREATE TABLE facultades (
    id     SMALLINT PRIMARY KEY,
    nombre TEXT NOT NULL UNIQUE CHECK (char_length(trim(nombre)) BETWEEN 3 AND 120)
);

INSERT INTO facultades (id, nombre) VALUES
    (1, 'Facultad de Ingeniería de Sistemas'),
    (2, 'Facultad de Ingeniería Civil'),
    (3, 'Facultad de Ingeniería Industrial');

ALTER TABLE usuarios ADD COLUMN facultad_id SMALLINT REFERENCES facultades(id);

-- Los datos existentes se asignan a la facultad inicial. Los escolares y apoderados
-- quedan sin facultad.
UPDATE usuarios SET facultad_id = 1 WHERE rol_id IN (3, 4, 5);

CREATE FUNCTION usuarios_facultad_coherente() RETURNS trigger AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND OLD.facultad_id IS DISTINCT FROM NEW.facultad_id THEN
        RAISE EXCEPTION 'la facultad no se puede modificar';
    END IF;
    IF NEW.rol_id IN (3, 4, 5) THEN
        IF NEW.facultad_id IS NULL THEN
            NEW.facultad_id := 1;
        END IF;
    ELSE
        NEW.facultad_id := NULL;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER usuarios_facultad_coherente
    BEFORE INSERT OR UPDATE ON usuarios
    FOR EACH ROW EXECUTE FUNCTION usuarios_facultad_coherente();

ALTER TABLE usuarios ADD CONSTRAINT usuarios_facultad_por_rol CHECK (
    (rol_id IN (3, 4, 5) AND facultad_id IS NOT NULL)
    OR
    (rol_id IN (1, 2) AND facultad_id IS NULL)
);

ALTER TABLE incidentes ADD COLUMN facultad_id SMALLINT REFERENCES facultades(id) DEFAULT 1;
UPDATE incidentes i
SET facultad_id = u.facultad_id
FROM usuarios u
WHERE u.id::text = i.tutor_id;
ALTER TABLE incidentes ALTER COLUMN facultad_id SET NOT NULL;
CREATE INDEX incidentes_facultad_idx ON incidentes (facultad_id, estado, creado_en DESC, id);
