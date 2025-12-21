#!/usr/bin/env bash
set -euo pipefail

NAME=${1:-}
DROP=${2:-0}

if [[ -z "${NAME}" ]]; then
  echo "Usage: scripts/clear.sh <Name> [drop]" >&2
  exit 1
fi

to_snake() {
  # CamelCase -> snake_case (simple)
  local s="$1"
  echo "$s" | sed -E 's/([A-Z])/_\1/g' | sed -E 's/^_//' | tr '[:upper:]' '[:lower:]'
}

SNAKE=$(to_snake "$NAME")
TABLE="$SNAKE"
[[ "${TABLE}" == *s ]] || TABLE="${TABLE}s"

echo "[clear] removing generated files for: ${NAME} (${SNAKE})"
rm -rf \
  "proto/${SNAKE}" \
  "gen/${SNAKE}" \
  "internal/domain/${SNAKE}.go" \
  "internal/usecase/${SNAKE}_usecase.go" \
  "internal/adapter/grpc/${SNAKE}_handler.go" \
  "internal/adapter/grpc/${SNAKE}_routes.go" \
  "internal/adapter/repository/memory/${SNAKE}_repository.go" \
  "internal/adapter/repository/mysql/${SNAKE}_repository.go" \
  2>/dev/null || true

echo "[clear] pruning db/schema.sql entries for: ${NAME}"
if [[ -f db/schema.sql ]]; then
  awk -v NM="$NAME" 'BEGIN{drop=0} { 
    if ($0 ~ "^--[[:space:]]*" NM "[[:space:]]+table[[:space:]]*$") {drop=1; next}
    if (drop) { if ($0 ~ /\) ENGINE=/) {drop=0; next} next }
    print }' db/schema.sql | \
  awk -v TBL="$TABLE" 'BEGIN{drop=0} {
    if ($0 ~ "^CREATE[[:space:]]+TABLE[[:space:]]+`?" TBL "`?[[:space:]]*\\(") {drop=1; next}
    if (drop) { if ($0 ~ /\) ENGINE=/) {drop=0; next} next }
    print }' > db/schema.sql.tmp && mv db/schema.sql.tmp db/schema.sql
fi

echo "[clear] pruning cmd/server/main.go entries for: ${NAME}"
if [[ -f cmd/server/main.go ]]; then
  # remove connect import for the entity
  sed -E -e "/gen\/${SNAKE}\/v1\/.*v1connect/d" -i '' cmd/server/main.go 2>/dev/null || sed -E -e "/gen\/${SNAKE}\/v1\/.*v1connect/d" -i cmd/server/main.go 2>/dev/null || true
  # remove route block starting with "// Name scaffold" until blank line
  awk -v NAME="$NAME" 'BEGIN{drop=0} {
    if ($0 ~ "^//[[:space:]]*" NAME "[[:space:]]+scaffold") {drop=1; next}
    if (drop) { if ($0 ~ /^[[:space:]]*$/) {drop=0; next} next }
    print }' cmd/server/main.go > cmd/server/main.go.tmp && mv cmd/server/main.go.tmp cmd/server/main.go
  # drop infra imports if alias not used
  if ! grep -q "mysqlrepo\." cmd/server/main.go 2>/dev/null; then
    sed -E -e "/internal\/adapter\/repository\/mysql/d" -i '' cmd/server/main.go 2>/dev/null || true
  fi
  if ! grep -q "inframysql\." cmd/server/main.go 2>/dev/null; then
    sed -E -e "/internal\/infra\/mysql/d" -i '' cmd/server/main.go 2>/dev/null || true
  fi
  go fmt cmd/server/main.go >/dev/null 2>&1 || true
fi

if [[ "${DROP}" == "1" ]]; then
  echo "[clear] applying DROP via mysqldef (--enable-drop)"
  make -s migrate DROP_FLAGS="--enable-drop" || true
else
  echo "[clear] (tip) run: make migrate DROP_FLAGS=\"--enable-drop\""
fi

docker compose restart api >/dev/null 2>&1 || true
echo "[clear] done"
