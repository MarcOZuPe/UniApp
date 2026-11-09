CREATE TABLE sesion_resultados (
    sesion_id TEXT PRIMARY KEY REFERENCES sesion_sesiones(id),
    usuario_id TEXT NOT NULL CHECK (length(trim(usuario_id)) > 0),
    resumen TEXT NOT NULL CHECK (char_length(trim(resumen)) BETWEEN 1 AND 2000),
    registrado_en TIMESTAMPTZ NOT NULL
);
