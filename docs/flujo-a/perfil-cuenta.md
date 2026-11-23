# Perfil de cuenta

## Objetivo

Dar a todos los roles una ventana para consultar los datos de su cuenta y cambiar su
contraseña. El correo y el nombre de identificación son inmutables desde el perfil.

## Alcance

- Ruta frontend `/perfil`, disponible para cualquier usuario autenticado.
- Muestra de forma read-only:
  - nombre de identificación;
  - correo;
  - rol;
  - estado de confirmación del correo.
- Permite cambiar la contraseña con:
  - contraseña actual;
  - nueva contraseña;
  - confirmación.
- La nueva contraseña se valida con la política del rol (`básica` o `estricta`).
- El cambio se hace con `PUT /api/perfil/clave`.

No existe endpoint para cambiar `email` ni `nombre_uso`. Esas columnas no se actualizan
desde este flujo.

## Seguridad

- Se verifica la contraseña actual contra el hash almacenado.
- No se devuelve ni registra la contraseña en la respuesta.
- La nueva contraseña no puede ser igual a la actual.
- Tras el cambio, el hash se actualiza y el evento de cuenta se publica para auditoría
  operativa.
