-- Catálogo sintético de facultades solicitado para el piloto.
UPDATE facultades SET nombre = 'Facultad de Ciencias' WHERE id = 1;
UPDATE facultades SET nombre = 'Facultad de Ingeniería Industrial y Sistemas' WHERE id = 2;
UPDATE facultades SET nombre = 'Facultad de Ingeniería Mecánica' WHERE id = 3;

INSERT INTO facultades (id, nombre)
VALUES (4, 'Facultad de Arquitectura, Urbanismo y Artes')
ON CONFLICT (id) DO UPDATE SET nombre = EXCLUDED.nombre;
