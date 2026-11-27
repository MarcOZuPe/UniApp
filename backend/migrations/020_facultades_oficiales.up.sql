-- 020 · Catálogo oficial de facultades de la universidad.
-- Mantiene los IDs 1–4 usados por las cuentas sintéticas existentes:
--   1 Ciencias, 2 Industrial y Sistemas, 3 Mecánica, 4 Arquitectura.
-- Agrega las facultades restantes con sus siglas y especialidades.

ALTER TABLE facultades ADD COLUMN IF NOT EXISTS siglas TEXT;
ALTER TABLE facultades ADD COLUMN IF NOT EXISTS especialidades TEXT[] NOT NULL DEFAULT '{}';

INSERT INTO facultades (id, nombre, siglas, especialidades) VALUES
    (1,  'Facultad de Ciencias', 'FC',
     ARRAY['Física', 'Matemática', 'Química', 'Ingeniería Física', 'Ciencia de la Computación']),
    (2,  'Facultad de Ingeniería Industrial y de Sistemas', 'FIIS',
     ARRAY['Ingeniería Industrial', 'Ingeniería de Sistemas', 'Ingeniería de Software', 'Ingeniería de Inteligencia Artificial']),
    (3,  'Facultad de Ingeniería Mecánica', 'FIM',
     ARRAY['Ingeniería Mecánica', 'Ingeniería Mecánica y Eléctrica', 'Ingeniería Naval', 'Ingeniería Mecatrónica', 'Ingeniería Aeroespacial']),
    (4,  'Facultad de Arquitectura, Urbanismo y Artes', 'FAUA',
     ARRAY['Arquitectura', 'Urbanismo']),
    (5,  'Facultad de Ingeniería Ambiental', 'FIA',
     ARRAY['Ingeniería Sanitaria', 'Ingeniería de Higiene y Seguridad Industrial', 'Ingeniería Ambiental']),
    (6,  'Facultad de Ingeniería Civil', 'FIC',
     ARRAY['Ingeniería Civil']),
    (7,  'Facultad de Ingeniería Económica, Estadística y Ciencias Sociales', 'FIEECS',
     ARRAY['Ingeniería Económica', 'Ingeniería Estadística']),
    (8,  'Facultad de Ingeniería Eléctrica y Electrónica', 'FIEE',
     ARRAY['Ingeniería Eléctrica', 'Ingeniería Electrónica', 'Ingeniería de Telecomunicaciones', 'Ingeniería de Ciberseguridad', 'Ingeniería Biomédica']),
    (9,  'Facultad de Ingeniería Geológica, Minera y Metalúrgica', 'FIGMM',
     ARRAY['Ingeniería Geológica', 'Ingeniería Metalúrgica', 'Ingeniería de Minas']),
    (10, 'Facultad de Ingeniería de Petróleo, Gas Natural y Petroquímica', 'FIP',
     ARRAY['Ingeniería Petroquímica', 'Ingeniería de Petróleo y Gas Natural']),
    (11, 'Facultad de Ingeniería Química y Textil', 'FIQT',
     ARRAY['Ingeniería Química', 'Ingeniería Textil'])
ON CONFLICT (id) DO UPDATE SET
    nombre = EXCLUDED.nombre,
    siglas = EXCLUDED.siglas,
    especialidades = EXCLUDED.especialidades;

ALTER TABLE facultades ALTER COLUMN siglas SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS facultades_siglas_uq ON facultades (siglas);
