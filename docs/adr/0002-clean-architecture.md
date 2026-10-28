# ADR 0002 · Migración a Clean Architecture por feature

**Estado:** aceptado.

**Contexto.** El backend había crecido con carpetas horizontales (`models`, `views`,
`controllers`, `infra`, `adapters`) y una subcarpeta por flujo. Eso obligaba a saltar
entre cinco carpetas para tocar una sola feature y mezclaba entidades, casos de uso y
persistencia dentro de `models`. El frontend replicaba el mismo esquema MVC.

**Decisión.** Organizar el código en las cuatro capas de Clean Architecture, con una
subcarpeta por feature dentro de cada capa:

```text
domain/          entidades, reglas de negocio y puertos
application/     casos de uso y orquestación
adapters/        HTTP, presentación y repositorios
infrastructure/  config, criptografía, correo, storage, eventos y Swagger
```

En el frontend:

```text
app/             arranque, router y providers
domain/          roles, validadores y errores
application/     casos de uso del cliente
adapters/        clientes HTTP y almacenamiento de sesión
infrastructure/  fetch y configuración
presentation/    páginas, componentes, hooks y estilos
```

**Reglas de dependencia.**

1. `domain` no importa `application`, `adapters` ni `infrastructure`.
2. `application` importa `domain`; no importa adaptadores ni infraestructura.
3. `adapters` importa `domain`/`application`; no importa `infrastructure`.
4. `infrastructure` contiene detalles técnicos y se conecta en `cmd/api/main.go`.
5. Una feature no importa el dominio de otra; la excepción es `acceso` (identidad).
6. `presentation` no importa `adapters` ni `infrastructure`; usa `application`.
7. Solo `cmd/api/main.go` importa repositorios concretos.

Las reglas 1–3 y 7 se verifican en `backend/tests/architecture/arquitectura_test.go`.

**Consecuencias.**

- Cada feature se puede probar sin base de datos ni HTTP.
- Cambiar Postgres, SMTP o el sistema de archivos afecta a un adaptador.
- El composition root concentra el cableado.
- Los archivos `aliases.go` de `application` y de los repositorios reexportan los tipos
  del dominio para mantener compatibilidad durante la migración. Se pueden eliminar
  cuando el código use calificadores `dominio.` de forma directa.
