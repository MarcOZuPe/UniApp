# Errores encontrados en el flujo E (F-26 a F-30)

Pruebas que **fallan a propósito**: cada una demuestra un bug real del flujo E (I17) y quedará en
verde cuando se corrija. Siguen el formato de las F-xx/B-xx de #66 (`// F-NN:` y `t.Fatalf("F-NN: ...")`)
y la lista de tipos de la lámina *Software Errors* (IntroTesting).

| Id | Tipo (Software Errors) | Dónde está el bug | Qué pasa | Prueba |
|---|---|---|---|---|
| F-26 | Funcional / flujo de control | `domain/sesion/minutos.go` · `Disputar` → `fueResuelto()` | Tras *Resolver*, coordinación puede *Ajustar* otra vez, pero ese ajuste nuevo ya no se puede disputar nunca | `domain/sesion/errores_flujo_e_test.go` |
| F-27 | Límites / validación de entrada | `domain/sesion/minutos.go` · `normalizarMotivo` | Un motivo con `\x00` es UTF-8 válido y pasa, pero PostgreSQL lo rechaza (`invalid byte sequence for encoding "UTF8": 0x00`) → 500 en vez de 422 | `domain/sesion/errores_flujo_e_test.go` |
| F-28 | Seguridad / orden de autorización | `adapters/http/sesion/minutos.go` · `consultar` | Un usuario ajeno hace `GET /api/sesion/{id}/minutos`: primero corre `Consultar` (que puede anular y **persistir**) y después recibe 403 | `adapters/http/sesion/errores_flujo_e_test.go` |
| F-29 | Datos / auditoría | `domain/sesion/minutos.go` · `Ajustar` | Ajustar al mismo valor crea una entrada "ajustar" sin ningún cambio | `domain/sesion/errores_flujo_e_test.go` |
| F-30 | Manejo de estado | `domain/sesion/minutos.go` · `RestaurarRegistroMinutos` | Acepta filas imposibles: `en_disputa` sin `disputadoPor`, o `propuesto` con minutos verificados | `domain/sesion/errores_flujo_e_test.go` |

## Cómo corregir cada uno (para el siguiente sprint)

- **F-26:** que `fueResuelto()` mire solo la **última** acción que fijó el valor; si después hubo un *Ajustar*, se puede disputar de nuevo.
- **F-27:** rechazar en `normalizarMotivo` los caracteres de control (`unicode.IsControl`) salvo el salto de línea.
- **F-28:** en `consultar`, autorizar antes de llamar al servicio (leer sin efectos, comprobar `Participa`/coordinación y recién después sincronizar). Va de la mano con F-06.
- **F-29:** en `Ajustar`, si el valor nuevo es igual al verificado, responder `ErrTransicionMinutos` (409) sin auditar.
- **F-30:** en `RestaurarRegistroMinutos`, exigir `DisputadoPor != ""` si el estado es `en_disputa` y `Verificados == nil` si es `propuesto`.

## Cómo ejecutarlas

```
cd backend
go test -run 'F2[6-9]|F30' -v ./internal/domain/sesion/ ./internal/adapters/http/sesion/
```

Resultado esperado hoy: **5 en rojo**, cada una con su mensaje `F-NN: ...`.
