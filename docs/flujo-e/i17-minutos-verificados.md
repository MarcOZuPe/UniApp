# I17 · Minutos verificados y ajustes auditados (flujo E)

**Estado:** implementado en la rama `feature/i17-minutos-verificados`; pendiente de revisión.
**Requisito:** RF-16 · **Depende de:** I11 · **Revisor:** Jhon (plan v0.4) · **Issue:** #22

## Objetivo

Interfaz de revisión para coordinación, que no autoaprueba. 45 minutos suman 45, una disputa congela
y una cancelación suma cero. Ajustes con auditoría.

## Ciclo de vida

```
            tutor propone               coordinación aprueba o ajusta
 (sin registro) ─────────► PROPUESTO ──────────────────────────► VERIFICADO ──► suma al total
                              │  ▲                                  │
           escolar disputa    │  │ coordinación resuelve            │ escolar disputa, o el tutor
                              ▼  │ (con motivo, final)              ▼ si hubo un ajuste a la baja
                           EN_DISPUTA (congelado: suma 0) ◄──────────┘

 Reserva cancelada ──► ANULADO (suma 0, en cualquier estado)
```

## Reglas de negocio (todas en `backend/internal/domain/sesion/minutos.go`)

| # | Regla | Error |
|---|---|---|
| R1 | El tutor asignado propone los minutos **realmente dictados** (entero 1–60), después del **fin** de la sesión y dentro de `PlazoResultado` (7 días). | `422 minutos_invalidos`, `sesion_no_finalizada`, `fuera_de_plazo` |
| R2 | Solo se propone si **tutor y escolar declararon asistencia** (I11). Un enlace abierto no acredita. | `422 asistencia_incompleta` |
| R3 | **No autoaprueba:** solo un rol con `coordinacion:acceder` verifica, y nunca un participante de la sesión. | `403 sin_permiso`, `403 autoaprobacion` |
| R4 | **Una disputa congela:** el escolar disputa lo propuesto o lo verificado; el tutor, solo un verificado menor que lo que propuso. Exige motivo. Mientras dure suma 0. | `409 disputa_no_permitida`, `422 motivo_invalido` |
| R5 | Coordinación **resuelve** la disputa con los minutos finales (0–60) y un motivo. La resolución es **final** (no se puede volver a disputar). | `409 transicion_invalida` |
| R6 | **Una cancelación suma cero:** si la reserva está cancelada, el registro pasa a ANULADO (persistido y auditado) y se rechaza cualquier acción. | `409 sesion_cancelada` |
| R7 | **Ajustes con auditoría:** cada acción guarda quién, cuándo, valor anterior, valor nuevo y motivo. La tabla de auditoría es de solo inserción; un *trigger* impide `UPDATE` y `DELETE`. | — |
| R8 | El **total** de un tutor es la suma de sus registros verificados. Los propuestos, en disputa y anulados suman 0. Es el insumo de I18. | — |

Ejemplo del objetivo: se proponen 45 → no suma (0) → coordinación aprueba → **suma 45**, no 60.

## Cancelación sin depender del flujo D (puerto)

El flujo E **no lee ni escribe** las tablas de reserva. Pregunta por un puerto propio:

```go
// domain/sesion/minutos.go
type EstadoReservas interface {
    Cancelada(ctx context.Context, reservaID string) (bool, error)
}
```

- **Hoy:** lo responde `CancelacionesFijas` (`adapters/repositories/memory/sesion/minutos.go`). En
  producción no hay cancelaciones, salvo que se defina `I17_CANCELACIONES_FIXTURE` apuntando a
  `backend/fixtures/sesion-cancelaciones.json` (datos sintéticos).
- **Cuando exista I16 (Jhon):** se agrega un adaptador que consulte el estado real de la reserva y solo
  cambia `estadoReservasI17()` en `backend/cmd/api/minutos.go`. Las reglas de I17 no se tocan.

## Alcance de la entrega

| Pieza | Archivo |
|---|---|
| Dominio | `backend/internal/domain/sesion/minutos.go` (+ `minutos_i17_test.go`) |
| Casos de uso | `backend/internal/application/sesion/minutos.go` (+ test) |
| Repositorios | `adapters/repositories/{memory,postgres}/sesion/minutos.go` (+ tests) |
| Migración | `backend/migrations/013_sesion_minutos.{up,down}.sql` |
| HTTP y vista JSON | `adapters/http/sesion/minutos.go`, `adapters/presentation/sesion/minutos.go` (+ test) |
| Cableado | `backend/cmd/api/minutos.go` y una línea en `main.go` |
| Fixture | `backend/fixtures/sesion-cancelaciones.json` |
| Frontend | `pages/sesion/SeccionMinutos.jsx` (tutor/escolar, dentro del detalle), `pages/sesion/RevisionMinutos.jsx` (coordinación, ruta `/coordinacion/minutos`), `adapters/http/minutosApi.js`, `domain/sesion/minutos.js` (+ tests) |
| Contrato | `docs/openapi.yaml`, rutas `/api/sesion/{id}/minutos*` y `/api/minutos/*` |

## Decisiones de diseño

- **Agregado propio.** `RegistroMinutos` es independiente de `Sesion`: no se modificaron las tablas
  ni el repositorio de I05/I11. La sesión solo se lee para validar la propuesta.
- **Atomicidad.** `ActualizarMinutos` bloquea la fila (`FOR UPDATE`), aplica el comando del dominio y
  persiste el estado y **solo** las entradas de auditoría nuevas. Si el comando falla no se guarda
  nada (rollback, o copia descartada en memoria).
- **Permiso existente.** Se reutiliza `coordinacion:acceder`: no se modificó `domain/acceso` (flujo A).
- **Rutas sin colisión.** `/api/sesion/{id}/minutos` es más específico que `/api/sesion/`, así que las
  rutas de I05 e I11 no cambian.
- **Hora y autor del servidor (RN-03).** El cliente solo envía minutos y motivos; los campos ajenos
  se rechazan con `400`.

## Pendiente fuera de I17

- **Enlace en el panel de coordinación** (`Coordinacion.jsx`, flujo A/B de Víctor) hacia
  `/coordinacion/minutos`: queda a criterio de su dueño para no editar su vista.
- **Adaptador real de cancelaciones** cuando se integre I16.
