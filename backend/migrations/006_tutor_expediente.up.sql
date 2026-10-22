CREATE TABLE tutor_expedientes (
    tutor_id UUID PRIMARY KEY REFERENCES usuarios(id) ON DELETE CASCADE,
    presentacion TEXT NOT NULL,
    cursos TEXT NOT NULL,
    experiencia TEXT NOT NULL,
    estado TEXT NOT NULL CHECK (estado IN ('pendiente', 'aprobado', 'rechazado')),
    presentado_en TIMESTAMPTZ NOT NULL,
    vence_en TIMESTAMPTZ,
    revisado_por UUID REFERENCES usuarios(id),
    revisado_en TIMESTAMPTZ,
    motivo_rechazo TEXT NOT NULL DEFAULT '',
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT tutor_aprobado_tiene_vencimiento CHECK (estado <> 'aprobado' OR vence_en IS NOT NULL),
    CONSTRAINT tutor_revision_coherente CHECK (
        (estado = 'pendiente' AND revisado_por IS NULL AND revisado_en IS NULL AND vence_en IS NULL)
        OR (estado = 'aprobado' AND revisado_por IS NOT NULL AND revisado_en IS NOT NULL)
        OR (estado = 'rechazado' AND revisado_por IS NOT NULL AND revisado_en IS NOT NULL AND length(trim(motivo_rechazo)) > 0)
    )
);

CREATE INDEX tutor_expedientes_revision_idx ON tutor_expedientes (presentado_en) WHERE estado = 'pendiente';

CREATE TABLE tutor_documentos (
    id UUID PRIMARY KEY,
    tutor_id UUID NOT NULL REFERENCES tutor_expedientes(tutor_id) ON DELETE CASCADE,
    nombre TEXT NOT NULL,
    tipo TEXT NOT NULL CHECK (tipo = 'application/pdf'),
    tamano_bytes BIGINT NOT NULL CHECK (tamano_bytes BETWEEN 6 AND 10485760),
    creado_en TIMESTAMPTZ NOT NULL
);

CREATE INDEX tutor_documentos_tutor_idx ON tutor_documentos (tutor_id, creado_en);
