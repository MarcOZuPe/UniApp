ALTER TABLE reserva_adjuntos DROP CONSTRAINT IF EXISTS reserva_adjuntos_estado;
DROP INDEX IF EXISTS reserva_adjuntos_clave_uq;
ALTER TABLE reserva_adjuntos
    DROP COLUMN IF EXISTS estado,
    DROP COLUMN IF EXISTS motivo,
    DROP COLUMN IF EXISTS huella,
    DROP COLUMN IF EXISTS clave_almacenamiento,
    DROP COLUMN IF EXISTS actualizado_en;

ALTER TABLE reservas DROP CONSTRAINT IF EXISTS reservas_sin_solapamiento;
DROP INDEX IF EXISTS reservas_idempotencia_uq;
ALTER TABLE reservas DROP COLUMN IF EXISTS clave_idempotencia;
