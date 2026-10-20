-- Flujo D · Reserva base (I04). Solo tablas del flujo D (ADR 0001).
-- Los identificadores del tutor y del slot vienen del flujo C (oferta) y se guardan como
-- copia del horario reservado; no hay clave foránea contra tablas de otro flujo.
CREATE TABLE reservas (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    escolar_id   UUID NOT NULL REFERENCES usuarios(id),
    tutor_id     TEXT NOT NULL,
    tutor_nombre TEXT NOT NULL,
    slot_id      TEXT NOT NULL,
    curso        TEXT NOT NULL,
    grado        TEXT NOT NULL,
    inicio       TIMESTAMPTZ NOT NULL,
    fin          TIMESTAMPTZ NOT NULL,
    tema         TEXT NOT NULL,
    objetivo     TEXT NOT NULL,
    estado       TEXT NOT NULL DEFAULT 'confirmada',
    creada_en    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT reservas_horario CHECK (fin > inicio),
    CONSTRAINT reservas_sesion_60_min CHECK (fin = inicio + INTERVAL '60 minutes'),
    CONSTRAINT reservas_tema CHECK (char_length(tema) BETWEEN 10 AND 200),
    CONSTRAINT reservas_objetivo CHECK (char_length(objetivo) BETWEEN 20 AND 1000),
    CONSTRAINT reservas_grado CHECK (char_length(btrim(grado)) BETWEEN 1 AND 80),
    CONSTRAINT reservas_tutor CHECK (btrim(tutor_id) <> ''),
    CONSTRAINT reservas_slot CHECK (btrim(slot_id) <> ''),
    CONSTRAINT reservas_estado CHECK (estado IN ('confirmada'))
);

-- Un horario publicado solo puede alojar una sesión. I09 añadirá la reserva
-- transaccional con pruebas de carreras; la unicidad ya protege el dato.
CREATE UNIQUE INDEX reservas_slot_uq ON reservas (slot_id);
CREATE INDEX reservas_escolar_idx ON reservas (escolar_id, inicio);

CREATE TABLE reserva_adjuntos (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reserva_id   UUID NOT NULL REFERENCES reservas(id) ON DELETE CASCADE,
    nombre       TEXT NOT NULL,
    tipo         TEXT NOT NULL,
    tamano_bytes BIGINT NOT NULL,
    creado_en    TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT reserva_adjuntos_nombre CHECK (char_length(nombre) BETWEEN 1 AND 255),
    CONSTRAINT reserva_adjuntos_tamano CHECK (tamano_bytes > 0 AND tamano_bytes <= 10485760)
);

CREATE INDEX reserva_adjuntos_reserva_idx ON reserva_adjuntos (reserva_id);
