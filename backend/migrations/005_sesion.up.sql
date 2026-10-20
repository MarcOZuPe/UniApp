CREATE TABLE sesion_sesiones (
    id TEXT PRIMARY KEY,
    reserva_id TEXT NOT NULL UNIQUE,
    tutor_id TEXT NOT NULL,
    escolar_id TEXT NOT NULL,
    inicio TIMESTAMPTZ NOT NULL,
    fin TIMESTAMPTZ NOT NULL,

    CONSTRAINT chk_sesion_identificadores
        CHECK (
            length(trim(id)) > 0
            AND length(trim(reserva_id)) > 0
            AND length(trim(tutor_id)) > 0
            AND length(trim(escolar_id)) > 0
        ),

    CONSTRAINT chk_sesion_participantes
        CHECK (tutor_id <> escolar_id),

    CONSTRAINT chk_sesion_duracion
        CHECK (fin = inicio + INTERVAL '60 minutes')
);

CREATE TABLE sesion_asistencias (
    sesion_id TEXT NOT NULL,
    usuario_id TEXT NOT NULL,
    rol TEXT NOT NULL,
    declarada_en TIMESTAMPTZ NOT NULL,

    CONSTRAINT pk_sesion_asistencias
        PRIMARY KEY (sesion_id, usuario_id),

    CONSTRAINT fk_asistencia_sesion
        FOREIGN KEY (sesion_id)
        REFERENCES sesion_sesiones(id),

    CONSTRAINT chk_asistencia_usuario
        CHECK (length(trim(usuario_id)) > 0),

    CONSTRAINT chk_asistencia_rol
        CHECK (rol IN ('tutor', 'escolar'))
);
