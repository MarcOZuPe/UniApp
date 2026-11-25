#!/usr/bin/env bash
set -u -o pipefail

# Ejecuta la suite en Docker y muestra los resultados en vivo.
# Solo deja pasar las líneas PASS/FAIL y los totales finales.
docker compose -f docker-compose.tests.yml run --rm --build --quiet-build tests 2>&1 \
  | grep --line-buffered -E '^(PASS|FAIL) |^(Backend|Frontend): '
