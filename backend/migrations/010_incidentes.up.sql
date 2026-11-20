-- I13 · Incidentes confidenciales del flujo A.
-- La evidencia vive fuera de la base de datos; aquí solo se guardan metadatos y una
-- clave opaca. El acceso se controla en la aplicación por rol y relación con el escolar.
CREATE TABLE incidentes (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    escolar_id           UUID NOT NULL REFERENCES usuarios(id),
    reportado_por        UUID NOT NULL REFERENCES usuarios(id),
    tutor_id             TEXT NOT NULL,
    tutor_nombre         TEXT NOT NULL,
    reserva_id           UUID NOT NULL REFERENCES reservas(id),
    curso                TEXT NOT NULL,
    inicio_sesion        TIMESTAMPTZ NOT NULL,
    tipo                 TEXT NOT NULL CHECK (tipo IN (
                             'trato_inadecuado', 'discriminacion', 'acoso', 'seguridad', 'otro'
                         )),
    descripcion          TEXT NOT NULL CHECK (char_length(trim(descripcion)) BETWEEN 20 AND 2000),
    estado               TEXT NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'revisado')),
    resultado            TEXT CHECK (resultado IN ('procede', 'no_procede')),
    comentario_revision  TEXT NOT NULL DEFAULT '',
    revisado_por         UUID REFERENCES usuarios(id),
    revisado_en          TIMESTAMPTZ,
    creado_en            TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT incidente_revision_coherente CHECK (
        (estado = 'pendiente' AND resultado IS NULL AND revisado_por IS NULL
            AND revisado_en IS NULL AND comentario_revision = '')
        OR
        (estado = 'revisado' AND resultado IS NOT NULL AND revisado_por IS NOT NULL
            AND revisado_en IS NOT NULL AND char_length(trim(comentario_revision)) > 0)
    )
);

CREATE INDEX incidentes_escolar_idx
    ON incidentes (escolar_id, creado_en DESC, id);

CREATE INDEX incidentes_estado_idx
    ON incidentes (estado, creado_en DESC, id);

CREATE INDEX incidentes_reserva_idx
    ON incidentes (reserva_id);

CREATE TABLE incidente_evidencias (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    incidente_id          UUID NOT NULL REFERENCES incidentes(id) ON DELETE CASCADE,
    nombre                TEXT NOT NULL CHECK (char_length(trim(nombre)) BETWEEN 1 AND 255),
    tipo                  TEXT NOT NULL CHECK (tipo IN ('application/pdf', 'image/jpeg', 'image/png')),
    tamano_bytes          BIGINT NOT NULL CHECK (tamano_bytes > 0 AND tamano_bytes <= 10485760),
    clave_almacenamiento  TEXT NOT NULL UNIQUE CHECK (char_length(clave_almacenamiento) = 32),
    creado_en             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX incidente_evidencias_incidente_idx
    ON incidente_evidencias (incidente_id, creado_en, id);
