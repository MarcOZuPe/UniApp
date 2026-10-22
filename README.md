# UniApp

## Evidencias

| Evidencia del laboratorio | Estado |
|---|---|
| Hooks: rechazo y aceptación | Validación técnica automatizada; capturas del colaborador pendientes |
| Protección de main | Pendiente de configurar y capturar en GitHub |
| PR de registro al laboratorio | Pendiente |
| Issue → PR → revisión → merge → Done | Pendiente de colaboración real |

Plataforma académica de tutorías solidarias entre estudiantes UNI y escolares. Propuesta independiente, sin reconocimiento institucional presumido. El MVP es virtual y utiliza datos sintéticos hasta contar con las validaciones necesarias.

## Organización

Arquitectura MVC, con cada integrante responsable de un flujo completo:

| Flujo | Responsable | Alcance |
|---|---|---|
| A | Víctor | Acceso, cuentas, apoderados e incidentes |
| B | Miguel | Tutor y coordinación |
| C | Mateo | Oferta y apelaciones |
| D | Jhon | Reserva |
| E | Marco | Sesión, agenda, asistencia, minutos y reporte |

El lenguaje y framework se acuerdan en E02. Esta entrega configura la colaboración; no implementa ni declara terminada la aplicación. n8n y Telegram no son dependencias del MVP.

## Estructura del backend

```text
backend/
├── cmd/server/main.go       # composition root
├── migrations/              # migraciones SQL
├── fixtures/                # datos de apoyo
└── internal/
    ├── models/              # dominio, con subcarpeta por flujo
    ├── views/               # serialización JSON
    ├── controllers/         # HTTP y middleware
    ├── infra/               # hash, correo, bus de eventos y observadores
    ├── arquitectura/        # pruebas de reglas MVC
    └── config/              # configuración de entorno
```

`models` no importa `infra`; la infraestructura implementa los contratos del dominio y se inyecta en `main.go`.

## Preparación local

```bash
git config core.hooksPath .githooks
chmod +x .githooks/*
python3 scripts/test_hooks.py
```

Ramas permitidas: `main`, `feature/kebab-case`, `bugfix/kebab-case`.
Commits: `feat`, `fix`, `docs`, `test`, `refactor`, `style`, `chore`, `build`, `ci`, `perf`; título de hasta 72 caracteres.

Cada colaborador copia `.env.example` a `.env` (que no se versiona) para
ejecutar en su máquina; para el correo de recuperación en desarrollo puede usar
un buzón local (Mailpit) y así no compartir credenciales.

## Despliegue en el VPS

`APP_ENV=prod` exige que `FRONTEND_ORIGIN` sea la URL pública, y la clave SMTP se
monta como Docker secret (`SMTP_PASSWORD_FILE`) en vez de viajar como variable de
entorno. El detalle de cada variable está en `.env.example`.

## GitHub y tablero

El administrador crea issues y gestiona el tablero. Los colaboradores implementan issues asignados y abren PR con `Closes #N`. Para pasar por revisión, abrir el PR como borrador y después marcarlo listo para revisar. Una aprobación y resolución de conversaciones preceden al merge.

Configuración pendiente de cuenta: [guía](docs/CONFIGURACION_GITHUB.md).
El workflow usa un secreto `GH_TOKEN` y las variables `PROJECT_OWNER` y `PROJECT_NUMBER`. No ejecuta código de PR con ese secreto. La automatización admite PR del mismo repositorio; para forks el workflow sin secretos no actualizará el tablero.

## Colaboradores

Pendiente: cada compañero debe agregar su nombre mediante su propio PR. No se registran participaciones simuladas.
