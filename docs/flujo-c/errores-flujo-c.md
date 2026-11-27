# Reproducción de errores del flujo C (F-31)

**Fecha:** 2026-11-27 · **Responsable:** Jhon · **Formato:** el mismo de los errores F-xx de #66 y del flujo E.

> Las pruebas están escritas contra el comportamiento **correcto** esperado. Mientras el error
> siga en el código, **deben fallar**. Si una pasa, el error ya se corrigió o la prueba dejó de
> comprobar lo que debe.

## Ejecución

```bash
cd backend
go test -run 'F31' -v ./internal/adapters/repositories/postgres/oferta/
```

Necesita `TEST_DATABASE_URL`; la prueba crea su propia base y aplica las migraciones.

Estado: **corregido**. El test queda en verde y sirve como prueba de regresión.

| Id | Tipo (lámina *Software Errors*) | Gravedad |
|---|---|---|
| F-31 | Integridad de datos / operaciones destructivas | Alta |

---

## F-31 · Eliminar un horario con reserva deja la reserva huérfana

- **Test:** `backend/internal/adapters/repositories/postgres/oferta/errores_flujo_c_test.go`,
  `TestF31NoDebeEliminarSlotConReserva`.
- **Archivo:** `backend/internal/adapters/repositories/postgres/oferta/repositorio_postgres.go`
  (`EliminarSlot`).
- **Reproducción GUI:**
  1. Un tutor publica el lunes de 09:00 a 10:00 y un escolar reserva ese horario.
  2. El tutor entra a "Mi disponibilidad" y elimina ese bloque.
  3. La reserva sigue confirmada, pero su `slot_id` ya no existe; el horario se puede volver a publicar y reservar otra vez.
- **Código anterior:**
  ```go
  func (r RepositorioPostgres) EliminarSlot(ctx context.Context, tutorID string, slotID int64) error {
      resultado, err := r.DB.ExecContext(ctx,
          `DELETE FROM slots WHERE id = $1 AND tutor_id = $2`, slotID, tutorID)
      ...
  }
  ```
- **Esperado:** rechazar la eliminación mientras el slot tenga una reserva confirmada
  (`409 slot_con_reserva`).
- **Corrección:** `EliminarSlot` cuenta las reservas confirmadas de ese slot y devuelve
  `ErrSlotConReserva` antes de borrar. El controlador lo traduce a `409 slot_con_reserva`.
  El test F-31 queda como regresión.

## Pendiente de cubrir

- **F-32 (propuesto):** un tutor con expediente **pendiente** puede publicar disponibilidad;
  el controlador solo verifica el rol. Debería responder `409 expediente_no_aprobado`.
