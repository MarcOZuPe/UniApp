# Abstract factory, strategy 

## 5. Archivos fuente necesarios

| Archivo | 
|---|
| `backend/internal/domain/acceso/permiso.go` |
| `backend/internal/domain/acceso/fabrica_rol.go` |
| `backend/internal/domain/acceso/rol.go` | 
| `backend/internal/domain/acceso/perfil.go` |
| `backend/internal/domain/acceso/validacion.go` |
| `backend/internal/adapters/http/acceso/middleware.go` | 

## 2. Diagrama de clases

```mermaid
classDiagram
    direction TB

    %% Value Object y roles
    class Permiso {
        <<value object>>
        +string valor
    }

    class PermisosConstantes {
        <<constants>>
        +PermisoCoordinacionAcceder
        +PermisoExpedientePresentar
        +PermisoReservaCrear
        +PermisoAutorizacionOtorgar
        +PermisoCatalogoGestionar
        +PermisoCatalogoSolicitarColegio
    }

    class Rol {
        <<domain type>>
        +Permisos() []Permiso
        +TienePermiso(permisoBuscado Permiso) bool
    }

    %% Fábrica de roles
    class FabricaRol {
        <<interface>>
        +CrearPerfil(dominiosTutor string) PerfilRol
        +Permisos() []Permiso
    }

    class FabricaParaRol {
        <<factory method>>
        +FabricaParaRol(rol Rol) FabricaRol, error
    }

    class fabricaRolEscolar {
        <<concrete factory>>
        +CrearPerfil(dominiosTutor string) PerfilRol
        +Permisos() []Permiso
    }

    class fabricaRolApoderado {
        <<concrete factory>>
        +CrearPerfil(dominiosTutor string) PerfilRol
        +Permisos() []Permiso
    }

    class fabricaRolTutor {
        <<concrete factory>>
        +CrearPerfil(dominiosTutor string) PerfilRol
        +Permisos() []Permiso
    }

    class fabricaRolCoordinador {
        <<concrete factory>>
        +CrearPerfil(dominiosTutor string) PerfilRol
        +Permisos() []Permiso
    }

    class PerfilRol {
        <<interface>>
        +Rol() Rol
        +Politica() PoliticaClave
        +EsPublico() bool
        +ValidarExtra(solicitud SolicitudRegistro) Map
    }

    %% Productos concretos de PerfilRol
    class perfilBase {
        <<abstract>>
        -rol Rol
        -politica PoliticaClave
        -publico bool
        +Rol() Rol
        +Politica() PoliticaClave
        +EsPublico() bool
        +ValidarExtra(solicitud SolicitudRegistro) Map
    }

    class perfilEscolar {
        <<concrete product>>
    }

    class perfilApoderado {
        <<concrete product>>
    }

    class perfilTutor {
        <<concrete product>>
        -sufijos []string
        +ValidarExtra(solicitud SolicitudRegistro) Map
    }

    class perfilCoordinador {
        <<concrete product>>
    }

    class PoliticaClave {
        <<interface>>
        +Validar(clave string, email string) []string
    }

    class PoliticaBasica {
        <<strategy>>
        +Min int
        +Validar(clave string, email string) []string
    }

    class PoliticaEstricta {
        <<strategy>>
        +Min int
        +Validar(clave string, email string) []string
    }

    class SolicitudRegistro {
        <<domain>>
        +Email string
        +NombreUso string
        +Clave string
        +Rol Rol
    }

    class ErrValidacion {
        <<error>>
        +Campos Map
    }

    %% Helper del archivo
    class FuncionesPermiso {
        <<functions>>
        +permisosDeRol(rol Rol) []Permiso
    }

    %% Consumidor del permiso
    class decoradorPermiso {
        -siguiente Handler
        -requeridos []Permiso
        +ServeHTTP(respuesta ResponseWriter, peticion Request)
    }

    %% Relaciones del catálogo
    Rol ..> FuncionesPermiso : Permisos / TienePermiso >
    FuncionesPermiso ..> FabricaParaRol : solicita fábrica >
    FuncionesPermiso ..> Permiso : devuelve >
    FabricaParaRol ..> FabricaRol : crea >
    FabricaParaRol ..> Rol : evalúa >
    FabricaParaRol ..> ErrValidacion : rol desconocido >

    %% Implementaciones de la fábrica
    FabricaRol <|.. fabricaRolEscolar
    FabricaRol <|.. fabricaRolApoderado
    FabricaRol <|.. fabricaRolTutor
    FabricaRol <|.. fabricaRolCoordinador

    %% Productos concretos de PerfilRol
    PerfilRol <|.. perfilBase
    perfilBase <|-- perfilEscolar
    perfilBase <|-- perfilApoderado
    perfilBase <|-- perfilTutor
    perfilBase <|-- perfilCoordinador

    perfilBase --> PoliticaClave : politica >
    PoliticaClave <|.. PoliticaBasica
    PoliticaClave <|.. PoliticaEstricta
    perfilEscolar ..> PoliticaBasica : usa >
    perfilApoderado ..> PoliticaBasica : usa >
    perfilTutor ..> PoliticaEstricta : usa >
    perfilCoordinador ..> PoliticaEstricta : usa >
    perfilBase ..> SolicitudRegistro : valida >

    %% Productos y permisos
    FabricaRol ..> PerfilRol : crea >
    FabricaRol ..> Permiso : devuelve >
    fabricaRolEscolar ..> perfilEscolar : CrearPerfil >
    fabricaRolApoderado ..> perfilApoderado : CrearPerfil >
    fabricaRolTutor ..> perfilTutor : CrearPerfil >
    fabricaRolCoordinador ..> perfilCoordinador : CrearPerfil >
    fabricaRolEscolar ..> Permiso : ReservaCrear / SolicitarColegio >
    fabricaRolApoderado ..> Permiso : AutorizacionOtorgar / SolicitarColegio >
    fabricaRolTutor ..> Permiso : ExpedientePresentar >
    fabricaRolCoordinador ..> Permiso : CoordinacionAcceder / CatalogoGestionar >

    %% Permisos concretos
    PermisosConstantes ..> Permiso : instancia >

    %% Consumo desde middleware
    decoradorPermiso ..> Permiso : requiere >
    decoradorPermiso ..> Rol : TienePermiso >
```

