# Evidencia de I04 · recorrido en navegador

**Fecha:** 2026-10-20 · **Rama:** `feature/reserva-i04` · **PR:** #39
**Entorno:** Chrome headless 1100×950 · API Go en `:8080` · Vite en `:5173` · PostgreSQL 16 (`:5433`)
**Datos:** sintéticos (`SEED_SYNTHETIC=true`), escolar `escolar.prueba@example.test`.

## Recorrido

| Paso | Evidencia |
|---|---|
| Ingreso del escolar | ![Ingreso](evidencia/01-ingreso.png) |
| Inicio con el enlace del flujo D | ![Inicio](evidencia/02-inicio-escolar.png) |
| Formulario con los horarios rechazados y su motivo | ![Formulario](evidencia/03-formulario.png) |
| Formulario completo con adjunto | ![Formulario completo](evidencia/04-formulario-completo.png) |
| Reserva confirmada (hora de Lima) | ![Confirmada](evidencia/05-confirmada.png) |
| El mismo horario ya no se puede elegir | ![Ocupado](evidencia/06-horario-ocupado.png) |

## Datos observados en la ejecución

Horarios que devuelve `GET /api/reserva/slots`:

| Slot | Curso | Inicio (Lima) | Disponible | Motivo |
|---|---|---|---|---|
| slot-i04-01 | Álgebra | 2026-10-22 15:00 | sí | — |
| slot-i04-02 | Aritmética | 2026-10-23 10:00 | sí | — |
| slot-i04-03 | Geometría | 2026-10-23 16:00 | sí | — |
| slot-i04-04 | Trigonometría | 2026-10-25 11:00 | sí | — |
| slot-i04-cercano | Álgebra | 2026-10-20 09:00 | **no** | la sesión debe empezar al menos 48 horas después de este momento |
| slot-i04-lejano | Geometría | 2026-11-18 15:00 | **no** | el horario está fuera del horizonte de 28 días |

Confirmación mostrada al escolar:

```text
Reserva confirmada

Tu sesión de Álgebra con Tutor Prueba (sintético) quedó agendada para el 2026-10-22 15:00 (hora de Lima).

Tema: Repaso de factorización
Grado: 5to de secundaria
Duración: 60 minutos
Número de reserva: ce530246-85b5-4cd1-89e0-4e75a9a8ad83
Adjuntos: practica.pdf
```

Al volver al formulario, el horario reservado aparece con el motivo
"el horario elegido ya fue reservado", junto al de 48 horas y al del horizonte.

## Criterios de aceptación

| Criterio (RF-10, RF-11, RN-03, RN-07) | Resultado |
|---|---|
| Formulario con tema, objetivo y adjuntos opcionales | Cumple: capturas 03 y 04 |
| La reserva se crea sobre un slot sintético y queda confirmada | Cumple: captura 05 y fila en `reservas` |
| Regla de 48 horas aplicada por el servidor | Cumple: `slot-i04-cercano` deshabilitado y 422 en la API |
| Tema 10–200, objetivo 20–1000, grado obligatorio | Cumple: validación en el navegador y en el modelo |
| Sin pruebas de concurrencia (I09) | Declarado fuera de alcance |

## Cómo reproducirlo

```bash
docker run -d --name uniapp-pg -e POSTGRES_USER=uniapp -e POSTGRES_PASSWORD=uniapp \
  -e POSTGRES_DB=uniapp -p 5433:5432 postgres:16-alpine
cd backend && DATABASE_URL="postgres://uniapp:uniapp@localhost:5433/uniapp?sslmode=disable" \
  SEED_SYNTHETIC=true SEED_PASSWORD="Clave-Segura-2026!" go run ./cmd/api
cd frontend && npm run dev
# http://localhost:5173/ingreso · escolar.prueba@example.test / Clave-Segura-2026!
```
