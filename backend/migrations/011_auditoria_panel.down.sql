DROP TRIGGER IF EXISTS auditoria_cupo_actualizacion ON cupos_semanales;
DROP TRIGGER IF EXISTS auditoria_cupo_insercion ON cupos_semanales;
DROP FUNCTION IF EXISTS auditoria_registrar_cupo();

DROP TRIGGER IF EXISTS auditoria_autorizacion_cambio ON autorizaciones_apoderado;
DROP FUNCTION IF EXISTS auditoria_registrar_autorizacion();

DROP TRIGGER IF EXISTS auditoria_curso_revision ON catalogo_cursos;
DROP FUNCTION IF EXISTS auditoria_registrar_curso();

DROP TRIGGER IF EXISTS auditoria_expediente_revision ON tutor_expedientes;
DROP FUNCTION IF EXISTS auditoria_registrar_expediente();

DROP TRIGGER IF EXISTS auditoria_incidente_revision ON incidentes;
DROP FUNCTION IF EXISTS auditoria_registrar_incidente();

DROP TRIGGER IF EXISTS auditoria_eventos_sin_truncate ON auditoria_eventos;
DROP TRIGGER IF EXISTS auditoria_eventos_sin_update_delete ON auditoria_eventos;
DROP FUNCTION IF EXISTS auditoria_impedir_modificacion();

DROP TABLE IF EXISTS auditoria_eventos;
