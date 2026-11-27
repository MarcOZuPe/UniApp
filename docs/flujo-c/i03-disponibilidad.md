# Flujo C · Disponibilidad del tutor (I03)

**Responsable:** Jhon · **Estado:** implementado en I03; la búsqueda por curso y semana (I08) sigue pendiente.

## Qué permite

El tutor publica bloques semanales de 60 minutos y define su cupo semanal entre 180 y 360 minutos.
El sistema impide solapamientos, controla que la disponibilidad no exceda el cupo y no permite
reducir el cupo mientras existan reservas confirmadas futuras.

## Reglas

| Regla | Comportamiento |
|---|---|
| Duración de bloque | Exactamente 60 minutos, alineado al minuto |
| Solapamiento | Se rechaza si dos bloques del mismo tutor y día se cruzan |
| Rango del cupo | 180 a 360 minutos, múltiplo de 60 |
| Reducción del cupo | Permitida sin compromisos; rechazada con reservas futuras |
| Disponibilidad publicada | No puede superar el cupo nuevo |
| Zona horaria | America/Lima para el día y la hora; se guarda en minutos del día |

## API

| Método | Ruta | Quién |
|---|---|---|
| GET | /api/oferta/slots | Tutor autenticado |
| POST | /api/oferta/slots | Tutor autenticado |
| DELETE | /api/oferta/slots/{id} | Tutor dueño del bloque |
| PUT | /api/oferta/cupo | Tutor autenticado |

Errores: `403` si la cuenta no es de tutor, `409` por solapamiento o capacidad,
`422` por día u hora inválidos.

## Modelo de datos

- `slots`: tutor, día de la semana, minuto de inicio y de fin. Restricción de exclusión para el solapamiento.
- `cupos_semanales`: tutor y minutos, con actualización transaccional.

## Seguridad

- Solo el tutor gestiona su propia disponibilidad: el identificador sale de la sesión, no del cuerpo.
- El tutor solo puede borrar bloques que le pertenecen.
- Las fechas se presentan en America/Lima.

## Pruebas

- Dominio: validación de bloques, solapamiento y cambio de cupo.
- Repositorio: reducción con y sin compromisos, y capacidad frente a la disponibilidad publicada.
- Errores documentados: ver [errores-flujo-c.md](errores-flujo-c.md).
