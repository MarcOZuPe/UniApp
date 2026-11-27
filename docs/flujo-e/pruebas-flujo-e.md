# Pruebas del flujo E (agenda, asistencia, minutos y reporte)

Responsable: Marco · Sprint 3 · Medido el 26/11/2026 sobre `feature/e05-integracion`
(incluye `main` con #66 y #67, las pruebas restauradas de #68 e I18).

## 1. Estrategia

El objetivo de las pruebas es **revelar bugs**, no solo confirmar que el código funciona (IntroTesting).
Por eso hay dos tipos de prueba:

- **En verde (regresión):** fijan el comportamiento correcto de cada issue (I05, I10, I11, I17, I18, E05).
- **En rojo a propósito (F-xx / B-xx):** cada una demuestra un bug real y pasará a verde cuando se
  corrija. Las del flujo E son F-04, F-06, F-14 y F-17 (de Víctor, #66) y **F-26 a F-30** (mías, ver
  `errores-flujo-e.md`).

Se prueba en cada capa de la arquitectura Clean (ADR 0002):

| Nivel | Qué se prueba | Herramienta | Ejemplo |
|---|---|---|---|
| Unitaria · dominio | Reglas del agregado sin I/O (plazos, transiciones, auditoría, seudónimos) | `go test` | `minutos_i17_test.go`, `reporte_i18_test.go` |
| Unitaria · aplicación | Casos de uso con repositorios en memoria y puertos falsos | `go test` | `TestI18ServicioNoEscribeAlVerReservaCancelada` |
| Integración · HTTP | Controladores reales con login, permisos y códigos de estado | `net/http/httptest` | `TestI17FlujoCompletoHTTP` |
| Integración · persistencia | Repositorios contra **PostgreSQL 16 real**, con migraciones | `go test` + `TEST_DATABASE_URL` | `repo_postgres_test.go`, `reportes_i18_test.go` |
| Punta a punta | Sesión → asistencia → resultado → minutos → disputa → cancelación → reporte | `httptest` | `TestE05RecorridoSesionAReporte` |
| Frontend | Dominio JS, clientes HTTP y páginas React | Vitest + Testing Library | `ReporteMinutos.test.jsx` |
| Arquitectura | Las capas no se importan al revés | `go test` en CI | test de arquitectura |

**Trabajo colaborativo:** todo entra por PR con revisión cruzada (Víctor → Jhon → Marco → Víctor) y CI
(`gofmt`, `go vet`, `go test -race` con Postgres de servicio, Vitest). Los hooks de `.githooks`
validan Conventional Commits.

## 2. Cobertura (por sentencias)

### Backend, paquetes `sesion`

| Paquete | Cobertura |
|---|---|
| `domain/sesion` | **92.5 %** |
| `adapters/presentation/sesion` (CSV/PDF/JSON) | **85.6 %** |
| `adapters/http/sesion` | **84.6 %** |
| `application/sesion` | **76.8 %** |
| `adapters/repositories/postgres/sesion` | **62.3 %** |
| `adapters/repositories/memory/sesion` | 42.2 % (se usa como doble de prueba; se ejercita desde otros paquetes) |
| **Total del flujo E** | **76.7 %** |

```
cd backend
go test -coverprofile=c.out ./internal/.../sesion/...
go tool cover -func=c.out | tail -1
```

### Frontend, archivos del flujo E

| Archivo | Sentencias |
|---|---|
| `domain/sesion/plazos.js`, `domain/sesion/reporte.js` | 100 % |
| `adapters/http/minutosApi.js`, `adapters/http/reporteApi.js` | 100 % |
| `presentation/pages/sesion/ReporteMinutos.jsx` | 90.8 % |
| **Total** | **95.1 %** (24/24 pruebas en verde) |

```
cd frontend
npx vitest run src/domain/sesion src/adapters/http/reporteApi.test.js \
  src/presentation/pages/sesion/Minutos.test.jsx src/presentation/pages/sesion/ReporteMinutos.test.jsx \
  --coverage.enabled --coverage.reporter=text
```

Las dos cifras superan el mínimo del 50 % que pide la rúbrica.

## 3. Ejecución

| Grupo | Resultado |
|---|---|
| Pruebas en verde del flujo E (backend) | **91 pasan** |
| Rojas intencionales de #66 en el flujo E | 5 fallan (F-04, F-06 ×2, F-14, F-17), como se espera |
| Rojas intencionales propias (F-26 a F-30, PR #70) | 5 fallan, como se espera |
| E05 punta a punta | Pasa |
| Frontend del flujo E | 24/24 pasan |

Nota: en Windows, *Smart App Control* puede bloquear los `.test.exe` que genera `go test`. La medición se
hizo en Docker (`golang:1.22` + `postgres:16-alpine`), igual que en CI.

## 4. Casos de prueba clave por criterio de aceptación

| Issue | Criterio | Prueba |
|---|---|---|
| I17 | 45 minutos suman 45 | `TestI17CuarentaYCincoMinutosSuman45` |
| I17 | Nadie se autoaprueba | `TestI17NoAutoaprueba` |
| I17 | Una disputa congela (suma 0) | `TestI17FlujoCompletoHTTP` |
| I17 | Una cancelación suma 0 | `TestI17CancelacionHTTP` |
| I17 | Cada ajuste queda auditado | `TestI17FlujoCompletoHTTP` (4 entradas) |
| I18 | Solo coordinación descarga | `TestI18SoloCoordinacionDescarga` |
| I18 | El escolar sale seudonimizado | `TestI18DescargaCSVConConstancia`, `TestI18DescargaPDF` |
| I18 | Sin constancia no se entrega el archivo | `TestI18SinConstanciaNoSeEntrega` |
| I18 | Leer el reporte no escribe | `TestI18ServicioNoEscribeAlVerReservaCancelada` |
| E05 | Cancelación, sesión parcial, disputa y reporte | `TestE05RecorridoSesionAReporte` |
