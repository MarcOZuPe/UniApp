# Administrador del sistema

## Objetivo

Contar con una cuenta impersonal de administración que gestione coordinadores
generales y consulte la bitácora de seguridad y administración.

## Alcance

- Rol `administrador`, sin facultad.
- Puede crear, listar y desactivar coordinadores generales de cualquier facultad.
- Puede consultar logs de seguridad y administración.
- Tiene acceso a `/perfil` para cambiar su propia contraseña.
- No se auto-registra, no se asigna facultad y no participa en coordinación,
  tutorías, reservas, incidentes ni catálogo.

La cuenta inicial de desarrollo es:

- correo: `admin@uni.edu.pe`
- nombre: `Administrador`
- clave: `SEED_PASSWORD`

En producción el correo institucional debe aprovisionarse fuera de la aplicación.
El cambio de administrador no forma parte de este alcance.

## APIs

| Método | Ruta | Uso |
|---|---|---|
| GET | `/api/administracion/coordinadores-generales` | Listar generales activos |
| POST | `/api/administracion/coordinadores-generales` | Crear general con facultad |
| DELETE | `/api/administracion/coordinadores-generales/{id}` | Desactivar general y revocar sesiones |
| GET | `/api/administracion/logs?limite=50` | Ver bitácora de seguridad/administración |

Todas requieren `administracion:acceder`.

## Bitácora

Tabla `logs_sistema`, append-only:

- categoría: `seguridad`, `administracion` o `sistema`;
- evento;
- actor y rol;
- IP de origen;
- detalle saneado;
- fecha.

Se registra como mínimo:

- `clave.cambiada` con la IP desde la que se cambió la contraseña;
- `coordinador_general.creado`;
- `coordinador_general.eliminado`.

La IP se toma de `X-Forwarded-For`, `X-Real-IP` o del socket según los encabezados
disponibles detrás del proxy.

La base rechaza `UPDATE`, `DELETE` y `TRUNCATE` sobre la bitácora. No existe endpoint
para modificarla o borrarla.
