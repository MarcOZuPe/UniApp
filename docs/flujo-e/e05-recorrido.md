# E05 · Integración y recorrido de sesión a reporte (flujo E)

**Estado:** prueba de punta a punta implementada en `feature/e05-integracion`.
**Issue:** #16 · **Depende de:** I05, I10, I11, I17, I18 · **Revisión cruzada:** plan v0.4.

## Objetivo

Demostrar que el flujo E funciona completo, de la sesión al reporte, con los cuatro casos que pide
el plan: **cancelación, sesión parcial, disputa y reporte**.

## La prueba

`backend/internal/adapters/http/sesion/e05_recorrido_test.go` → `TestE05RecorridoSesionAReporte`.

- Usa **los mismos controladores HTTP que `cmd/api/main.go`** (`Rutas`, `RutasMinutos`, `RutasReporte`)
  y el login real de `acceso`, con repositorios en memoria y un reloj controlado.
- Cada paso es una petición HTTP con su cookie de sesión y su código de estado esperado.
- No necesita base de datos ni Docker. Se ejecuta en CI con el resto de pruebas.

```
cd backend
go test -run E05 -v ./internal/adapters/http/sesion/
```

## Recorrido

```
          tutor                 escolar                coordinación             flujo D
            │                      │                         │                     │
 I05  POST /asistencias     POST /asistencias  (completa y cancelada)              │
 I11        │               POST /ausencias    (parcial)     │                     │
 I11  POST /resultado              │                         │                     │
 I17  POST /minutos 45 y 60        │                         │                     │
      POST /minutos (parcial) ──► 422 asistencia_incompleta  │                     │
 I17        │                      │              POST /aprobacion (x2)               │
 I17        │               POST /minutos/disputa ──► congela (suma 0)             │
 I17        │                      │              POST /resolucion 35                │
      GET /minutos/total = 95      │                         │                     │
            │                      │                         │       cancela la reserva
 I18        │                      │              GET /reportes/minutos csv y pdf  │
 I18        │                      │              GET /reportes/registro           │
      GET /minutos/total = 35      │                         │                     │
```

## Qué se comprueba

| Caso | Resultado esperado | Cómo se verifica |
|---|---|---|
| Sesión completa con disputa | 45 propuestos → aprobados → disputa (suma 0) → resueltos en **35** | Respuestas de `/disputa` y `/resolucion`, y la fila del CSV `…,45,35,…` |
| Sesión parcial | El escolar reporta su ausencia; el tutor **no puede** proponer minutos | `422 asistencia_incompleta`; en el reporte: `ausencia_reportada` y `sin_registro` |
| Cancelación | Una reserva cancelada **suma 0**, aunque ya estuviera verificada | Fila `anulado` con `…,60,0,…`; total del tutor de 95 a **35** |
| Reporte CSV y PDF | Las 3 sesiones del día, aviso *"Sin validez oficial"*, escolar seudonimizado `ESC-…` | El CSV/PDF **no contiene** el ID del escolar |
| Trazabilidad | Una constancia por descarga con la huella SHA-256 del archivo entregado | `/reportes/registro` devuelve 2 constancias, cuyas huellas coinciden con `X-Reporte-Sha256` |

## Resultado

`TestE05RecorridoSesionAReporte` **pasa** (ver `docs/flujo-e/pruebas-flujo-e.md` para la ejecución completa).

## Límites (sin inventar)

- El estado de las reservas usa un **puerto falso** (`cancelacionesE05`): I16 de Jhon todavía no
  publica las cancelaciones al flujo E. El contrato (`EstadoReservas.Cancelada`) es el mismo.
- La persistencia real (Postgres) se prueba aparte, por repositorio, con `TEST_DATABASE_URL`.
- El reconocimiento oficial de horas (RF-18) **no existe**: no hay convenio con la UNI (E01).
