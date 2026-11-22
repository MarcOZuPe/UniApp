# I18 · Reporte y registro externo (flujo E)

**Estado:** implementado en la rama `feature/i18-reporte`; pendiente de revisión.
**Requisitos:** RF-17, RF-18 · **Depende de:** I17 · **Revisor:** Jhon (plan v0.4) · **Issue:** #23

## Objetivo

Reporte PDF y CSV con trazabilidad y escolares seudonimizados. Reconocimiento oficial solo con
evidencia externa, sin conexión ficticia.

## Qué hace

Coordinación elige un período (y, si quiere, un tutor) y descarga el reporte de minutos en CSV o
PDF desde `/coordinacion/reportes`. Cada descarga deja una **constancia** inmutable con la huella
SHA-256 del archivo exacto entregado.

```
coordinador ──► GET /api/reportes/minutos?desde&hasta&formato[&tutor_id]
                  │
                  ├─ sesiones del período (I05) con asistencia y resultado (I11)
                  ├─ minutos de I17, sincronizados con las cancelaciones (anulado = 0)
                  ├─ seudonimiza a los escolares (HMAC-SHA256)
                  ├─ arma filas y totales: solo VERIFICADO suma
                  ├─ genera CSV o PDF ──► huella SHA-256
                  └─ guarda la constancia en sesion_reportes ──► recién ahí entrega el archivo
```

## Reglas de negocio (en `backend/internal/domain/sesion/reporte.go`)

| # | Regla | Error |
|---|---|---|
| R1 | Solo coordinación (`coordinacion:acceder`) descarga reportes y ve las constancias. | `403 sin_permiso` |
| R2 | Período obligatorio en días de Lima, `hasta` incluido, `desde ≤ hasta`, máximo **31 días**. | `400 fecha_invalida`, `422 periodo_invalido` |
| R3 | **El escolar nunca aparece con su ID ni su nombre**: `ESC-` + 10 caracteres de `HMAC-SHA256(clave, escolarID)`. Es estable (el mismo escolar tiene el mismo seudónimo) e irreversible sin la clave. | — |
| R4 | La clave viene de `I18_SEUDONIMO_CLAVE` (mínimo 16 caracteres). Es **obligatoria en `APP_ENV=prod`**; en dev se usa una clave fija y el log lo avisa. Nunca va en el reporte ni en Git. | el servidor no arranca |
| R5 | El total por tutor suma **solo VERIFICADO**. PROPUESTO, EN_DISPUTA, ANULADO y sin registro aparecen, pero suman 0 (regla R8 de I17). | — |
| R6 | **Trazabilidad** por fila: estado de los minutos, propuestos, los que cuentan, número de ajustes/resoluciones y la última acción de la auditoría de I17 (acción, actor y fecha). | — |
| R7 | **Constancia** de cada descarga: id, autor, fecha, período, tutor, formato, filas y SHA-256. Tabla de solo inserción (*trigger* como la auditoría de I17). **Si la constancia no se guarda, el archivo no se entrega.** | `500 error_interno` |
| R8 | **Sin conexión ficticia (RF-18)**: el reporte lleva siempre el aviso *"Sin validez oficial: no existe convenio ni registro externo verificado con la UNI (E01 pendiente)"*. No hay botón de "enviar a la UNI" ni API simulada. | — |
| R9 | Hora del servidor (RN-03) para la fecha de generación; el cliente solo envía filtros. | — |
| R10 | CSV seguro: UTF-8 con BOM (Excel) y toda celda que empiece con `= + - @` lleva `'` delante para evitar inyección de fórmulas. | — |

## Lo de otros flujos (sin tocar su código)

| Dato | Cómo se obtiene |
|---|---|
| Nombre del tutor y del coordinador | Puerto `DirectorioNombres` del dominio de sesión. El adaptador sobre `acceso` vive en `cmd/api/reporte.go` (composition root). Si no hay nombre, se muestra el ID. |
| Cancelaciones | El mismo puerto `EstadoReservas` de I17 (fixture hoy; I16 de Jhon después). |
| Nombre del curso | No se incluye en esta versión: es un dato del flujo D. Se muestra el ID de la reserva, que sirve para rastrear. |

## El PDF, sin dependencias

El ADR 0001 fija "solo librería estándar + `lib/pq`". El PDF se genera con un escritor mínimo propio
(`adapters/presentation/sesion/pdf.go`): PDF 1.4, A4 horizontal, fuentes estándar Helvetica (no se
incrustan) con codificación WinAnsi para tildes y ñ, tabla paginada con encabezado repetido y
"Página X de Y". Las pruebas comprueban la estructura (cabecera, tabla `xref` y `%%EOF`) y la
paginación con 120 filas.

## Alcance de la entrega

| Pieza | Archivo |
|---|---|
| Dominio | `backend/internal/domain/sesion/reporte.go` (+ `reporte_i18_test.go`) |
| Caso de uso | `backend/internal/application/sesion/reporte.go` (+ test) |
| Repositorios | `adapters/repositories/{memory,postgres}/sesion/reportes.go` (+ test de Postgres) |
| Migración | `backend/migrations/014_sesion_reportes.{up,down}.sql` (tabla, *trigger* e índice por inicio) |
| CSV y PDF | `adapters/presentation/sesion/reporte.go`, `pdf.go` (+ tests) |
| HTTP | `adapters/http/sesion/reporte.go` (+ test): `GET /api/reportes/minutos`, `GET /api/reportes/registro` |
| Cableado | `backend/cmd/api/reporte.go` y una línea en `main.go` |
| Frontend | `pages/sesion/ReporteMinutos.jsx` (ruta `/coordinacion/reportes`), `adapters/http/reporteApi.js`, `domain/sesion/reporte.js`, enlace en el menú (+ tests) |
| Configuración | `I18_SEUDONIMO_CLAVE` en `.env.example` y `docker-compose.yml` |
| Contrato | `docs/openapi.yaml`: rutas `/api/reportes/*` y esquema `ConstanciaReporteI18` |

## Cómo probarlo en local

1. Levantar la API con `SEED_SYNTHETIC=true`.
2. Entrar como `coord.prueba@example.test` y abrir **Reporte de minutos** en el menú.
3. Elegir el período y descargar CSV o PDF. La constancia aparece en la tabla de abajo, con la
   misma huella que trae la cabecera `X-Reporte-Sha256` de la descarga.

Para que el reporte tenga minutos verificados: el tutor propone minutos en una sesión terminada
con asistencia de ambos (I17) y coordinación los aprueba en **Revisión de minutos**.

## Pendiente fuera de I18

- **Nombre del curso** en el reporte, cuando el flujo D exponga un puerto o una API para leerlo.
- **Reporte propio del tutor** (solo sus sesiones): se dejó para S4.
- **Carga y latencia** de la generación (I23).
