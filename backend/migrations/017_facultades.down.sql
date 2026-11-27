DROP INDEX IF EXISTS incidentes_facultad_idx;
ALTER TABLE incidentes DROP COLUMN IF EXISTS facultad_id;

ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_facultad_por_rol;
DROP TRIGGER IF EXISTS usuarios_facultad_coherente ON usuarios;
DROP FUNCTION IF EXISTS usuarios_facultad_coherente();
ALTER TABLE usuarios DROP COLUMN IF EXISTS facultad_id;

DROP TABLE IF EXISTS facultades;
