DROP TABLE IF EXISTS logs_sistema;
DROP FUNCTION IF EXISTS logs_sistema_impedir_modificacion();

ALTER TABLE auditoria_eventos DROP CONSTRAINT IF EXISTS auditoria_eventos_actor_rol_check;
ALTER TABLE auditoria_eventos ADD CONSTRAINT auditoria_eventos_actor_rol_check
    CHECK (actor_rol IN ('escolar', 'apoderado', 'tutor', 'coordinador', 'coordinador_general'));

ALTER TABLE usuarios DROP CONSTRAINT IF EXISTS usuarios_facultad_por_rol;
ALTER TABLE usuarios ADD CONSTRAINT usuarios_facultad_por_rol CHECK (
    (rol_id IN (3, 4, 5) AND facultad_id IS NOT NULL)
    OR
    (rol_id IN (1, 2) AND facultad_id IS NULL)
);

DELETE FROM roles WHERE id = 6 AND nombre = 'administrador'
    AND NOT EXISTS (SELECT 1 FROM usuarios WHERE rol_id = 6);
