.PHONY: up down logs sh test restart \
        protogen proto scaffold scaffold-all clear \
        migrate migrate-dev migrate-test \
        dry-run dry-run-dev dry-run-test \
        reset-test-db reset-dev-db \
        lint

# ---- dev-only guard -----------------------------------------------
# 破壊的/生成系ターゲットは開発環境のみ実行可能にします。
# 許可条件（いずれかを満たす）
#  - 環境変数 APP_ENV=dev
#  - 環境変数 ALLOW_DEV=1（明示許可）
#  - リポジトリ直下に .dev-allow ファイルが存在
APP_ENV ?=
ALLOW_DEV ?= 0
DEV_ALLOW_FILE ?= .dev-allow
define dev_only
	@if [ "$(APP_ENV)" != "dev" ] && [ "$(ALLOW_DEV)" != "1" ] && [ ! -f "$(DEV_ALLOW_FILE)" ]; then \
	  echo "[guard] このターゲットは開発環境専用です。APP_ENV=dev または ALLOW_DEV=1 を指定するか、$(DEV_ALLOW_FILE) を作成してください。"; \
	  exit 1; \
	fi
endef

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f api

sh:
	docker compose exec api bash

restart:
	docker compose restart api

test:
	docker compose exec api ./scripts/test.sh

# ---- lint (golangci-lint) ----
# ローカルにGoを入れなくても動くようDockerイメージで実行します。
# 開発中の変更を優先して検査するため、ワークスペースをマウントします。
lint:
	@echo "[lint] running golangci-lint via docker image"
	@docker run --rm -t \
	  -v $(PWD):/work -w /work \
	  -v go_mod_cache:/go/pkg/mod \
	  -v go_build_cache:/root/.cache/go-build \
	  golangci/golangci-lint:latest \
	  golangci-lint run ./...

# ---- buf (proto -> gen) ----
protogen:
	docker compose run --rm -T buf generate

# alias: more discoverable for proto-only codegen
proto: protogen

# ---- scaffold ----
# 例: make scaffold name=User fields="name:string email:string age:int"
SC_NAME := $(if $(strip $(name)),$(strip $(name)),$(word 2,$(MAKECMDGOALS)))
scaffold:
	$(call dev_only)
		@if [ -z "$(strip $(SC_NAME))" ] || [ -z "$(strip $(fields))" ]; then \
			echo "Usage: make scaffold name=Entity fields=\"k1:t1 k2:t2 ...\""; \
			exit 1; \
		fi
		go run ./cmd/scaffold -name "$(SC_NAME)" -fields "$(strip $(fields))" $(if $(mem),-with-memory,)
		$(MAKE) protogen
		go fmt ./...

# 生成物クリーンアップ（同名エンティティを作り直す場合に使用）
# 例: make scaffold-clean name=Article
scaffold-clean:
	$(call dev_only)
	@name='$(name)'; \
    snake=$$(printf "%s" $$name | sed -E 's/([A-Z])/_\1/g' | sed -E 's/^_//' | tr '[:upper:]' '[:lower:]'); \
    rm -rf proto/$$snake gen/$$snake \
      internal/domain/entity/$${snake}.go \
      internal/domain/repository/$${snake}_repository.go \
      internal/usecase/$${snake}_usecase.go \
      internal/adapter/grpc/$${snake}_handler.go \
      internal/adapter/grpc/$${snake}_routes.go \
      internal/adapter/repository/memory/$${snake}_repository.go \
      internal/adapter/repository/mysql/$${snake}_repository.go || true

# ---- clear (make clear <Name>) ----
# 例: make clear Article
# 第二引数（エンティティ名）を拾って scaffold-clean を呼び出します
CLEAR_NAME := $(word 2,$(MAKECMDGOALS))
clear:
	$(call dev_only)
	@if [ -z "$(CLEAR_NAME)" ]; then \
		echo "Usage: make clear <Name>"; exit 1; \
	fi
	@echo "[clear] removing generated files for: $(CLEAR_NAME)"
	@$(MAKE) -s scaffold-clean name="$(CLEAR_NAME)"
		@echo "[clear] pruning db/schema.sql entries for: $(CLEAR_NAME)"
		@name='$(CLEAR_NAME)'; \
		# CamelCase -> snake_case
		snake=$$(printf "%s" $$name | sed -E 's/([A-Z])/_\1/g' | sed -E 's/^_//' | tr '[:upper:]' '[:lower:]'); \
		table=$$snake; [ "$${table##*s}" = "" ] || table=$${table}s; \
		tmp=$$(mktemp); \
		# 1) remove header marker and matching CREATE TABLE block in one pass \
		awk -v NM="$$name" -v TBL="$$table" 'BEGIN{drop=0} \
		  ($$0 ~ "^--[[:space:]]*" NM "[[:space:]]+table[[:space:]]*$$") {next} \
		  ($$0 ~ "^CREATE[[:space:]]+TABLE[[:space:]]+`?" TBL "`?[[:space:]]*\\(") {drop=1; next} \
		  (drop && $$0 ~ /\) ENGINE=/) {drop=0; next} \
		  drop {next} \
		  {print} \
		' db/schema.sql > "$$tmp" \
		&& mv "$$tmp" db/schema.sql && echo "[clear] db/schema.sql updated" || true
		@echo "[clear] (no main.go edits needed; registry-based routes)"
	@if [ "$(drop)" = "1" ]; then \
	  echo "[clear] applying DROP via mysqldef (--enable-drop)"; \
	  $(MAKE) -s migrate DROP_FLAGS="--enable-drop"; \
	else \
	  echo "[clear] (tip) run: make migrate DROP_FLAGS=\"--enable-drop\""; \
	fi

