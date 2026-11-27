# Flujo D · I16 Cancelar una reserva con historial

**Responsable:** Jhon · **Estado:** cancelación implementada; la reprogramación queda pendiente.

## Qué permite

El escolar dueño, el tutor asignado o coordinación pueden cancelar una reserva **antes de que
empiece la sesión**. El cambio de estado no borra nada: la reserva queda en `cancelada` y se
guarda un evento con actor, fecha y motivo.

## Reglas

| Regla | Comportamiento |
|---|---|
| Estado de origen | Solo se cancela una reserva `confirmada` |
| Momento | Solo antes del inicio de la sesión |
| Motivo | Opcional, hasta 200 caracteres |
| Historial | Un evento por cambio, con actor y fecha |
| Permisos | Escolar dueño, tutor asignado o coordinación |

Errores: `409 reserva_no_cancelable` si ya estaba cancelada, `409 cancelacion_tardia` si la
sesión ya empezó, `403` si quien cancela no participa, `404` si la reserva no existe.

## API

| Método | Ruta | Respuesta |
|---|---|---|
| POST | /api/reserva/{id}/cancelar | 200 con la reserva en estado `cancelada` |
| GET | /api/reserva/{id}/historial | 200 con la lista de eventos |
| POST | /api/reserva/{id}/reprogramar | 200 con la reserva en el horario nuevo |

Ejemplo del cuerpo opcional:

```json
{ "motivo": "Enfermedad del escolar" }
```

## Modelo de datos

```text
reservas: estado (confirmada | cancelada), cancelada_en, motivo_cancelacion
reserva_historial: reserva_id, estado_anterior, estado_nuevo, motivo, actor_id, creado_en
```

La reserva cancelada deja de contar para el cupo futuro y para la agenda; el horario vuelve a
estar disponible. La migración es `021_reserva_cancelacion`.

## Pruebas

- `TestI16CancelarReservaConHistorial`: estado, fecha, motivo y evento con actor.
- `TestI16NoSePuedeCancelarDosVeces`: `ErrReservaNoCancelable`.
- `TestI16NoSeCancelaUnaSesionEmpezada`: `ErrCancelacionTardia`.

## Reprogramar

`POST /api/reserva/{id}/reprogramar` con `{ "slot_id": "...", "motivo": "..." }` mueve la
reserva a otro horario **conservando su identificador y su historial**. Verifica que el
horario exista, esté libre, no sea el mismo y cumpla la ventana de 48 horas y el horizonte.
El evento queda en `reserva_historial` con el slot anterior y el nuevo.

Errores: `409 mismo_horario`, `409 slot_ocupado`, `409 reserva_no_cancelable`,
`422 anticipacion_insuficiente` u `horizonte_excedido`.

## Pendiente

- Botón de cancelar y reprogramar en la agenda del escolar.
