ALTER TABLE auditoria_eventos
    ADD CONSTRAINT auditoria_eventos_actor_id_fkey
    FOREIGN KEY (actor_id) REFERENCES usuarios(id);
