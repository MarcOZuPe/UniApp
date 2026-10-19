#!/bin/sh
# Falla si hay secretos probables versionados. Se ejecuta en CI (job "secretos").
set -u
fallo=0
if git ls-files --error-unmatch .env >/dev/null 2>&1; then echo "ERROR: .env está versionado"; fallo=1; fi
if git ls-files | grep -E '(^|/)(id_rsa|.*\.pem|.*\.key)$' >/dev/null 2>&1; then echo "ERROR: llaves privadas versionadas"; fallo=1; fi
# Patrones de credenciales reales (el .env.example y los tests quedan fuera).
if git grep -nIE '(AKIA[0-9A-Z]{16}|-----BEGIN [A-Z ]*PRIVATE KEY-----|postgres://[^:@/ ]+:[^@ ]{8,}@)' -- . ':!.env.example' ':!*_test.go' ':!README.md' ':!docker-compose.yml'; then
  echo "ERROR: posible credencial en el código"; fallo=1
fi
[ "$fallo" -eq 0 ] && echo "verificación de secretos: OK"
exit $fallo
