# I14 · Panel de coordinación y auditoría

## Objetivo

Dar a coordinación una vista agregada del piloto (RF-22) y un registro inmutable de
las decisiones y cambios de reglas. El panel no expone secretos y la auditoría no
puede modificarse ni borrarse desde la aplicación.

## Alcance implementado

- `/coordinacion` muestra métricas agregadas de:
  - reservas totales, próximas y finalizadas;
  - sesiones finalizadas con asistencia declarada por escolar y tutor;
  - horas internas de sesiones finalizadas;
  - pendientes de expedientes, incidentes, cursos y solicitudes de colegio.
- La API del panel vive en `/api/coordinacion/metricas` y `/api/coordinacion/auditoria`.
- La auditoría registra automáticamente:
  - revisión de incidentes;
  - revisión de expedientes de tutores;
  - revisión de cursos del catálogo;
  - cambios de autorización del apoderado;
  - cambios de cupo semanal del tutor.
- La tabla `auditoria_eventos` es append-only. La base rechaza `UPDATE`, `DELETE` y
  `TRUNCATE`; no existe endpoint de borrado.
- El detalle auditado se limita a datos de negocio. No guarda tokens, claves, correos,
  hashes, huellas ni contenido de evidencias.

## API

| Método | Ruta | Permiso | Uso |
|---|---|---|---|
| GET | `/api/coordinacion/metricas` | `coordinacion:acceder` | Métricas agregadas del piloto |
| GET | `/api/coordinacion/auditoria?limite=50` | `coordinacion:acceder` | Últimos eventos de auditoría |

El contrato está en `docs/openapi.yaml`, etiqueta `I14-panel`.

## Métricas

- **Reservas:** total, futuras (`inicio > ahora`) y finalizadas (`fin <= ahora`).
- **Asistencia:** sesiones finalizadas y cuántas tienen asistencia declarada por el
  escolar o por el tutor; incluye porcentajes.
- **Horas internas:** suma de la duración de sesiones finalizadas, expresada en horas.
- **Gestión:** pendientes operativos de expedientes, incidentes, cursos y solicitudes
  de colegio.

## Auditoría

Migración `backend/migrations/011_auditoria_panel.up.sql`.

Tabla `auditoria_eventos`:

- `ocurrido_en`, `actor_id`, `actor_rol`, `accion`, `entidad`, `entidad_id`, `detalle`.
- El actor se toma de `revisado_por`, `apoderado_id` o `tutor_id` según el cambio.
- Los triggers solo insertan cuando hay una transición real de estado, por lo que una
  siembra idempotente no duplica decisiones.

La interfaz muestra la fecha en hora de Lima, la acción, el actor, la entidad y un
resumen saneado del detalle.

## Capas

- Dominio: `backend/internal/domain/panel`
- Aplicación: `backend/internal/application/panel`
- HTTP: `backend/internal/adapters/http/panel`
- Presentación: `backend/internal/adapters/presentation/panel`
- PostgreSQL: `backend/internal/adapters/repositories/postgres/panel`
- Interfaz: `frontend/src/presentation/pages/acceso/Coordinacion.jsx`
- Cliente: `frontend/src/application/acceso/coordinacion.js`

## Pruebas

- Aplicación: solo coordinación consulta métricas/auditoría y se aplican límites.
- PostgreSQL: métricas, triggers de auditoría, roles de actor, ausencia de campos
  sensibles e inmutabilidad frente a `UPDATE`, `DELETE` y `TRUNCATE`.
- Frontend: render de métricas y auditoría saneada.

## Límites

- La auditoría comienza con la migración `011`; no reconstruye decisiones históricas
  anteriores.
- Cohorte y calendario no tienen todavía métricas propias; se incorporarán cuando sus
  modelos estén disponibles.
