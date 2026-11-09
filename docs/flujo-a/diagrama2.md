# Observer 

## 1. Archivos fuente necesarios

| Archivo | 
|---|
| `backend/internal/adapters/http/acceso/controlador_vinculacion_test.go` | 
| `backend/internal/adapters/http/acceso/controlador_test.go` | 
| `backend/internal/domain/acceso/eventos.go` | 
| `backend/internal/infrastructure/eventos.go` | 
| `backend/internal/application/acceso/vinculacion_service.go` | 
| `backend/internal/application/acceso/servicio.go` | 

## 2. Diagrama de clases

```mermaid
classDiagram
    direction TB

    %% Evento publicado por el sujeto
    class Event {
        +Nombre string
        +Payload map[string]any
    }

    %% Interfaz Observer
    class Observador {
        <<interface>>
        +Actualizar(evento Event)
        +ID() string
    }

    %% Interfaz Subject
    class Sujeto {
        <<interface>>
        +Registrar(nombre string, observador Observador)
        +Desregistrar(nombre string, observador Observador)
        +Notificar(evento Event)
    }

    %% Sujeto concreto
    class Bus {
        -mu sync.RWMutex
        -observadores map[string][]Observador
        +Registrar(nombre string, observador Observador)
        +Desregistrar(nombre string, observador Observador)
        +Notificar(evento Event)
    }

    %% Observador concreto creado por la prueba
    class ObservadorFunc {
        -identificador string
        -alRecibir func(Event)
        +Actualizar(evento Event)
        +ID() string
    }

    %% Productor del evento
    class VinculacionService {
        +SolicitarVinculacion(...)
        -notificarSolicitud(apoderado, escolar, token)
    }

    %% Infraestructura de prueba
    class entornoPruebas {
        +manejador http.Handler
        +bus Bus
    }

    %% Clientes del patrón
    class TestVinculacionApoderadoYEscolar {
        +TestVinculacionApoderadoYEscolar(testingT)
    }

    class TestAutorizacionDeNuevasReservas {
        +TestAutorizacionDeNuevasReservas(testingT)
    }

    Observador <|.. ObservadorFunc
    Sujeto <|.. Bus

    Bus o--> Observador : notifica a >
    ObservadorFunc ..> Event : recibe >
    Sujeto ..> Event : publica >

    VinculacionService ..> Sujeto : notifica VinculacionSolicitada >
    VinculacionService ..> Event : construye >

    entornoPruebas --> Bus : expone >
    TestVinculacionApoderadoYEscolar ..> entornoPruebas : usa >
    TestAutorizacionDeNuevasReservas ..> entornoPruebas : usa >

    TestVinculacionApoderadoYEscolar ..> ObservadorFunc : registra closure >
    TestAutorizacionDeNuevasReservas ..> ObservadorFunc : registra closure >
```

