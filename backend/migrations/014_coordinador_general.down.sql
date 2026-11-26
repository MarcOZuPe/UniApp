DELETE FROM roles
WHERE id = 5
  AND nombre = 'coordinador_general'
  AND NOT EXISTS (SELECT 1 FROM usuarios WHERE rol_id = 5);

ALTER TABLE auditoria_eventos
    DROP CONSTRAINT IF EXISTS auditoria_eventos_actor_rol_check;

ALTER TABLE auditoria_eventos
    ADD CONSTRAINT auditoria_eventos_actor_rol_check
    CHECK (actor_rol IN ('escolar', 'apoderado', 'tutor', 'coordinador'));
