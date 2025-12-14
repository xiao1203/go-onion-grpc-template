#!/usr/bin/env bash
set -euo pipefail

HOST="${1:?host}"
PORT="${2:-3306}"
USER="${3:-app}"
PASS="${4:-apppass}"

echo "Waiting for MySQL: ${HOST}:${PORT} ..."
for i in $(seq 1 60); do
  if mysqladmin ping -h "${HOST}" -P "${PORT}" -u"${USER}" -p"${PASS}" --silent >/dev/null 2>&1; then
    echo "MySQL is ready."
    exit 0
  fi
  sleep 1
done

echo "MySQL not ready in time" >&2
exit 1
