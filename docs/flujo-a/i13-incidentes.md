# I13 · Incidentes confidenciales

## Objetivo

Permitir que un escolar, o su apoderado, reporte una posible falta cometida por un
tutor durante una tutoría ya finalizada. El reporte exige evidencia y queda disponible
para revisión de coordinación. Las evidencias son sensibles y no se comparten con el
tutor denunciado.

## Alcance implementado

- Ruta frontend exclusiva `/incidentes` para `escolar`, `apoderado` y `coordinador`.
- Escolar y apoderado registran reportes; coordinación revisa y decide si procede.
- El formulario solo ofrece tutorías confirmadas que ya terminaron.
- El apoderado reporta únicamente por escolares con vinculación aceptada.
- Evidencia obligatoria: entre 1 y 5 archivos, 10 MB por archivo.
- Formatos aceptados: PDF, JPG/JPEG y PNG. Se valida la extensión y el contenido real,
  no solo el `Content-Type` declarado por el navegador.
- Las evidencias se almacenan fuera de PostgreSQL, con nombre opaco y permisos privados.
- La descarga se sirve como adjunto con `Cache-Control: private, no-store`.

## Matriz de visibilidad

| Rol | Lista reportes | Descarga evidencia | Registra | Revisa |
|---|---|---|---|---|
| Escolar afectado | Sí, solo los suyos | Sí | Sí | No |
| Apoderado vinculado | Sí, los de sus escolares | Sí | Sí, por sus escolares | No |
| Coordinación | Sí, todos | Sí | No | Sí |
| Tutor denunciado | No | No | No | No |
| Otro escolar/apoderado | No | No | No | No |

La autorización se comprueba en backend, no solo en la interfaz. El tutor recibe
`403` aunque conozca el identificador del incidente o de la evidencia.

## Modelo de datos

Migración `backend/migrations/010_incidentes.up.sql`:

- `incidentes`: escolar afectado, usuario que reporta, tutor denunciado, reserva previa,
  curso, inicio de la tutoría, tipo, descripción, estado y datos de revisión.
- `incidente_evidencias`: nombre original saneado, tipo MIME permitido, tamaño y clave
  opaca de almacenamiento.

Estados:

- `pendiente`: recién registrado, sin decisión de coordinación.
- `revisado`: tiene `resultado` (`procede` o `no_procede`), comentario, revisor y fecha.

## API

| Método | Ruta | Permiso | Uso |
|---|---|---|---|
| GET | `/api/incidentes/tutorias` | `incidente:acceder` | Tutorías previas del escolar o de los escolares del apoderado |
| GET | `/api/incidentes` | `incidente:acceder` | Reportes visibles para el usuario |
| POST | `/api/incidentes` | `incidente:registrar` | Registro multipart con evidencia |
| GET | `/api/incidentes/{id}` | `incidente:acceder` | Detalle visible |
| GET | `/api/incidentes/{id}/evidencias/{evidenciaId}` | `incidente:acceder` | Descarga confidencial |
| POST | `/api/incidentes/{id}/revision` | `incidente:revisar` | Revisión de coordinación |

El contrato detallado está en `docs/openapi.yaml`, etiqueta `A-incidentes`.

## Capas

- Dominio: `backend/internal/domain/incidente`
- Aplicación: `backend/internal/application/incidente`
- Adaptador HTTP: `backend/internal/adapters/http/incidente`
- Presentación backend: `backend/internal/adapters/presentation/incidente`
- PostgreSQL: `backend/internal/adapters/repositories/postgres/incidente`
- Memoria (pruebas): `backend/internal/adapters/repositories/memory/incidente`
- Interfaz: `frontend/src/presentation/pages/incidente/Incidentes.jsx`
- Cliente HTTP: `frontend/src/adapters/http/incidenteApi.js`
- Validaciones de navegador: `frontend/src/domain/incidente/validadores.js`

La feature `incidente` se agregó a la prueba de arquitectura. Solo depende de `acceso`
como identidad compartida, no de las tablas ni del dominio de reservas.

## Pruebas

- Backend:
  - Dominio: validación de formatos, extensión, tamaño, tutoría previa y revisión.
  - Aplicación: permisos por rol, apoderado vinculado, visibilidad y descarga.
  - HTTP: registro multipart, rechazo de evidencia inválida, `403` para tutor,
    descarga para escolar/apoderado/coordinación, revisión única.
  - PostgreSQL: migración, tutorías con asistencia declarada, persistencia de incidente,
    filtros por escolar/apoderado/coordinación y revisión.
  - Arquitectura Clean.
- Frontend:
  - Reporte con PDF.
  - Evidencia obligatoria y rechazo de archivos no permitidos.
  - Flujo de apoderado con `escolar_id` derivado de la tutoría.
  - Revisión de coordinación.

## Datos sintéticos

Con `SEMBRAR_SINTETICO=true`, el arranque del backend crea una tutoría ya finalizada
entre `escolar.prueba@example.test` y `tutor.prueba@<dominio>`. La reserva usa el slot
`slot-i13-demo`, tiene una sesión asociada y una asistencia declarada por el escolar.

La siembra es idempotente: si la reserva o la sesión ya existen, no se duplican. Esto
permite entrar como escolar o apoderado y elegir esa tutoría en `/incidentes` para
probar el registro de una falta con evidencia.

## Límites y pendientes

- La revisión de coordinación registra una decisión y un comentario; no implementa
  sanciones ni notificaciones automáticas al tutor. Esas acciones corresponden a otros
  flujos del plan.
- La "tutoría atendida" se interpreta como una reserva confirmada cuya hora de fin ya
  pasó. La asistencia declarada se muestra como contexto, pero no bloquea el reporte:
  un escolar puede necesitar reportar una conducta aunque no haya declarado asistencia.
- La revisión de otra persona y el PR quedan pendientes para cerrar el criterio de I13.
