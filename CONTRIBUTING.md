# Cómo trabajamos (E02)

**Ramas:** una por issue, `feature/<flujo>-<tema>` (ej. `feature/acceso-cuentas`). `main` solo recibe PR.
**PR:** revisado por la persona designada en el issue (I01: Jhon) y con CI en verde. Cerrar un PR sin integrar no termina el issue.
**Terminado (flujo):** vista + controlador + modelo + pruebas + documentación, integrado en `main`.
**Capas:** ver `docs/adr/0001-stack-y-mvc.md`. `go test ./internal/arquitectura/` falla si se rompen.
**Contrato:** cada dueño documenta sus endpoints en `docs/openapi.yaml` y publica fixtures en `backend/fixtures/<flujo>.json` la primera semana de cada sprint.
**Migraciones:** `backend/migrations/NNN_flujo_tema.up.sql` (+ `.down.sql`). No se edita una migración ya integrada.
**Secretos:** nunca en Git. Van en `.env` (ignorado); `.env.example` solo lleva valores de ejemplo. `scripts/verificar-secretos.sh` corre en CI.
**Datos:** solo sintéticos o anonimizados; nada de documentación privada de tutores o menores en issues públicos.
**Comandos:** se documentarán (arranque, pruebas y verificación) cuando exista el código de aplicación.
