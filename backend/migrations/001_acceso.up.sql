CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE roles (
    id     SMALLINT PRIMARY KEY,
    nombre TEXT NOT NULL UNIQUE
);
INSERT INTO roles (id, nombre) VALUES
    (1, 'escolar'), (2, 'apoderado'), (3, 'tutor'), (4, 'coordinador');

CREATE TABLE usuarios (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          TEXT NOT NULL,
    nombre_uso     TEXT NOT NULL,
    clave_hash     TEXT NOT NULL,
    rol_id         SMALLINT NOT NULL REFERENCES roles(id),
    activo         BOOLEAN NOT NULL DEFAULT TRUE,
    creado_en      TIMESTAMPTZ NOT NULL DEFAULT now(),
    ultimo_ingreso TIMESTAMPTZ,
    CONSTRAINT usuarios_email_minuscula CHECK (email = lower(email))
);
CREATE UNIQUE INDEX usuarios_email_uq ON usuarios (email);

CREATE TABLE sesiones (
    token_hash BYTEA PRIMARY KEY,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    creada_en  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expira_en  TIMESTAMPTZ NOT NULL
);
CREATE INDEX sesiones_usuario_idx ON sesiones (usuario_id);
