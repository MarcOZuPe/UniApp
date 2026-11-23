-- El actor de auditoría queda como identificador histórico, sin FK que obligue a
-- borrar eventos al eliminar una cuenta. La auditoría sigue siendo inmutable.
ALTER TABLE auditoria_eventos
    DROP CONSTRAINT IF EXISTS auditoria_eventos_actor_id_fkey;
