# Singleton y observer

## 1. Archivos fuente necesarios

| Archivo |
|---|
| `frontend/src/application/acceso/almacenSesion.js` |
| `frontend/src/presentation/hooks/useSesion.js` |
| `frontend/src/application/acceso/sesion.js` |
| `frontend/src/adapters/http/apiAcceso.js` |
| `frontend/src/presentation/pages/acceso/Ingreso.jsx` |
| `frontend/src/presentation/pages/acceso/Inicio.jsx` |
| `frontend/src/presentation/pages/acceso/Coordinacion.jsx` |
| `frontend/src/presentation/components/RutaProtegida.jsx` |

## 2. Diagrama de clases

```mermaid
classDiagram
    direction TB

    %% Singleton
    class AlmacenSesionSingleton {
        <<singleton state>>
        -instancia AlmacenSesion
    }

    class getInstance {
        <<function>>
        +getInstance() AlmacenSesion
    }

    class crearAlmacen {
        <<private function>>
        +crearAlmacen() AlmacenSesion
    }

    class almacenSesion {
        <<singleton instance>>
    }

    %% Estado observable
    class AlmacenSesion {
        <<observable>>
        -estadoSesion EstadoSesion
        -suscriptores SetSuscriptores
        +getState() EstadoSesion
        +subscribe(suscriptor Suscriptor) Desuscribir
        +set(cambios CambiosEstado)
        -notificarSuscriptores()
    }

    class EstadoSesion {
        <<state>>
        +usuario Usuario
        +cargando bool
    }

    class CambiosEstado {
        <<parameter object>>
        +usuario Usuario
        +cargando bool
    }

    class Suscriptor {
        <<function>>
        +invoke()
    }

    class Desuscribir {
        <<function>>
        +invoke()
    }

    %% Consumidores
    class useSesionModulo {
        <<hook module>>
        +useSesion() EstadoSesion
    }

    class sesionModulo {
        <<module>>
        +cargarSesion()
        +ingresar(credenciales)
        +salir()
        +guardarUsuarioEnSesion(usuario)
        +cerrarSesionLocal()
    }

    class apiAccesoModulo {
        <<module>>
        +api
        +cerrarSesionLocal()
    }

    class IngresoModulo {
        <<page>>
        +enviarIngreso(evento)
    }

    class useSyncExternalStore {
        <<react API>>
        +useSyncExternalStore(subscribe, getState)
    }

    class Usuario {
        <<domain>>
        +id string
        +rol string
        +nombre_uso string
    }

    %% Relaciones Singleton
    getInstance ..> crearAlmacen : usa si no existe >
    getInstance ..> AlmacenSesion : retorna >
    getInstance --> almacenSesion : inicializa >
    AlmacenSesionSingleton --> AlmacenSesion : instancia >
    almacenSesion --> AlmacenSesion : instancia >

    %% Relaciones Observer
    AlmacenSesion o--> EstadoSesion : estadoSesion >
    AlmacenSesion o--> Suscriptor : notifica >
    AlmacenSesion ..> CambiosEstado : recibe >
    AlmacenSesion ..> Desuscribir : devuelve >
    EstadoSesion --> Usuario : usuario >

    %% Consumidores
    useSesionModulo --> almacenSesion : observa >
    useSesionModulo ..> useSyncExternalStore : usa >
    useSesionModulo ..> AlmacenSesion : getState / subscribe >
    sesionModulo --> almacenSesion : actualiza >
    sesionModulo ..> CambiosEstado : construye >
    sesionModulo ..> Usuario : recibe >
    apiAccesoModulo --> almacenSesion : cierra sesión >
    IngresoModulo ..> sesionModulo : ingresar >
```
