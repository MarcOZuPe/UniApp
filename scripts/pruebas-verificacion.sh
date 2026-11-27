#!/usr/bin/env sh
# Ejecuta solo las pruebas de verificación.
#
# Las pruebas de defectos se mantienen en el repositorio, pero no forman parte
# de este comando: se ejecutan aparte para conservar la evidencia de errores.
#
# Uso:
#   ./scripts/pruebas-verificacion.sh          # backend + frontend
#   ./scripts/pruebas-verificacion.sh backend
#   ./scripts/pruebas-verificacion.sh frontend
set -eu

RAIZ="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
MODO="${1:-todo}"

ejecutar_backend() {
  echo "== Backend: pruebas de verificación =="
  cd "$RAIZ/backend"
  # Tests de defectos: Bxx, Fxx, TestDesign..., TestDecodificarCuerpo...
  # Se omiten aquí porque deben permanecer rojos mientras el defecto exista.
  go test -race ./... \
    -skip '^(TestB[0-9]|TestF[0-9]|TestDesign|TestDecodificarCuerpo)'
}

ejecutar_frontend() {
  echo "== Frontend: pruebas de verificación =="
  cd "$RAIZ/frontend"
  CI=true npx vitest run \
    src/adapters/http/apiModulos.test.js \
    src/application/acceso/servicios.test.js \
    src/presentation/navegacion.test.js \
    src/presentation/components/componentes.test.jsx \
    src/presentation/pages/acceso/GestionarCoordinadores.test.jsx \
    src/presentation/pages/acceso/Ingreso.test.jsx \
    src/presentation/pages/paginasSimples.test.jsx \
    src/presentation/pages/tutor/RevisarTutores.test.jsx
}

case "$MODO" in
  backend) ejecutar_backend ;;
  frontend) ejecutar_frontend ;;
  todo)
    ejecutar_backend
    ejecutar_frontend
    ;;
  *)
    echo "Modo no reconocido: $MODO (usa backend, frontend o todo)" >&2
    exit 2
    ;;
esac

echo "Pruebas de verificación completadas."
