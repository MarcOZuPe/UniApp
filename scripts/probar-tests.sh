#!/bin/sh
set -u

# Ejecuta los tests y muestra cada resultado en cuanto termina.
# Al final de cada área imprime fallos/total.

cd /src/backend || exit 1
go test -json ./... 2>/dev/null | node /usr/local/bin/resumir-backend
estado_backend=$?

cd /src/frontend || exit 1
CI=true ./node_modules/.bin/vitest run \
  --reporter=/usr/local/bin/reporter-tests-frontend.mjs \
  --silent 2>/dev/null
estado_frontend=$?

if [ "$estado_backend" -ne 0 ] || [ "$estado_frontend" -ne 0 ]; then
  exit 1
fi
exit 0
