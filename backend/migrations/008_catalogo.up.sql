CREATE TABLE catalogo_importaciones (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    fuente       TEXT NOT NULL CHECK (length(trim(fuente)) BETWEEN 1 AND 120),
    importado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE catalogo_cursos (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    importacion_id  BIGINT NOT NULL REFERENCES catalogo_importaciones(id) ON DELETE CASCADE,
    nombre          TEXT NOT NULL CHECK (length(trim(nombre)) BETWEEN 1 AND 120),
    grado           TEXT NOT NULL CHECK (length(trim(grado)) BETWEEN 1 AND 80),
    estado          TEXT NOT NULL DEFAULT 'pendiente'
                    CHECK (estado IN ('pendiente', 'aprobado', 'rechazado')),
    motivo_rechazo  TEXT NOT NULL DEFAULT '',
    revisado_por    UUID REFERENCES usuarios(id),
    revisado_en     TIMESTAMPTZ,
    CONSTRAINT catalogo_curso_revision_coherente CHECK (
        (estado = 'pendiente' AND revisado_por IS NULL AND revisado_en IS NULL AND motivo_rechazo = '')
        OR (estado = 'aprobado' AND revisado_por IS NOT NULL AND revisado_en IS NOT NULL AND motivo_rechazo = '')
        OR (estado = 'rechazado' AND revisado_por IS NOT NULL AND revisado_en IS NOT NULL AND length(trim(motivo_rechazo)) > 0)
    ),
    CONSTRAINT catalogo_curso_grado_uq UNIQUE (importacion_id, nombre, grado)
);

CREATE INDEX catalogo_cursos_importacion_estado_idx
    ON catalogo_cursos (importacion_id, estado, grado, nombre);

CREATE TABLE catalogo_colegios (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    importacion_id  BIGINT NOT NULL REFERENCES catalogo_importaciones(id) ON DELETE CASCADE,
    nombre          TEXT NOT NULL CHECK (length(trim(nombre)) BETWEEN 1 AND 160),
    distrito        TEXT NOT NULL CHECK (length(trim(distrito)) BETWEEN 1 AND 100),
    CONSTRAINT catalogo_colegio_distrito_uq UNIQUE (importacion_id, nombre, distrito)
);

CREATE INDEX catalogo_colegios_importacion_idx ON catalogo_colegios (importacion_id, distrito, nombre);

CREATE TABLE catalogo_solicitudes_colegio (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nombre          TEXT NOT NULL CHECK (length(trim(nombre)) BETWEEN 1 AND 160),
    distrito        TEXT NOT NULL CHECK (length(trim(distrito)) BETWEEN 1 AND 100),
    solicitante_id  UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    estado          TEXT NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'incorporada')),
    solicitado_en  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX catalogo_solicitud_colegio_pendiente_uq
    ON catalogo_solicitudes_colegio (lower(trim(nombre)), lower(trim(distrito)))
    WHERE estado = 'pendiente';
CREATE INDEX catalogo_solicitudes_colegio_pendiente_idx
    ON catalogo_solicitudes_colegio (solicitado_en, id)
    WHERE estado = 'pendiente';

INSERT INTO catalogo_importaciones (fuente)
VALUES ('sintetico: fixture inicial I07');

INSERT INTO catalogo_cursos (importacion_id, nombre, grado)
SELECT id, datos.nombre, datos.grado
FROM (SELECT id FROM catalogo_importaciones ORDER BY id DESC LIMIT 1) importacion
CROSS JOIN (VALUES
    ('Matemática básica', '1.° de secundaria'),
    ('Comunicación', '1.° de secundaria'),
    ('Ciencias naturales', '2.° de secundaria')
) AS datos(nombre, grado);

INSERT INTO catalogo_colegios (importacion_id, nombre, distrito)
SELECT id, datos.nombre, datos.distrito
FROM (SELECT id FROM catalogo_importaciones ORDER BY id DESC LIMIT 1) importacion
CROSS JOIN (VALUES
    ('Colegio Sintético 001', 'Distrito Sintético'),
    ('Colegio Sintético 002', 'Distrito Sintético')
) AS datos(nombre, distrito);
