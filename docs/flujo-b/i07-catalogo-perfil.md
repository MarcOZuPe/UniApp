# I07 · Catálogo de cursos y perfil del tutor

## Alcance y datos

- La migración `008_catalogo` conserva importaciones versionadas de cursos y padrón escolar. Cada importación registra `fuente` y `importado_en` (hora UTC del servidor).
- No hay una fuente institucional integrada. La migración inicial registra expresamente `sintetico: fixture inicial I07`; esos colegios y cursos no representan datos reales.
- Coordinación puede registrar otra importación mediante JSON. No se consulta ni se simula una API externa. Sus cursos comienzan en estado `pendiente`; el padrón publicado corresponde a la importación más reciente.
- Coordinación aprueba o rechaza cada curso para el grado importado. Solo cursos aprobados de la última importación aparecen en el catálogo público. Una nueva importación no hereda aprobaciones anteriores.
- Escolares y apoderados pueden solicitar un colegio que no figure en el padrón. La solicitud queda local para coordinación; se marca incorporada al aparecer en una importación posterior. No se devuelve la identidad de quien la solicitó.
- El directorio de tutores consume I02. Solo incluye expedientes aprobados cuyo `vence_en` sea posterior al momento de la consulta. La respuesta pública contiene presentación, cursos y experiencia; no incluye correo, contactos, constancias, documentos, vencimiento ni estado de revisión.

## Rutas

| Método | Ruta | Acceso | Propósito |
|---|---|---|---|
| `GET` | `/api/catalogo` | Público | Fuente, fecha de importación, colegios y cursos aprobados vigentes |
| `POST` | `/api/catalogo/importaciones` | `catalogo:gestionar` | Registrar una importación de cursos y colegios (el formulario de coordinación genera la solicitud) |
| `GET` | `/api/catalogo/cursos/pendientes` | `catalogo:gestionar` | Obtener cursos de la importación actual que requieren revisión |
| `POST` | `/api/catalogo/cursos/{id}/revision` | `catalogo:gestionar` | Aprobar o rechazar un curso por grado |
| `POST` | `/api/catalogo/colegios/solicitudes` | Escolar o apoderado | Solicitar un colegio faltante |
| `GET` | `/api/catalogo/colegios/solicitudes` | `catalogo:gestionar` | Consultar solicitudes pendientes, sin identidad del solicitante |
| `GET` | `/api/tutor/perfiles` | Público | Listar perfiles I02 aprobados y vigentes con campos públicos |
| `GET` | `/api/tutor/perfiles/{id}` | Público | Consultar un perfil I02 aprobado y vigente con los mismos campos públicos |

En una importación, `fuente` identifica de forma explícita su procedencia,
`cursos` contiene elementos `{ "nombre": "...", "grado": "..." }`, y
`colegios` elementos `{ "nombre": "...", "distrito": "..." }`. Este es el
contrato interno de la API; la interfaz de coordinación presenta campos
normales y genera ese cuerpo sin mostrar JSON. El servidor valida límites,
duplicados y campos desconocidos; registra el momento de la importación. La
operación de persistencia y la incorporación de solicitudes coincidentes son
transaccionales.

## Decisiones de diseño

- `models/catalogo` contiene validación, estado de aprobación, servicio y puerto de repositorio; el controlador solo traduce HTTP y permisos.
- PostgreSQL e implementación en memoria comparten el puerto del dominio. La revisión solo afecta cursos de la importación actual y no puede repetirse.
- `catalogo:gestionar` se asigna a coordinación. El permiso de solicitud escolar se asigna a escolares y apoderados. Los controles se validan en el servidor.
- I07 amplía el puerto/servicio I02 con el directorio público. Un tipo de respuesta dedicado es una lista permitida: no se serializa el expediente completo.
- La vista React distingue el catálogo público de la administración de coordinación. Datos sintéticos se rotulan como tales; no se inventa una fuente externa.

## Verificación local

Desde `backend/`: `go test ./...` y `go vet ./...`.
Desde `frontend/`: `npm test` y `npm run build`.
