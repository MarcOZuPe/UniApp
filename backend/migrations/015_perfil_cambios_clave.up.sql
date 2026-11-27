-- Historial mínimo para limitar cambios de contraseña por usuario.
CREATE TABLE perfil_cambios_clave (
    id          BIGSERIAL PRIMARY KEY,
    usuario_id  UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    cambiado_en TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX perfil_cambios_clave_usuario_fecha_idx
    ON perfil_cambios_clave (usuario_id, cambiado_en DESC);
