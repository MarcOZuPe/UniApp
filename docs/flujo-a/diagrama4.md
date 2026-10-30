# Observer

## 1. Archivos fuente necesarios

| Archivo |
|---|
| `backend/internal/application/acceso/infra_prueba_test.go` |
| `backend/internal/application/acceso/acceso_test.go` | 
| `backend/internal/domain/acceso/eventos.go` | 
| `backend/internal/domain/acceso/hasher.go` | 
| `backend/internal/application/acceso/servicio.go` |
| `backend/internal/infrastructure/eventos.go` | 

## 2. Diagrama de clases

```mermaid
classDiagram
    direction TB

    %% Puertos del dominio
    class Observador {
        <<interface>>
        +Actualizar(evento Event)
        +ID() string
    }

    class Sujeto {
        <<interface>>
        +Registrar(nombre string, observador Observador)
        +Desregistrar(nombre string, observador Observador)
        +Notificar(evento Event)
    }

    class PasswordHasher {
        <<interface>>
        +Hash(clave string) string, error
        +Verify(clavePlana string, hashAlmacenado string) bool
    }

    class Event {
        <<event>>
        +Nombre string
        +Payload Map
    }

    %% Dobles de prueba
    class observadorDePrueba {
        <<concrete observer>>
        -identificador string
        -alRecibir func(Event)
        +Actualizar(evento Event)
        +ID() string
    }

    class busDePrueba {
        <<concrete subject>>
        -observadores map[string][]Observador
        +Registrar(nombre string, observador Observador)
        +Desregistrar(nombre string, observador Observador)
        +Notificar(evento Event)
    }

    class hasherDePrueba {
        <<test adapter>>
        +Hash(clave string) string, error
        +Verify(clavePlana string, hashAlmacenado string) bool
    }

    %% Fábricas de fixtures
    class FuncionesDePrueba {
        <<functions>>
        +nuevoObservadorDePrueba(identificador string, alRecibir func(Event)) observadorDePrueba
        +nuevoBusDePrueba() busDePrueba
    }

    %% Consumidores de los dobles
    class PruebasDeAcceso {
        <<test module>>
        +TestFabricaPorRol(testingT TestingT)
        +TestRegistroIngresoYSesion(testingT TestingT)
        +TestConfirmacionDeCorreo(testingT TestingT)
        +TestRecuperacion(testingT TestingT)
    }

    %% Implementación de puertos
    Observador <|.. observadorDePrueba
    Sujeto <|.. busDePrueba
    PasswordHasher <|.. hasherDePrueba

    %% Relaciones Observer
    busDePrueba o--> Observador : notifica >
    Observador ..> Event : recibe >
    busDePrueba ..> Event : publica >
    observadorDePrueba ..> Event : usa >

    %% Fábricas
    FuncionesDePrueba ..> observadorDePrueba : crea >
    FuncionesDePrueba ..> busDePrueba : crea >
    FuncionesDePrueba ..> Event : recibe callback >

    %% Uso desde pruebas
    PruebasDeAcceso ..> FuncionesDePrueba : usa >
    PruebasDeAcceso ..> busDePrueba : registra observadores >
    PruebasDeAcceso ..> hasherDePrueba : configura >
```

