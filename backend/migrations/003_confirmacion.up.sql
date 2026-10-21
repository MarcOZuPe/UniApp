ALTER TABLE usuarios ADD COLUMN correo_confirmado BOOLEAN NOT NULL DEFAULT TRUE;

CREATE TABLE tokens_confirmacion (
    token_hash BYTEA PRIMARY KEY,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    expira_en  TIMESTAMPTZ NOT NULL,
    usado_en   TIMESTAMPTZ
);
CREATE INDEX tokens_confirmacion_usuario_idx ON tokens_confirmacion (usuario_id);
