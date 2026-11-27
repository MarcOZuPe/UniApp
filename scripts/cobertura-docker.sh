#!/usr/bin/env bash
# Genera un log detallado y un resumen de cobertura de backend y frontend.
#
# Uso:
#   ./scripts/cobertura-docker.sh
#   ./scripts/cobertura-docker.sh --no-build
#   ./scripts/cobertura-docker.sh --ignorar-tests
#
# Salidas:
#   coverage/backend/backend-tests.log
#   coverage/backend/backend-resumen.txt
#   coverage/frontend/frontend-tests.log
#   coverage/frontend/frontend-resumen.txt

set -u

RAIZ="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COBERTURA="$RAIZ/coverage"
COMPOSE=(docker compose -f "$RAIZ/docker-compose.tests.yml")

construir=1
ignorar_tests=0
for argumento in "$@"; do
  case "$argumento" in
    --no-build) construir=0 ;;
    --ignorar-tests) ignorar_tests=1 ;;
    -h|--help)
      sed -n '2,15p' "$0"
      exit 0
      ;;
    *)
      echo "Argumento no reconocido: $argumento" >&2
      exit 2
      ;;
  esac
done

rm -rf "$COBERTURA/backend" "$COBERTURA/frontend" 2>/dev/null || true
mkdir -p "$COBERTURA/backend" "$COBERTURA/frontend"

if [ "$construir" -eq 1 ]; then
  echo "Construyendo imagen de tests"
  "${COMPOSE[@]}" build tests || exit 1
fi

echo
echo "Backend"
"${COMPOSE[@]}" run --rm \
  -e GO_SKIP_TESTS="${GO_SKIP_TESTS:-}" \
  -e IGNORAR_TESTS="$ignorar_tests" \
  -v "$COBERTURA/backend:/cobertura" \
  tests sh -c '
    cd /src/backend || exit 1
    set +e
    rm -rf /cobertura/*
    salida=/cobertura/backend-tests.log
    perfil=/tmp/backend.out
    if [ "${IGNORAR_TESTS:-0}" = "1" ]; then
      go test ./... -v -run "^$" -coverpkg=./... -coverprofile="$perfil" >"$salida" 2>&1
    elif [ -n "${GO_SKIP_TESTS:-}" ]; then
      go test ./... -v -skip "$GO_SKIP_TESTS" -coverpkg=./... -coverprofile="$perfil" >"$salida" 2>&1
    else
      go test ./... -v -coverpkg=./... -coverprofile="$perfil" >"$salida" 2>&1
    fi
    aprobados=$(grep -c "^--- PASS" "$salida")
    fallidos=$(grep -c "^--- FAIL" "$salida")
    cobertura="0.0%"
    if [ -f "$perfil" ]; then
      go tool cover -func="$perfil" > /tmp/backend-func.txt
      cobertura=$(awk "/^total:/{print \$3}" /tmp/backend-func.txt)
      {
        echo
        echo "== Detalle de cobertura =="
        cat /tmp/backend-func.txt
      } >> "$salida"
    fi
    {
      echo "Aprobados: $aprobados"
      echo "Fallidos: $fallidos"
      echo "Cobertura: $cobertura"
    } > /cobertura/backend-resumen.txt
    {
      echo
      echo "Aprobados: $aprobados"
      echo "Fallidos: $fallidos"
      echo "Cobertura: $cobertura"
    } >> "$salida"
    cat /cobertura/backend-resumen.txt
    echo "Resumen: backend-resumen.txt"
    echo "Log: backend-tests.log"
    exit 0
  '

echo
echo "Frontend"
"${COMPOSE[@]}" run --rm \
  -e VITEST_EXCLUDE="${VITEST_EXCLUDE:-}" \
  -e IGNORAR_TESTS="$ignorar_tests" \
  -v "$COBERTURA/frontend:/cobertura" \
  tests sh -c '
    cd /src/frontend || exit 1
    set +e
    rm -rf /cobertura/*
    archivo_vacio=""
    if [ "${IGNORAR_TESTS:-0}" = "1" ]; then
      archivo_vacio=/src/frontend/.cobertura-sin-tests.test.js
      printf "import { test } from \x27vitest\x27;\ntest(\x27cobertura sin tests\x27, () => {});\n" > "$archivo_vacio"
    fi
    argumentos=""
    if [ -n "${VITEST_EXCLUDE:-}" ]; then
      antiguo_ifs=$IFS
      IFS=","
      for patron in $VITEST_EXCLUDE; do
        [ -n "$patron" ] && argumentos="$argumentos --exclude=$patron"
      done
      IFS=$antiguo_ifs
    fi
    salida=/cobertura/frontend-tests.log
    json=/tmp/frontend-tests.json
    npm run test:coverage -- $argumentos $archivo_vacio \
      --coverage.reporter=text --coverage.reporter=text-summary \
      --reporter=default --reporter=json --outputFile="$json" >"$salida" 2>&1
    [ -n "$archivo_vacio" ] && rm -f "$archivo_vacio"

    conteo=$(node -e "try{const d=require(\"$json\"); process.stdout.write(d.numPassedTests+\" \"+d.numFailedTests)}catch(e){process.stdout.write(\"0 0\")}")
    aprobados=${conteo% *}
    fallidos=${conteo#* }

    sed -r "s/\x1B\[[0-9;]*[mK]//g" "$salida" > /tmp/frontend-limpio.txt
    cobertura=$(sed -n "s/^Statements[[:space:]]*:[[:space:]]*\([0-9.]*\)%.*/\1%/p" /tmp/frontend-limpio.txt | head -1)
    [ -z "$cobertura" ] && cobertura="0%"

    {
      echo "Aprobados: $aprobados"
      echo "Fallidos: $fallidos"
      echo "Cobertura: $cobertura"
    } > /cobertura/frontend-resumen.txt
    {
      echo
      echo "Aprobados: $aprobados"
      echo "Fallidos: $fallidos"
      echo "Cobertura: $cobertura"
    } >> "$salida"
    cat /cobertura/frontend-resumen.txt
    echo "Resumen: frontend-resumen.txt"
    echo "Log: frontend-tests.log"
    exit 0
  '
