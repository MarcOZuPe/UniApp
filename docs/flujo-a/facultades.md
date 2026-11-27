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

Catálogo oficial cargado por las migraciones `017`, `018` y `020`:

| ID | Siglas | Facultad |
|---|---|---|
| 1 | FC | Facultad de Ciencias |
| 2 | FIIS | Facultad de Ingeniería Industrial y de Sistemas |
| 3 | FIM | Facultad de Ingeniería Mecánica |
| 4 | FAUA | Facultad de Arquitectura, Urbanismo y Artes |
| 5 | FIA | Facultad de Ingeniería Ambiental |
| 6 | FIC | Facultad de Ingeniería Civil |
| 7 | FIEECS | Facultad de Ingeniería Económica, Estadística y Ciencias Sociales |
| 8 | FIEE | Facultad de Ingeniería Eléctrica y Electrónica |
| 9 | FIGMM | Facultad de Ingeniería Geológica, Minera y Metalúrgica |
| 10 | FIP | Facultad de Ingeniería de Petróleo, Gas Natural y Petroquímica |
| 11 | FIQT | Facultad de Ingeniería Química y Textil |

La migración `020_facultades_oficiales` conserva los IDs 1–4 usados por las
cuentas sintéticas y agrega las siete facultades restantes. También guarda
`siglas` y `especialidades` en la tabla `facultades`.

Cuentas sintéticas por facultad:

| ID | Facultad | General sintético | Coordinador sintético | Tutor sintético |
|---|---|---|---|---|
| 1 | Ciencias | `general.ciencias@example.test` | `coord.ciencias@example.test` | `tutor.ciencias@uni.edu.pe` |
| 2 | Industrial y Sistemas | `general.sistemas@example.test` | `coord.sistemas@example.test` | `tutor.sistemas@uni.edu.pe` |
| 3 | Mecánica | `general.mecanica@example.test` | `coord.mecanica@example.test` | `tutor.mecanica@uni.edu.pe` |
| 4 | Arquitectura | `general.arquitectura@example.test` | `coord.arquitectura@example.test` | `tutor.arquitectura@uni.edu.pe` |

Las facultades 5–11 no tienen cuentas sintéticas propias; se usan en registro y
operación real.

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
