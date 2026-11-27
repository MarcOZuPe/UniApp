-- Reversión de 020: vuelve al catálogo sintético de 018.
DROP INDEX IF EXISTS facultades_siglas_uq;

DELETE FROM facultades WHERE id BETWEEN 5 AND 11;

UPDATE facultades SET nombre = 'Facultad de Ciencias' WHERE id = 1;
UPDATE facultades SET nombre = 'Facultad de Ingeniería Industrial y Sistemas' WHERE id = 2;
UPDATE facultades SET nombre = 'Facultad de Ingeniería Mecánica' WHERE id = 3;
UPDATE facultades SET nombre = 'Facultad de Arquitectura, Urbanismo y Artes' WHERE id = 4;

ALTER TABLE facultades DROP COLUMN IF EXISTS especialidades;
ALTER TABLE facultades DROP COLUMN IF EXISTS siglas;
