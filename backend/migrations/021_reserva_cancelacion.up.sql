-- Flujo D · I16: cancelacion con historial.
ALTER TABLE reservas DROP CONSTRAINT reservas_estado;
ALTER TABLE reservas ADD CONSTRAINT reservas_estado
    CHECK (estado IN ('confirmada', 'cancelada'));
ALTER TABLE reservas ADD COLUMN cancelada_en TIMESTAMPTZ;
ALTER TABLE reservas ADD COLUMN motivo_cancelacion TEXT NOT NULL DEFAULT '';
ALTER TABLE reservas ADD CONSTRAINT reservas_cancelacion_coherente CHECK (
    (estado = 'confirmada' AND cancelada_en IS NULL AND motivo_cancelacion = '')
    OR (estado = 'cancelada' AND cancelada_en IS NOT NULL)
);

CREATE TABLE reserva_historial (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reserva_id      UUID NOT NULL REFERENCES reservas(id) ON DELETE CASCADE,
    estado_anterior TEXT NOT NULL,
    estado_nuevo    TEXT NOT NULL,
    motivo          TEXT NOT NULL DEFAULT '',
    actor_id        UUID NOT NULL REFERENCES usuarios(id),
    creado_en       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX reserva_historial_reserva_idx ON reserva_historial (reserva_id, creado_en);
