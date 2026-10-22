CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE cupos_semanales (
    tutor_id       UUID PRIMARY KEY REFERENCES usuarios(id) ON DELETE CASCADE,
    minutos        SMALLINT NOT NULL,
    actualizado_en TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT cupos_semanales_rango
        CHECK (minutos BETWEEN 180 AND 360 AND minutos % 60 = 0)
);

CREATE TABLE slots (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tutor_id       UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    dia_semana     SMALLINT NOT NULL,
    inicio_minuto  SMALLINT NOT NULL,
    fin_minuto     SMALLINT NOT NULL,
    zona_horaria   TEXT NOT NULL DEFAULT 'America/Lima',
    creado_en      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT slots_dia_semana_rango CHECK (dia_semana BETWEEN 0 AND 6),
    CONSTRAINT slots_hora_rango CHECK (
        inicio_minuto >= 0 AND inicio_minuto < 1440
        AND fin_minuto > 0 AND fin_minuto <= 1440
    ),
    CONSTRAINT slots_duracion_60_minutos CHECK (fin_minuto - inicio_minuto = 60),
    CONSTRAINT slots_zona_horaria_lima CHECK (zona_horaria = 'America/Lima'),
    CONSTRAINT slots_sin_solapamiento EXCLUDE USING gist (
        tutor_id WITH =,
        dia_semana WITH =,
        (int4range(inicio_minuto, fin_minuto, '[)')) WITH &&
    )
);

CREATE INDEX slots_tutor_dia_idx ON slots (tutor_id, dia_semana);
