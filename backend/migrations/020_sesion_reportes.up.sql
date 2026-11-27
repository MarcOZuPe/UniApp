-- I18: constancia de cada reporte de minutos descargado (flujo E).
-- Solo se inserta; el trigger impide modificarla o borrarla.
CREATE TABLE sesion_reportes (
    id TEXT PRIMARY KEY CHECK (length(trim(id)) > 0),
    generado_por TEXT NOT NULL CHECK (length(trim(generado_por)) > 0),
    generado_en TIMESTAMPTZ NOT NULL,
    desde TIMESTAMPTZ NOT NULL,
    hasta TIMESTAMPTZ NOT NULL,
    tutor_id TEXT NOT NULL DEFAULT '',
    formato TEXT NOT NULL CHECK (formato IN ('csv', 'pdf')),
    filas INTEGER NOT NULL CHECK (filas >= 0),
    sha256 TEXT NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT chk_sesion_reportes_periodo CHECK (desde < hasta)
);
CREATE INDEX sesion_reportes_recientes ON sesion_reportes (generado_en DESC);
-- Consulta de sesiones por periodo (I18).
CREATE INDEX sesion_sesiones_inicio ON sesion_sesiones (inicio, tutor_id);

CREATE FUNCTION sesion_reportes_inmutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'la constancia de reportes es inmutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sesion_reportes_sin_cambios
    BEFORE UPDATE OR DELETE ON sesion_reportes
    FOR EACH ROW EXECUTE FUNCTION sesion_reportes_inmutable();
