# Despliegue en el VPS y manejo del secreto SMTP

El backend lee la clave SMTP de `SMTP_PASSWORD` o, si se define
`SMTP_PASSWORD_FILE`, del archivo indicado. En el VPS usamos la segunda forma
con **Docker secrets**, para que la clave no quede como variable de entorno (no
aparece en `docker inspect` ni en `docker compose config`).

> Nota: los `docker-compose*.yml` son el *pegamento* de despliegue y se
> mantienen fuera de este repositorio de código; aquí se documenta el
> procedimiento y las variables que usa el backend.

> Límite realista: quien tenga `root` o esté en el grupo `docker` puede leer
> cualquier secreto en caliente. Estas medidas evitan filtraciones accidentales
> (Git, backups, salidas de comandos) y limitan quién lo lee en disco.

## 1. `.env` del servidor (sin la clave)

En `/opt/uniapp/.env` (fuera del repo, `chmod 600`):

```dotenv
APP_ENV=prod
FRONTEND_ORIGIN=http://IP_DEL_VPS:5173     # o https://tu-dominio con TLS
POSTGRES_PASSWORD=<clave de la base>
SEED_SYNTHETIC=false
TUTOR_EMAIL_DOMAIN=uni.edu.pe,uni.pe
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=vcadillog@uni.pe
SMTP_FROM=vcadillog@uni.pe
SMTP_PASSWORD=                              # vacío: la clave va en el secreto
```

Con `APP_ENV=prod` el arranque falla si `FRONTEND_ORIGIN` apunta a `localhost`
(evita el problema de INC-02: enlaces de recuperación irresolubles).

## 2. Crear el secreto

```bash
install -d -m 700 /opt/uniapp/secrets
printf '%s' 'TU_CLAVE_SMTP' > /opt/uniapp/secrets/smtp_password
chmod 600 /opt/uniapp/secrets/smtp_password
```

## 3. Levantar con el override de producción

`docker-compose.prod.yml` monta el secreto en `/run/secrets/smtp_password` y
define `SMTP_PASSWORD_FILE`:

```bash
cd /opt/uniapp
docker compose -f docker-compose.yml -f docker-compose.override.yml \
  -f docker-compose.prod.yml up -d --build
```

## 4. Verificación

```bash
# La clave no debe verse en la configuración del contenedor:
docker inspect uniapp-c12-api-1 --format '{{range .Config.Env}}{{println .}}{{end}}' \
  | grep -E 'SMTP_PASSWORD|SMTP_PASSWORD_FILE'
#   SMTP_PASSWORD=
#   SMTP_PASSWORD_FILE=/run/secrets/smtp_password

# El archivo del secreto sí está dentro del contenedor:
docker exec uniapp-c12-api-1 wc -c /run/secrets/smtp_password

# Arranque correcto:
docker logs uniapp-c12-api-1 | grep 'frontend de enlaces'
```

Después pide una recuperación con un correo real y comprueba que llega.

## Permisos y endurecimiento

- `.env` y `secrets/smtp_password`: `600`; directorios `700`; `umask 077` al
  escribirlos.
- Despliega con un usuario no-root. Recuerda que pertenecer al grupo `docker`
  equivale a `root` en la práctica.
- `secrets/` está en `.gitignore`: nunca se sube al repositorio.
- La clave se rota regenerando el secreto y reiniciando `api`.

## Alternativa sin contraseña (la más segura)

Si se configura el **relay SMTP por IP**, el contenedor no guarda credenciales:

- Google Workspace: `smtp-relay.gmail.com:587` con la IP del VPS autorizada por
  el administrador → `SMTP_USER` y `SMTP_PASSWORD` vacíos.
- Amazon SES: autorización por rol IAM (su interfaz SMTP sí pide credenciales).

Con esto no hay ningún secreto SMTP que proteger.

## GitHub Secrets

Los secretos de GitHub solo existen dentro de un workflow (CI/CD); no los lee el
contenedor ni las máquinas del equipo. Sirven si se automatiza el despliegue:
un workflow puede escribir `/opt/uniapp/secrets/smtp_password` por SSH usando
`${{ secrets.SMTP_PASSWORD }}`.
