# I02 · Expediente de tutor y revisión mínima

## Alcance

El tutor presenta una descripción, cursos que puede enseñar, experiencia/formación y entre uno y cinco documentos PDF sintéticos (máximo 10 MiB por archivo). El perfil permanece privado mientras la revisión esté pendiente o rechazada. Coordinación puede aprobarlo indicando una fecha y hora de vencimiento o rechazarlo con motivo.

La aprobación deja de publicar el perfil al llegar a `vence_en`; no se requiere una tarea periódica para ocultar perfiles vencidos. Una revisión solo puede aplicarse una vez al expediente pendiente. Un tutor rechazado puede volver a presentar el expediente.

## Rutas

| Método | Ruta | Autorización | Propósito |
|---|---|---|---|
| `GET` | `/api/tutor/mi-expediente` | Tutor | Consultar el expediente propio y metadatos de sus documentos |
| `POST` | `/api/tutor/expediente` | Tutor | Presentar o volver a presentar el expediente (`multipart/form-data`) |
| `GET` | `/api/tutor/revision` | Coordinación | Listar expedientes pendientes |
| `POST` | `/api/tutor/{id}/revision` | Coordinación | Aprobar con `vence_en` o rechazar con `motivo_rechazo` |
| `GET` | `/api/tutor/documentos/{id}` | Tutor dueño o coordinación | Descargar un PDF privado |
| `GET` | `/api/tutor/perfiles/{id}` | Público | Obtener solo un perfil aprobado y vigente; nunca incluye documentos |

El contrato OpenAPI en `docs/openapi.yaml` es la referencia para cuerpos y respuestas.

## Persistencia y arquitectura

- El dominio y sus reglas viven en `backend/internal/models/tutor`.
- La API y autorización siguen MVC por flujo; el permiso de coordinación se verifica en el servidor.
- `Repositorio` aísla PostgreSQL; hay implementación en memoria para pruebas.
- `AlmacenDocumentos` es el puerto de archivos y `AlmacenDocumentosFS` el adaptador. Los nombres originales solo son metadatos: los archivos se guardan con identificadores aleatorios en `TUTOR_DOCUMENTS_DIR` (por defecto `backend/data/tutor-documents` al arrancar desde `backend/`).
- El almacenamiento no se sirve como contenido estático. La descarga requiere sesión y autorización; el directorio se ignora en Git.
- La migración `006_tutor_expediente` crea únicamente tablas del flujo B.

## Límites

La lista de campos y restricciones corresponde al alcance mínimo del issue I02; no recopila DNI ni datos personales adicionales. Los PDFs deben comenzar con la firma `%PDF-`. No se implementa análisis antimalware, almacenamiento externo ni publicación de ofertas/disponibilidad del flujo C.
