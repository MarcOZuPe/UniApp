-- Flujo D · I09: reserva transaccional y adjuntos protegidos.
-- La exclusión y la unicidad se resuelven en la base, no en la aplicación.
CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE reservas ADD COLUMN clave_idempotencia TEXT;

CREATE UNIQUE INDEX reservas_idempotencia_uq
    ON reservas (escolar_id, clave_idempotencia)
    WHERE clave_idempotencia IS NOT NULL;

-- Un escolar no puede tener dos sesiones confirmadas que se crucen (D07).
ALTER TABLE reservas ADD CONSTRAINT reservas_sin_solapamiento
    EXCLUDE USING gist (
        escolar_id WITH =,
        tstzrange(inicio, fin, '[)') WITH &&
    ) WHERE (estado = 'confirmada');

ALTER TABLE reserva_adjuntos
    ADD COLUMN estado TEXT NOT NULL DEFAULT 'en_cuarentena',
    ADD COLUMN motivo TEXT NOT NULL DEFAULT '',
    ADD COLUMN huella TEXT,
    ADD COLUMN clave_almacenamiento TEXT,
    ADD COLUMN actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now();

ALTER TABLE reserva_adjuntos ADD CONSTRAINT reserva_adjuntos_estado
    CHECK (estado IN ('en_cuarentena', 'validado', 'rechazado'));

CREATE UNIQUE INDEX reserva_adjuntos_clave_uq
    ON reserva_adjuntos (clave_almacenamiento)
    WHERE clave_almacenamiento IS NOT NULL;
