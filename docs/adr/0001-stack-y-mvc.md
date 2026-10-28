# ADR 0001 · Stack y arquitectura MVC (E02)

**Estado:** reemplazado en su parte de organización por el [ADR 0002](0002-clean-architecture.md); el stack (React, Go, PostgreSQL) sigue vigente.

**Decisión.** Frontend React (Vite) · API Go (`net/http`, Go 1.22) · PostgreSQL 16. Aplicación web con MVC.
Tres capas (`models`, `views`, `controllers`) con una subcarpeta por flujo: `acceso` (A), `tutor` (B/I02), `catalogo` (B/I07), `oferta` (C), `reserva` (D), `sesion` (E).
En la API, la *vista* es la serialización JSON; en React, la vista son las pantallas.
La infraestructura técnica (hash de claves, correo, bus de eventos y observadores) vive en `internal/infra`; el dominio conserva solo sus contratos.

**Reglas (las hace cumplir `internal/arquitectura`).**
1. El modelo no importa controladores ni vistas.
2. Un flujo solo usa su propio modelo y el de `acceso` (usuario y rol). Los demás datos, por la API del dueño o por fixture.
3. Un controlador no importa el controlador de otro flujo (salvo `acceso`: `Autenticar` y `RequierePermiso`).
4. Reglas de negocio (48 h, cupo, solapamiento, minutos) solo en el modelo; el servidor es la autoridad temporal (RN-03).
5. `domain` no importa `internal/infrastructure`. La infraestructura implementa los puertos del dominio (`PasswordHasher`, `Mailer`, `Sujeto`) y se conecta en `cmd/api/main.go`.

**Migraciones.** `backend/migrations/NNN_flujo_tema.up.sql`, numeración global creciente; cada flujo crea y modifica solo sus tablas.
**Dependencias.** Una sola externa en el backend (`lib/pq`); lo demás es librería estándar.
**Consecuencia.** Cambiar de motor, hash o proveedor de correo afecta a un adaptador, no al dominio.
