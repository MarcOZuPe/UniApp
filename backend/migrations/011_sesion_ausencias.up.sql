CREATE TABLE sesion_ausencias (
    sesion_id TEXT NOT NULL REFERENCES sesion_sesiones(id),
    -- El ausente y quien reporta son el mismo participante autenticado.
    usuario_id TEXT NOT NULL CHECK (length(trim(usuario_id)) > 0),
    rol TEXT NOT NULL CHECK (rol IN ('tutor', 'escolar')),
    reportada_en TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (sesion_id, usuario_id)
);
