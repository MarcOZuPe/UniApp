# Separación por facultad

## Alcance

- Tutores, coordinadores y coordinadores generales pertenecen a una facultad.
- Escolares y apoderados no pertenecen a una facultad y pueden reservar tutorías con
  cualquier tutor.
- La facultad se fija al crear la cuenta y la base rechaza cualquier intento de
  modificarla.
- La coordinación solo puede ver y gestionar tutores, incidentes, métricas y
  auditoría de su propia facultad.
- La coordinación general solo puede crear, listar y desactivar coordinadores de su
  propia facultad.

## Creación de cuentas

- El registro público de tutor exige seleccionar una facultad.
- Las cuentas internas de coordinación heredan la facultad del coordinador general que
  las crea.
- Las cuentas de escolares y apoderados se crean sin facultad.
- Los coordinadores generales iniciales se aprovisionan con la facultad 1 o mediante
  administración de base de datos.

## Datos iniciales

| ID | Facultad | General sintético | Coordinador sintético | Tutor sintético |
|---|---|---|---|---|
| 1 | Facultad de Ciencias | `general.ciencias@example.test` | `coord.ciencias@example.test` | `tutor.ciencias@uni.edu.pe` |
| 2 | Facultad de Ingeniería Industrial y Sistemas | `general.sistemas@example.test` | `coord.sistemas@example.test` | `tutor.sistemas@uni.edu.pe` |
| 3 | Facultad de Ingeniería Mecánica | `general.mecanica@example.test` | `coord.mecanica@example.test` | `tutor.mecanica@uni.edu.pe` |
| 4 | Facultad de Arquitectura, Urbanismo y Artes | `general.arquitectura@example.test` | `coord.arquitectura@example.test` | `tutor.arquitectura@uni.edu.pe` |

Todas usan la clave sintética configurada en `SEED_PASSWORD`. Las cuentas genéricas
anteriores (`coord.prueba`, `coord.general.prueba`, `tutor.prueba`, `tutor2.prueba`)
se conservan en la facultad 1. Cada tutor sintético recibe un expediente pendiente
para que coordinación pueda probar el flujo de revisión de su facultad.

## Reglas de acceso

- Expedientes y documentos de tutores: solo coordinación de la misma facultad.
- Incidentes: solo coordinación de la facultad del tutor denunciado.
- Métricas del panel: solo reservas, sesiones, asistencia, horas y pendientes de la
  facultad del coordinador.
- Auditoría: solo eventos cuyo actor pertenece a la facultad.
- Catálogo: sigue siendo global, porque no pertenece a una facultad.
- Agenda y reservas: el escolar puede reservar con tutores de cualquier facultad.

## Base de datos

Migraciones `backend/migrations/017_facultades.up.sql` y
`backend/migrations/018_facultades_sinteticas.up.sql`:

- tabla `facultades`;
- columna `usuarios.facultad_id`;
- trigger que asigna la facultad por defecto a cuentas internas y prohíbe cambiarla;
- constraint por rol;
- columna `incidentes.facultad_id` con backfill desde el tutor.
