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
# Archivos comprimidos versionados: se inspeccionan por si esconden un .env.
# Si no se pueden listar (p. ej. formato desconocido), se considera un riesgo.
listar_comprimido() {
  case "$1" in
    *.zip) unzip -Z1 "$1" 2>/dev/null || unzip -l "$1" 2>/dev/null ;;
    *.tar|*.tgz|*.tar.gz) tar -tf "$1" 2>/dev/null ;;
    *.7z) 7z l -ba "$1" 2>/dev/null ;;
    *.rar) unrar lb "$1" 2>/dev/null ;;
  esac
}
comprimidos=$(git ls-files | grep -E '\.(zip|7z|rar|tgz|tar\.gz|tar)$' || true)
for archivo in $comprimidos; do
  listado=$(listar_comprimido "$archivo")
  if [ -z "$listado" ]; then
    echo "ERROR: no se pudo inspeccionar $archivo (puede ocultar secretos)"; fallo=1
  elif printf '%s\n' "$listado" | grep -qE '(^|/)\.env$'; then
    echo "ERROR: $archivo contiene un .env versionado"; fallo=1
  fi
done
[ "$fallo" -eq 0 ] && echo "verificación de secretos: OK"
exit $fallo
