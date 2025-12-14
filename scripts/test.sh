#!/usr/bin/env bash
set -euo pipefail

# test DB を待つ（apiコンテナ内で実行される想定）
./scripts/wait-mysql.sh "${TEST_DB_HOST}" "${TEST_DB_PORT}" "${TEST_DB_USER}" "${TEST_DB_PASS}"

go test ./...
