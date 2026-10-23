-- Cada fila es una entrega independiente; el canal interno sirve también de bandeja.
CREATE TABLE sesion_outbox (
    id BIGSERIAL PRIMARY KEY,
    reserva_id UUID NOT NULL REFERENCES reservas(id),
    destinatario_id TEXT NOT NULL,
    tipo TEXT NOT NULL CHECK (tipo IN ('confirmacion', 'recordatorio')),
    canal TEXT NOT NULL CHECK (canal IN ('interno', 'correo')),
    asunto TEXT NOT NULL,
    cuerpo TEXT NOT NULL,
    disponible_en TIMESTAMPTZ NOT NULL,
    estado TEXT NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'procesando', 'entregado', 'agotado')),
    intentos INTEGER NOT NULL DEFAULT 0 CHECK (intentos >= 0),
    bloqueado_hasta TIMESTAMPTZ,
    ultimo_error TEXT NOT NULL DEFAULT '',
    entregado_en TIMESTAMPTZ,
    UNIQUE (reserva_id, destinatario_id, tipo, canal)
);
CREATE INDEX sesion_outbox_pendientes ON sesion_outbox (disponible_en, id) WHERE estado IN ('pendiente', 'procesando');
CREATE INDEX sesion_outbox_bandeja ON sesion_outbox (destinatario_id, entregado_en DESC, id DESC) WHERE canal = 'interno' AND estado = 'entregado';
CREATE INDEX reservas_tutor_agenda ON reservas (tutor_id, inicio);
