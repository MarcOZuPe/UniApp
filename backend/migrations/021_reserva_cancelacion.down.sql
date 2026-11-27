DROP TABLE IF EXISTS reserva_historial;
ALTER TABLE reservas DROP CONSTRAINT IF EXISTS reservas_cancelacion_coherente;
ALTER TABLE reservas DROP COLUMN IF EXISTS motivo_cancelacion;
ALTER TABLE reservas DROP COLUMN IF EXISTS cancelada_en;
ALTER TABLE reservas DROP CONSTRAINT IF EXISTS reservas_estado;
ALTER TABLE reservas ADD CONSTRAINT reservas_estado CHECK (estado IN ('confirmada'));
