# Evidencia de I09 · reserva transaccional y adjuntos

**Fecha:** 2026-10-22 · **Rama:** `feature/reserva-i09` · **Entorno:** API Go `:8080`, Vite `:5173`, PostgreSQL 16 limpio, datos sintéticos.

## Recorrido en navegador

| Paso | Captura |
|---|---|
| Formulario con el archivo elegido antes de confirmar | ![Formulario](evidencia-i09/01-formulario-con-adjunto.png) |
| Reserva confirmada con el adjunto validado y enlace de descarga | ![Confirmada](evidencia-i09/02-confirmada-con-adjunto.png) |
| El horario ya no se puede elegir al volver al formulario | ![Ocupado](evidencia-i09/03-horario-ocupado.png) |

Datos observados en esa ejecución:

```text
Reserva confirmada · Álgebra · 2026-10-25 15:00 (hora de Lima)
Número de reserva: 19e1748f-9e07-4114-897a-9a83f2592ba9
practica9.pdf · 69 B — Listo para descargar · Descargar
GET del enlace privado -> 200, Content-Type application/pdf, prefijo "%PDF-1.4"
Motivos al volver al formulario:
  Álgebra · 2026-10-25 15:00: el horario elegido ya fue reservado
  Álgebra · 2026-10-23 09:00: la sesión debe empezar al menos 48 horas después de este momento
  Geometría · 2026-11-21 15:00: el horario está fuera del horizonte de 28 días
```

## Recorrido HTTP (curl con cookie de sesión)

| Prueba | Resultado |
|---|---|
| `POST /api/reserva` con `Idempotency-Key` | **201** |
| El mismo POST con la misma clave y datos | **200** con el **mismo id** |
| Misma clave con otro tema | **409** `idempotencia_conflicto` |
| Tercera reserva futura | **409** `cuota_escolar` |
| `POST /api/reserva/{id}/adjuntos` con PDF | **201** `estado=validado`, `descargable=true` |
| Subir `pagina.html` | **422** `adjunto_invalido` |
| `GET /api/reserva/{id}/adjuntos/{adjuntoId}` con sesión | **200**, `Content-Disposition: attachment; filename=practica.pdf`, `Cache-Control: no-store` |
| El mismo GET sin sesión | **401** |
| Servidor con `RESERVA_ATTACHMENTS_SCAN=ninguno`: subir PDF | **201** `estado=en_cuarentena` con motivo |
| Descargar ese adjunto | **409** `adjunto_en_cuarentena` |
| Consultar la reserva con el adjunto en cuarentena | **200** (la reserva sigue confirmada) |

Estado en la base:

```text
reserva_adjuntos: practica.pdf | validado | 69 | huella cfa3181c1ee3 | clave 25d9ecc0fd3ab6dad4b9be360d1f777d
disco: data/reserva-adjuntos/25d9ecc0fd3ab6dad4b9be360d1f777d (69 bytes, permisos 0600)
schema_migrations: 001 … 007, incluida 007_reserva_transaccional
```

## Pruebas automáticas

| Tipo | Comando | Resultado |
|---|---|---|
| Unidad de modelo y controlador | `go test ./...` | 13 paquetes en verde |
| Integración con PostgreSQL | `TEST_DATABASE_URL=... go test ./internal/models/reserva` | carreras, idempotencia, solapamiento y cuota concurrentes en verde |
| Frontend | `npm test` y `npm run build` | 33 pruebas y compilación correctas |

## Cómo reproducirlo

```bash
docker run -d --name uniapp-pg -e POSTGRES_USER=uniapp -e POSTGRES_PASSWORD=uniapp \
  -e POSTGRES_DB=uniapp -p 5432:5432 postgres:16-alpine
cd backend && DATABASE_URL="postgres://uniapp:uniapp@localhost:5432/uniapp?sslmode=disable" \
  SEED_SYNTHETIC=true SEED_PASSWORD="Clave-Segura-2026!" go run ./cmd/server
cd frontend && npm run dev
# http://localhost:5173/ingreso · escolar.prueba@example.test / Clave-Segura-2026!
```
