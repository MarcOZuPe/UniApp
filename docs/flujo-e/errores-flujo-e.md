# Reproducción de errores del flujo E (F-26 a F-30)

**Fecha:** 2026-11-26 · **Responsable:** Marco · **Formato:** el mismo de los errores F-xx/B-xx de #66.

> Las pruebas están escritas contra el comportamiento **correcto** esperado. Mientras el error
> siga en el código, **deben fallar**. Si una pasa, el error ya se corrigió o la prueba dejó de
> comprobar lo que debe.

## Ejecución

```bash
cd backend
go test -run 'F2[6-9]|F30' -v ./internal/domain/sesion/ ./internal/adapters/http/sesion/
```

Resultado esperado hoy: **5 en rojo**, cada uno con su mensaje `F-NN: ...`.

| Id | Tipo (lámina *Software Errors*) | Gravedad |
|---|---|---|
| F-26 | Funcional / flujo de control | Media |
| F-27 | Límites / validación de entrada | Media |
| F-28 | Seguridad / orden de autorización | Alta |
| F-29 | Datos / auditoría | Baja |
| F-30 | Manejo de estado | Media |

> Rutas: desde #72 las acciones de coordinación son `.../aprobacion`, `.../ajustes` y
> `.../resolucion` (antes `.../aprobar`, `.../ajustar` y `.../resolver`).

---

## F-26 · Un ajuste posterior a la resolución ya no se puede disputar

- **Test:** `backend/internal/domain/sesion/errores_flujo_e_test.go`, `TestF26AjusteTrasResolucionDebeSerDisputable`.
- **Archivo:** `backend/internal/domain/sesion/minutos.go` (`Disputar` y `fueResuelto`).
- **Reproducción GUI:**
  1. El tutor propone 45 minutos; coordinación aprueba; el escolar disputa; coordinación resuelve en 40.
  2. Coordinación hace un ajuste nuevo a 10 minutos (`POST /api/minutos/{id}/ajustes`, por ejemplo con Bruno).
  3. El tutor intenta disputar ese 10 desde "Minutos de la sesión": responde **409 `disputa_no_permitida`**.
- **Código actual:**
  ```go
  // La resolucion de una disputa es final: evita disputas en bucle.
  if r.fueResuelto() {
      return ErrDisputaNoPermitida
  }
  ```
  `fueResuelto()` devuelve `true` si **alguna vez** hubo un `resolver` en la auditoría.
- **Comportamiento esperado:** la resolución es final para **el valor resuelto**. Si después
  coordinación cambia el valor, ese cambio nuevo debe poder disputarse.
- **Guía de corrección:** que `fueResuelto()` mire la **última** acción que fijó el valor
  (`aprobar`, `ajustar` o `resolver`). Si es `resolver`, se bloquea; si es un `ajustar` posterior, se permite.

## F-27 · Un motivo con el carácter NUL produce 500 en vez de 422

- **Test:** `backend/internal/domain/sesion/errores_flujo_e_test.go`, `TestF27MotivoConCaracterNULDebeRechazarse`.
- **Archivos:** `backend/internal/domain/sesion/minutos.go` (`normalizarMotivo`) y
  `backend/internal/adapters/repositories/postgres/sesion/` (el `INSERT` en `sesion_minutos_auditoria`).
- **Reproducción GUI:**
  1. Como coordinación, enviar un ajuste con el motivo `"Revisado\u0000"` (Bruno o `curl`; en JSON, `\u0000` llega como NUL).
  2. La API responde **500 `error_interno`**. En el log de PostgreSQL:
     `invalid byte sequence for encoding "UTF8": 0x00` (verificado contra PostgreSQL 16).
- **Código actual:**
  ```go
  motivo = strings.TrimSpace(motivo)
  if motivo == "" || !utf8.ValidString(motivo) || utf8.RuneCountInString(motivo) > MaxCaracteresMotivo {
      return "", ErrMotivoInvalido
  }
  ```
  El NUL es UTF-8 válido, así que pasa.
- **Comportamiento esperado:** el dominio rechaza el motivo con `ErrMotivoInvalido` y la API responde **422**.
- **Guía de corrección:** rechazar en `normalizarMotivo` cualquier `unicode.IsControl(r)` salvo `\n` y `\t`.