# absorb second arg for `make clear <Name>` so Make doesn't try to build `<Name>`
# Note: keep it conditional to reduce risk of overriding real rules.
ifneq ($(strip $(CLEAR_NAME)),)
$(CLEAR_NAME):
	@:
endif

# フルセット: 起動 -> scaffold -> 生成 -> マイグレーション -> コマンド例出力
scaffold-all:
	$(call dev_only)
	$(MAKE) up
	$(MAKE) scaffold name="$(SC_NAME)" fields="$(strip $(fields))"
	$(MAKE) migrate
	# 再起動して最新コードを反映
	docker compose restart api
	sleep 3
	@echo ""
	@echo "[疎通確認を実行]" && \
	name='$(SC_NAME)'; fields='$(strip $(fields))'; \
	# CamelCase -> snake_case (POSIX tools)
	snake=$$(printf "%s" $$name | sed -E 's/([A-Z])/_\1/g' | sed -E 's/^_//' | tr '[:upper:]' '[:lower:]'); \
	service=$${name}Service; \
	base=$${snake}.v1.$${service}; \
	items=""; comma=""; \
	for p in $$fields; do \
	  key=$${p%%:*}; typ=$${p##*:}; \
	  case "$$typ" in \
	    int|int32|int64) val=123;; \
	    bool) val=true;; \
	    *) val='"sample"';; \
	  esac; \
	  items="$$items$$comma\"$$key\":$$val"; comma=","; \
	done; \
	json="{$$items}"; \
	docker compose run --rm -T curler -sS -H 'Content-Type: application/json' -d "$$json" http://api:8080/$$base/Create$$name | sed -E 's/.*/[Create] &/'; \
	docker compose run --rm -T curler -sS -H 'Content-Type: application/json' -d '{"id":1}' http://api:8080/$$base/Get$$name | sed -E 's/.*/[Get] &/'; \
	docker compose run --rm -T curler -sS -H 'Content-Type: application/json' -d '{}' http://api:8080/$$base/List$${name}s | sed -E 's/.*/[List] &/'
	@echo ""
	@echo "[ローカルで直接叩く例]" && \
	name='$(SC_NAME)'; fields='$(strip $(fields))'; \
	snake=$$(printf "%s" $$name | sed -E 's/([A-Z])/_\1/g' | sed -E 's/^_//' | tr '[:upper:]' '[:lower:]'); \
	service=$${name}Service; \
	base=$${snake}.v1.$${service}; \
	echo "curl -sS -X POST -H 'Content-Type: application/json' --data '{\"id\":1}' http://127.0.0.1:8080/$$base/Get$$name"; \
		echo "curl -sS -X POST -H 'Content-Type: application/json' --data '{}' http://127.0.0.1:8080/$$base/List$${name}s"


# ---- sqldef / mysqldef ----
DB_USER ?= app
DB_PASS ?= apppass

DEV_DB_HOST ?= mysql_dev
DEV_DB_PORT ?= 3306
DEV_DB_NAME ?= app_dev

TEST_DB_HOST ?= mysql_test
TEST_DB_PORT ?= 3306
TEST_DB_NAME ?= app_test

SCHEMA_FILE ?= db/schema.sql

# dev/test 両方に適用
migrate: 
	$(call dev_only)
	$(MAKE) migrate-dev
	$(MAKE) migrate-test

DROP_FLAGS ?=

migrate-dev:
	$(call dev_only)
	docker compose run --rm -T mysqldef \
	  -h $(DEV_DB_HOST) -P $(DEV_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --apply $(DROP_FLAGS) $(DEV_DB_NAME) < $(SCHEMA_FILE)

migrate-test:
	$(call dev_only)
	docker compose run --rm -T mysqldef \
	  -h $(TEST_DB_HOST) -P $(TEST_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --apply $(DROP_FLAGS) $(TEST_DB_NAME) < $(SCHEMA_FILE)

dry-run:
	$(call dev_only)
	$(MAKE) dry-run-dev
	$(MAKE) dry-run-test

dry-run-dev:
	$(call dev_only)
	docker compose run --rm -T mysqldef \
	  -h $(DEV_DB_HOST) -P $(DEV_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --dry-run $(DROP_FLAGS) $(DEV_DB_NAME) < $(SCHEMA_FILE)

dry-run-test:
	$(call dev_only)
	docker compose run --rm -T mysqldef \
	  -h $(TEST_DB_HOST) -P $(TEST_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --dry-run $(DROP_FLAGS) $(TEST_DB_NAME) < $(SCHEMA_FILE)

# テストDBを完全リセット（tmpfsでも、コンテナが生きてる間は状態が残るため）
reset-test-db:
	$(call dev_only)
	docker compose rm -sf mysql_test
	docker compose up -d mysql_test
	$(MAKE) migrate-test

# 開発DB(app_dev)を完全リセット（状態が怪しい時の復旧用）
reset-dev-db:
	$(call dev_only)
	docker compose rm -sf mysql_dev
	docker compose up -d mysql_dev
	$(MAKE) migrate-dev
