#!/usr/bin/env bash
set -euo pipefail

# gopls の CodeAction を使って自動リファクタ（modernize 等）を実行します。
# 既定は source.fixAll を各 .go ファイルに適用して書き込みます。
#
# 環境変数:
#   KIND  … CodeActionKind（既定: source.fixAll）
#   TITLE … アクションタイトルで絞り込み（正規表現）
#   MODE  … write|diff|list（既定: write）

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
cd "$ROOT_DIR"

# gopls のビルドキャッシュで権限エラーにならないよう、ローカルディレクトリに固定
export GOCACHE=${GOCACHE:-"$ROOT_DIR/.gocache"}
export GOMODCACHE=${GOMODCACHE:-"$ROOT_DIR/.gomodcache"}

KIND=${KIND:-"source.fixAll"}
TITLE=${TITLE:-""}
MODE=${MODE:-"write"}

# 実行方法: ローカル gopls がなければ docker compose の api サービス内で実行
USE_DOCKER=${USE_DOCKER:-0}
if ! command -v gopls >/dev/null 2>&1; then
  USE_DOCKER=1
fi

case "$MODE" in
  write) EXEC_FLAGS="-exec -write"; INFO="apply (write)" ;;
  diff)  EXEC_FLAGS="-exec -diff";  INFO="dry-run (diff)" ;;
  list)  EXEC_FLAGS="-list";        INFO="list actions" ;;
  *) echo "[modernize] unknown MODE: $MODE (use write|diff|list)"; exit 2 ;;
esac

echo "[modernize] kind=$KIND mode=$MODE $([ -n "$TITLE" ] && printf 'title="%s"' "$TITLE")"

# 対象ファイル一覧（生成物など一部を除外）
FILES_LIST=$(git ls-files "**/*.go" | grep -vE '^(gen/|vendor/|ent/|\.go-\w+/)' || true)

if [ -z "$FILES_LIST" ]; then
  echo "[modernize] 対象となる .go ファイルが見つかりませんでした"
  exit 0
fi

run_one_file_local() {
  local file="$1"
  if [ -n "$TITLE" ]; then
    # shellcheck disable=SC2086
    gopls codeaction -kind="$KIND" -title="$TITLE" $EXEC_FLAGS "$file" || true
  else
    # shellcheck disable=SC2086
    gopls codeaction -kind="$KIND" $EXEC_FLAGS "$file" || true
  fi
}

run_one_file_docker() {
  local file="$1"
  local in_container_cmd
  if [ -n "$TITLE" ]; then
    in_container_cmd="gopls codeaction -kind=\"$KIND\" -title=\"$TITLE\" $EXEC_FLAGS \"$file\" || true"
  else
    in_container_cmd="gopls codeaction -kind=\"$KIND\" $EXEC_FLAGS \"$file\" || true"
  fi
  docker compose run --rm -T api bash -lc "$in_container_cmd"
}

echo "$FILES_LIST" | while IFS= read -r f; do
  if [ "$USE_DOCKER" = "1" ]; then
    run_one_file_docker "$f"
  else
    run_one_file_local "$f"
  fi
done

echo "[modernize] done: $INFO"
