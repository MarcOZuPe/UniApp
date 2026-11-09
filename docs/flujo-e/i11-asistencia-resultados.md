# I11 · Asistencia y resultados completos (flujo E)

**Estado:** implementado en la rama `feature/i11-asistencia-resultados`; pendiente de revisión.
**Requisito:** RF-15 · **Depende de:** I05 (integrada), I09 (integrada) · **Revisor:** Miguel · **Issue:** #15

## Objetivo

Declaración de ambas partes y ausencia reportada. Un enlace abierto no acredita asistencia.
Registro del resultado de la sesión.

## Alcance de la entrega

| Pieza | Archivo | Qué resuelve |
|---|---|---|
| Ausencia | `backend/internal/domain/sesion/ausencia.go` | ausencia **propia**, excluyente con la asistencia y sin duplicados |
| Resultado | `backend/internal/domain/sesion/resultado.go` | un único resultado del tutor asignado (1–2000 caracteres) |
| Plazos | `backend/internal/domain/sesion/plazos.go` | `Intervalo` (value object) y ventanas de declaración y de resultado |
| Persistencia | migraciones `010_sesion_ausencias`, `011_sesion_resultados` | tablas con clave única por participante / por sesión |
| Repositorios | `adapters/repositories/{postgres,memory}/sesion` | transacción con `SELECT … FOR UPDATE` sobre la sesión |
| HTTP | `adapters/http/sesion/{ausencia,resultado,controlador}.go` | endpoints, mapeo de errores y plazos publicados |
| Vista | `frontend/src/presentation/pages/sesion/DetalleSesion.jsx` | botones de asistencia/ausencia, formulario de resultado y avisos de plazo |
| Regla en UI | `frontend/src/domain/sesion/plazos.js` | estado del plazo (pendiente/abierto/cerrado) como función pura |

## Reglas de negocio

1. **La identidad sale de la autenticación.** Asistencia y ausencia no admiten cuerpo; el resultado
   rechaza campos ajenos (`usuario_id`, `registrado_en`…). Nadie declara por otra persona.
2. **Asistencia y ausencia son excluyentes** por participante (`409 declaracion_contradictoria`),
   y cada una se registra una sola vez (`409 asistencia_duplicada` / `ausencia_duplicada`).
3. **Abrir el detalle no registra nada.** `GET /api/sesion/{id}` es de solo lectura: un enlace abierto
   no acredita asistencia.
4. **El resultado** lo registra solo el tutor asignado, una vez (`409 resultado_duplicado`).
   No acredita asistencia ni minutos; eso corresponde a I17.
5. **Plazos** (`422 fuera_de_plazo`):

   | Acción | Desde | Hasta |
   |---|---|---|
   | Declarar asistencia o reportar ausencia | 15 min antes del inicio | 24 h después del fin |
   | Registrar resultado | inicio de la sesión | 7 días después del fin |

   Los valores son constantes del dominio (`AnticipacionDeclaracion`, `PlazoDeclaracion`,
   `PlazoResultado`). El backend los publica en `plazos` dentro de `GET /api/sesion/{id}`, y la interfaz
   solo los muestra: la regla tiene una única fuente.

## Decisiones de diseño

- **Comando frente a reconstitución.** `DeclararAsistencia`, `ReportarAusencia` y `RegistrarResultado`
  son comandos: validan identidad, plazo e invariantes. Los repositorios rehidratan con
  `RestaurarAsistencia`, `RestaurarAusencia` y `RestaurarResultado`, que aplican las invariantes pero
  **no** la política temporal. Así los datos históricos, guardados antes de existir los plazos, siguen
  cargando y no provocan un 500.
- **La autorización va antes que el plazo.** Un tercero recibe `403` aunque esté fuera de plazo, para
  no revelar detalles de una sesión ajena.
- **Reloj inyectable (DIP).** `Rutas(..., ConReloj(f))` sigue el patrón *functional options*.
  Producción usa `time.Now`; las pruebas fijan el momento. La firma pública de `Rutas` no cambia.
- **Concurrencia.** Asistencia, ausencia y resultado bloquean la fila de la sesión (`FOR UPDATE`) dentro
  de la transacción. Las restricciones únicas de la base son la última defensa (`23505` → `409`).
- **Arquitectura.** El dominio no conoce HTTP ni SQL. Los alias de `application`/`adapters` se generan
  con `tools/aliasgen`, y `tests/architecture` verifica las dependencias.

## Endpoints

```http
GET  /api/sesion/{id}                      # incluye ausencias, resultado y plazos
GET  /api/sesion/por-reserva/{reservaID}   # abrir el detalle desde la agenda (I05)
POST /api/sesion/{id}/asistencias          # sin cuerpo
POST /api/sesion/{id}/ausencias            # sin cuerpo
POST /api/sesion/{id}/resultado            # {"resumen": "..."}
```

El contrato completo está en `docs/openapi.yaml`, etiqueta `E-sesion`.

## Pruebas

```bash
cd backend && go vet ./... && go test ./...
# con PostgreSQL real (tablas temporales, no toca tablas públicas):
I05_TEST_DATABASE_URL=postgres://... go test -count=1 ./internal/adapters/repositories/postgres/sesion
cd frontend && npm test && npm run build
```

- **Dominio:** límites exactos de cada ventana, la autorización antes que el plazo, y que la
  reconstitución ignora plazos pero mantiene las invariantes.
- **HTTP:** 401/403/404/409/422, origen no permitido, cuerpo rechazado, plazos publicados, y
  `fuera_de_plazo` antes y después de la sesión.
- **Memoria:** carreras entre asistencia y ausencia, y entre dos resultados (solo uno gana).
- **PostgreSQL:** persistencia, contradicciones y rehidratación de datos históricos fuera de plazo.
  Se ejecutan en CI con `I05_TEST_DATABASE_URL`.
- **Frontend:** estado del plazo (límites), botones bloqueados con el plazo cerrado y formulario de
  resultado oculto antes del inicio.

## Fuera de alcance (declarado)

- Minutos verificados, disputas y ajustes auditados: **I17**.
- Reporte PDF/CSV: **I18**.
- Reportar la ausencia de otra persona: I11 no lo especifica, y no se amplían permisos.
- **Sesión de demostración:** `sesion-i05-demo` se siembra una sola vez, al día siguiente del primer
  arranque. Pasado su plazo ya no admite declaraciones; para volver a demostrarlas hay que sembrar una
  sesión nueva.
