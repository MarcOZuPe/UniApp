# Adapter, decorator y observer

## 1. Archivos fuente necesarios

Todas las rutas son relativas a la raíz de `UniApp`.

| Archivo | 
|---|
| `backend/internal/domain/acceso/mailer.go` |
| `backend/internal/infrastructure/mailer.go` |
| `backend/internal/infrastructure/observadores.go` |
| `backend/internal/domain/acceso/eventos.go` | 
| `backend/internal/application/acceso/servicio.go` |
| `backend/cmd/api/main.go` |

## 2. Diagrama de clases

```mermaid
classDiagram
    direction TB

    %% Contrato del dominio
    class Mailer {
        <<interface>>
        +Enviar(contexto Context, mensaje Mensaje) error
    }

    class Mensaje {
        <<parameter object>>
        +Para string
        +Asunto string
        +Cuerpo string
    }

    class Context {
        <<external>>
    }

    %% Adaptadores concretos
    class LogMailer {
        <<adapter>>
        +MostrarCuerpo bool
        +Enviar(contexto Context, mensaje Mensaje) error
    }

    class SMTPMailer {
        <<adapter>>
        +Host string
        +Puerto string
        +Usuario string
        +Clave string
        +Remitente string
        +transporte TransporteSMTP
        +Enviar(contexto Context, mensaje Mensaje) error
    }

    %% Decoradores del puerto
    class ReintentoMailer {
        <<decorator>>
        +Siguiente Mailer
        +Intentos int
        +Espera Duration
        +Enviar(contexto Context, mensaje Mensaje) error
    }

    class TiempoLimiteMailer {
        <<decorator>>
        +Siguiente Mailer
        +Limite Duration
        +Enviar(contexto Context, mensaje Mensaje) error
    }

    class TransporteSMTP {
        <<interface>>
        +EnviarCorreo(envio EnvioSMTP) error
    }

    %% Fábricas de decoradores
    class FuncionesMailer {
        <<functions>>
        +NuevoReintentoMailer(mailer Mailer, intentos int, espera Duration) Mailer
        +NuevoTiempoLimiteMailer(mailer Mailer, limite Duration) Mailer
    }

    %% Observadores que usan el puerto
    class Observador {
        <<interface>>
        +Actualizar(evento Event)
        +ID() string
    }

    class CorreoRecuperacion {
        <<observer>>
        -mailer Mailer
        -origen string
        +Actualizar(evento Event)
        +ID() string
    }

    class CorreoConfirmacion {
        <<observer>>
        -mailer Mailer
        -origen string
        +Actualizar(evento Event)
        +ID() string
    }

    class CorreoVinculacion {
        <<observer>>
        -mailer Mailer
        -origen string
        +Actualizar(evento Event)
        +ID() string
    }

    class Event {
        <<event>>
        +Nombre string
        +Payload Map
    }

    %% Implementación del puerto
    Mailer <|.. LogMailer
    Mailer <|.. SMTPMailer
    Mailer <|.. ReintentoMailer
    Mailer <|.. TiempoLimiteMailer

    ReintentoMailer o--> Mailer : siguiente >
    TiempoLimiteMailer o--> Mailer : siguiente >
    SMTPMailer --> TransporteSMTP : usa >

    %% Contrato
    Mailer ..> Mensaje : recibe >
    Mailer ..> Context : recibe >
    LogMailer ..> Mensaje : recibe >
    SMTPMailer ..> Mensaje : recibe >

    %% Fábricas
    FuncionesMailer ..> ReintentoMailer : crea >
    FuncionesMailer ..> TiempoLimiteMailer : crea >
    FuncionesMailer ..> Mailer : devuelve >

    %% Observadores
    Observador <|.. CorreoRecuperacion
    Observador <|.. CorreoConfirmacion
    Observador <|.. CorreoVinculacion
    CorreoRecuperacion --> Mailer : envía >
    CorreoConfirmacion --> Mailer : envía >
    CorreoVinculacion --> Mailer : envía >
    CorreoRecuperacion ..> Event : recibe >
    CorreoConfirmacion ..> Event : recibe >
    CorreoVinculacion ..> Event : recibe >
    CorreoRecuperacion ..> Mensaje : construye >
    CorreoConfirmacion ..> Mensaje : construye >
    CorreoVinculacion ..> Mensaje : construye >
```

