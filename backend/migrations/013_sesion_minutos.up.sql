-- I17: minutos verificados y ajustes auditados (flujo E).
CREATE TABLE sesion_minutos (
    sesion_id TEXT PRIMARY KEY REFERENCES sesion_sesiones(id),
    estado TEXT NOT NULL CHECK (estado IN ('propuesto', 'en_disputa', 'verificado', 'anulado')),
    minutos_propuestos INTEGER NOT NULL CHECK (minutos_propuestos BETWEEN 1 AND 60),
    minutos_verificados INTEGER CHECK (minutos_verificados BETWEEN 0 AND 60),
    disputado_por TEXT NOT NULL DEFAULT '',
    CONSTRAINT chk_minutos_verificado_con_valor
        CHECK (estado <> 'verificado' OR minutos_verificados IS NOT NULL)
);

-- Historia inmutable: solo se inserta; el trigger impide modificarla o borrarla.
CREATE TABLE sesion_minutos_auditoria (
    id BIGSERIAL PRIMARY KEY,
    sesion_id TEXT NOT NULL REFERENCES sesion_minutos(sesion_id),
    accion TEXT NOT NULL CHECK (accion IN ('proponer', 'aprobar', 'ajustar', 'disputar', 'resolver', 'anular')),
    actor_id TEXT NOT NULL CHECK (length(trim(actor_id)) > 0),
    valor_anterior INTEGER CHECK (valor_anterior BETWEEN 0 AND 60),
    valor_nuevo INTEGER CHECK (valor_nuevo BETWEEN 0 AND 60),
    motivo TEXT NOT NULL DEFAULT '' CHECK (char_length(motivo) <= 500),
    registrada_en TIMESTAMPTZ NOT NULL
);
CREATE INDEX sesion_minutos_auditoria_sesion ON sesion_minutos_auditoria (sesion_id, id);
CREATE INDEX sesion_minutos_pendientes ON sesion_minutos (estado) WHERE estado IN ('propuesto', 'en_disputa');

CREATE FUNCTION sesion_minutos_auditoria_inmutable() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'la auditoria de minutos es inmutable';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sesion_minutos_auditoria_sin_cambios
    BEFORE UPDATE OR DELETE ON sesion_minutos_auditoria
    FOR EACH ROW EXECUTE FUNCTION sesion_minutos_auditoria_inmutable();
