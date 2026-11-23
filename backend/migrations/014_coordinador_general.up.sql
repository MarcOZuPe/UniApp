-- Rol de coordinación general: mismos permisos que coordinación y gestión de
-- cuentas de coordinación. La auditoría debe aceptar el nuevo rol como actor.
INSERT INTO roles (id, nombre) VALUES (5, 'coordinador_general');

ALTER TABLE auditoria_eventos
    DROP CONSTRAINT IF EXISTS auditoria_eventos_actor_rol_check;

ALTER TABLE auditoria_eventos
    ADD CONSTRAINT auditoria_eventos_actor_rol_check
    CHECK (actor_rol IN ('escolar', 'apoderado', 'tutor', 'coordinador', 'coordinador_general'));
