# Registro de incidentes · Sprint I01 (Flujo A · Acceso) + E02

Bitácora de lo que **salió mal durante el sprint** y cómo se resolvió. Complementa al
código y a las pruebas: estas demuestran que el sistema funciona, el registro guarda
**qué falló, por qué y qué se hizo**. Cubre los cuatro tipos de "incidente" del equipo:

| Tipo | Qué recoge |
|---|---|
| `defecto` | Bug en el código o en las pruebas (comportamiento incorrecto). |
| `impedimento` | Bloqueo del equipo: entorno, infraestructura, dependencias entre flujos. |
| `incidencia` | Evento operativo: fallo de despliegue, caída del servicio, conflicto de recursos. |
| `riesgo` | Riesgo materializado o detectado que amenaza un entregable del sprint. |

**Severidad:** `alta` (bloquea el sprint) · `media` (bloquea una tarea) · `baja` (molestia).

## Incidentes

| id | tipo | fecha | título | descripción | impacto | causa raíz | acción / resolución | estado |
|---|---|---|---|---|---|---|---|---|
| INC-01 | defecto | 2026-10-18 | Falta la base `uniapp_test` | `TestRepositorioPostgresFlujoCompleto` fallaba con `pq: database "uniapp_test" does not exist`. | **medio**: prueba de integración en rojo | La prueba no crea su base de datos ni hay script que la prepare | `CREATE DATABASE uniapp_test` en el contenedor. *Deuda:* automatizarlo (script/`Makefile`) | resuelto |

### Notas
- Incidentes resueltos: 1 · abiertos: 0 · severidad alta: 0.

## Plantilla para nuevas entradas

Copiar una fila y completar. El `id` es global y creciente: no se reutilizan números ya usados.

| id | tipo | fecha | título | descripción | impacto | causa raíz | acción / resolución | estado |
|---|---|---|---|---|---|---|---|---|
| INC-XX | defecto \| impedimento \| incidencia \| riesgo | AAAA-MM-DD | resumen en una línea | qué se observó, con el mensaje de error si lo hay | severidad + a qué bloqueó | por qué ocurrió (no el síntoma) | qué se hizo y **evidencia** (comando, test, PR) | abierto \| en curso \| resuelto |

**Convenciones**
- Escribir la **causa raíz**, no el síntoma (regla del *análisis de causa raíz*).
- Adjuntar **evidencia** reproducible (comando, salida del test, enlace al PR).
- Marcar **deuda técnica** cuando la solución sea un parche (p. ej. INC-01).
- Un incidente se cierra solo con su causa raíz resuelta y prueba de que no reaparece.
