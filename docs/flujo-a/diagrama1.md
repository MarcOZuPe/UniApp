### Decorator

## 1. Archivos fuente necesarios

Todas las rutas son relativas a la raíz de `UniApp`.

| Archivo |
|---|
| `backend/internal/adapters/http/acceso/middleware.go` |
| `backend/internal/adapters/http/acceso/autorizacion_controlador.go` |
| `backend/internal/adapters/http/acceso/controlador_base.go` | 
| `backend/internal/adapters/http/acceso/controlador.go` | 
| `backend/internal/application/acceso/permiso.go` | 
| `backend/internal/domain/acceso/sesion.go` | 

## 2. Diagrama de clases 

```mermaid
classDiagram
    direction TB

    %% Componente del patrón Decorator
    class Handler {
        <<interface>>
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    %% Componente concreto: handler final de una ruta
    class HandlerFunc {
        <<concrete component>>
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    %% Tipo función que construye decoradores
    class Middleware {
        <<function>>
        +invoke(siguiente Handler) Handler
    }

    %% Decoradores concretos
    class decoradorRegistro {
        <<decorator>>
        -siguiente Handler
        -logger Logger
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    class decoradorRecuperar {
        <<decorator>>
        -siguiente Handler
        -logger Logger
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    class decoradorCabeceras {
        <<decorator>>
        -siguiente Handler
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    class decoradorOrigen {
        <<decorator>>
        -siguiente Handler
        -permitidos ListaDeOrigenes
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    class decoradorAutenticar {
        <<decorator>>
        -siguiente Handler
        -controlador controladorBase
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    class decoradorPermiso {
        <<decorator>>
        -siguiente Handler
        -requeridos ListaDePermisos
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    %% Fábricas de decoradores
    class FabricaDecoradores {
        <<factory functions>>
        +Registro(logger Logger) Middleware
        +Recuperar(logger Logger) Middleware
        +CabecerasSeguras() Middleware
        +OrigenPermitido(permitidos ListaDeOrigenes) Middleware
        +Autenticar(controlador controladorBase) Middleware
        +RequierePermiso(requeridos ListaDePermisos) Middleware
    }

    %% Cliente que compone la cadena
    class Encadenar {
        <<client helper>>
        +Encadenar(manejador Handler, middleware ListaDeMiddleware) Handler
    }

    %% Dependencias del flujo
    class controladorBase {
        +Autenticar() Middleware
    }

    class AutorizacionControlador {
        +Montar(mux ServeMux)
        +cambiarAutorizacion(respuesta ResponseWriter, peticion Request)
        +historialAutorizaciones(respuesta ResponseWriter, peticion Request)
    }

    class Permiso {
        <<value object>>
    }

    class escritorEstado {
        <<helper>>
        -ResponseWriter
        -estado int
        +WriteHeader(estado int)
    }

    %% El handler final y los decoradores implementan la misma interfaz
    Handler <|.. HandlerFunc
    Handler <|.. decoradorRegistro
    Handler <|.. decoradorRecuperar
    Handler <|.. decoradorCabeceras
    Handler <|.. decoradorOrigen
    Handler <|.. decoradorAutenticar
    Handler <|.. decoradorPermiso

    %% Cada decorador envuelve al siguiente Handler
    decoradorRegistro o--> Handler : siguiente (wrappee) >
    decoradorRecuperar o--> Handler : siguiente (wrappee) >
    decoradorCabeceras o--> Handler : siguiente (wrappee) >
    decoradorOrigen o--> Handler : siguiente (wrappee) >
    decoradorAutenticar o--> Handler : siguiente (wrappee) >
    decoradorPermiso o--> Handler : siguiente (wrappee) >

    %% Relaciones de apoyo
    FabricaDecoradores ..> Middleware : construye >
    FabricaDecoradores ..> decoradorRegistro : crea >
    FabricaDecoradores ..> decoradorRecuperar : crea >
    FabricaDecoradores ..> decoradorCabeceras : crea >
    FabricaDecoradores ..> decoradorOrigen : crea >
    FabricaDecoradores ..> decoradorAutenticar : crea >
    FabricaDecoradores ..> decoradorPermiso : crea >

    Middleware ..> Handler : recibe y devuelve >
    Encadenar ..> Handler : recibe y devuelve >
    Encadenar ..> Middleware : aplica >

    AutorizacionControlador ..> Encadenar : compone rutas >
    AutorizacionControlador --> controladorBase : usa >
    controladorBase ..> FabricaDecoradores : Autenticar() >

    decoradorAutenticar --> controladorBase : valida sesión >
    decoradorPermiso ..> Permiso : exige >
    decoradorRegistro ..> escritorEstado : mide estado HTTP >
```