## F-28 · Un usuario ajeno provoca una escritura antes de recibir 403

- **Test:** `backend/internal/adapters/http/sesion/errores_flujo_e_test.go`, `TestF28AjenoNoDebeProvocarEscriturasAntesDel403`.
- **Archivo:** `backend/internal/adapters/http/sesion/minutos.go` (`consultar`).
- **Reproducción GUI:**
  1. El tutor propone minutos; luego la reserva se cancela (flujo D).
  2. Un usuario que **no** participa (por ejemplo, un apoderado) abre `GET /api/sesion/{id}/minutos`.
  3. Recibe **403**, pero `GET /api/minutos/{id}/auditoria` (coordinación) ya muestra una entrada
     `anular` con la hora de **esa** petición: el registro quedó anulado en la base por culpa de un usuario sin permiso.
- **Código actual:**
  ```go
  registro, err := c.servicio.Consultar(r.Context(), r.PathValue("id"), c.ahora())
  if err == nil && !registro.Participa(usuario.ID) && !esCoordinacion(usuario) {
      err = modelo.ErrNoAutorizado
  }
  ```
  `Consultar` sincroniza con la cancelación y **persiste** la anulación antes de revisar el permiso.
- **Comportamiento esperado:** un usuario sin permiso no causa ningún efecto. Primero se autoriza; después se lee.
- **Guía de corrección:** leer el registro sin efectos (`ServicioMinutos.Vista`, agregado en #69),
  comprobar `Participa`/coordinación y recién entonces responder. Se corrige junto con F-06.

## F-29 · Ajustar al mismo valor deja una entrada de auditoría sin cambio

- **Test:** `backend/internal/domain/sesion/errores_flujo_e_test.go`, `TestF29AjusteSinCambioNoDebeAuditarse`.
- **Archivo:** `backend/internal/domain/sesion/minutos.go` (`Ajustar`).
- **Reproducción GUI:**
  1. Coordinación aprueba 45 minutos.
  2. Coordinación ajusta a 45 con cualquier motivo.
  3. `GET /api/minutos/{id}/auditoria` muestra una entrada `ajustar` con `valor_anterior = 45` y `valor_nuevo = 45`.
- **Código actual:**
  ```go
  if minutos < 0 || minutos > MaxMinutosSesion {
      return ErrMinutosInvalidos
  }
  ...
  r.verificar(AccionAjustar, coordinadorID, minutos, motivo, ahora)
  ```
  No compara el valor nuevo con el verificado.
- **Comportamiento esperado:** un ajuste sin cambio se rechaza (409) y no ensucia la bitácora inmutable.
- **Guía de corrección:** si `r.estado == MinutosVerificados` y `minutos == *r.verificados`, devolver `ErrTransicionMinutos`.

## F-30 · Restaurar acepta registros en estados imposibles

- **Test:** `backend/internal/domain/sesion/errores_flujo_e_test.go`, `TestF30RestaurarDebeRechazarEstadosInconsistentes`.
- **Archivo:** `backend/internal/domain/sesion/minutos.go` (`RestaurarRegistroMinutos`).
- **Reproducción (base de datos):**
  1. `UPDATE sesion_minutos SET estado = 'en_disputa', disputado_por = '' WHERE sesion_id = '...';`
  2. Abrir la bandeja de coordinación: el registro aparece "en disputa" **sin nadie que lo haya disputado**, y nada avisa del dato corrupto.
- **Código actual:** valida el estado, los rangos y que haya auditoría, pero no la coherencia entre campos:
  ```go
  if d.Estado == MinutosVerificados && d.Verificados == nil {
      return nil, ErrMinutosInvalidos
  }
  ```
- **Comportamiento esperado:** una fila incoherente se rechaza al cargarla (`ErrAuditoriaInconsistente`), en vez de
  convertirse en un agregado válido.
- **Guía de corrección:** exigir `DisputadoPor != ""` si el estado es `en_disputa`, y `Verificados == nil` si es `propuesto`.
  Como refuerzo, agregar un `CHECK` equivalente en `sesion_minutos`.
