.PHONY: up down logs sh test \
        generate scaffold scaffold-all \
        migrate migrate-dev migrate-test \
        dry-run dry-run-dev dry-run-test \
        reset-test-db

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f api

sh:
	docker compose exec api bash

test:
	docker compose exec api ./scripts/test.sh

# ---- buf (proto -> gen) ----
generate:
	docker compose run --rm -T buf generate

# ---- scaffold ----
# 例: make scaffold name=User fields="name:string email:string age:int"
scaffold:
	go run ./cmd/scaffold -name "$(name)" -fields "$(fields)"
	$(MAKE) generate
	go fmt ./...

# 生成物クリーンアップ（同名エンティティを作り直す場合に使用）
# 例: make scaffold-clean name=Article
scaffold-clean:
	@name='$(name)'; \
	snake=$$(printf "%s" $$name | sed -E 's/([A-Z])/_\1/g' | sed -E 's/^_//' | tr '[:upper:]' '[:lower:]'); \
	rm -rf proto/$$snake gen/$$snake \
	  internal/usecase/$${snake}_usecase.go \
	  internal/adapter/grpc/$${snake}_handler.go \
	  internal/adapter/repository/memory/$${snake}_repository.go \
	  internal/adapter/repository/mysql/$${snake}_repository.go || true

# フルセット: 起動 -> scaffold -> 生成 -> マイグレーション -> コマンド例出力
scaffold-all:
	$(MAKE) up
	$(MAKE) scaffold name="$(name)" fields="$(fields)"
	$(MAKE) migrate
	# 再起動して最新コードを反映
	docker compose restart api
	sleep 1
	@echo ""
	@echo "[疎通確認を実行]" && \
	name='$(name)'; fields='$(fields)'; \
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
	name='$(name)'; fields='$(fields)'; \
	snake=$$(printf "%s" $$name | sed -E 's/([A-Z])/_\1/g' | sed -E 's/^_//' | tr '[:upper:]' '[:lower:]'); \
	service=$${name}Service; \
	base=$${snake}.v1.$${service}; \
	echo "curl -sS -X POST -H 'Content-Type: application/json' --data '{\"id\":1}' http://127.0.0.1:8080/$$base/Get$$name"; \
	echo "curl -sS -X POST -H 'Content-Type: application/json' --data '{}' http://127.0.0.1:8080/$$base/List$$name\"s"


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
migrate: migrate-dev migrate-test

DROP_FLAGS ?=

migrate-dev:
	docker compose run --rm -T mysqldef \
	  -h $(DEV_DB_HOST) -P $(DEV_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --apply $(DROP_FLAGS) $(DEV_DB_NAME) < $(SCHEMA_FILE)

migrate-test:
	docker compose run --rm -T mysqldef \
	  -h $(TEST_DB_HOST) -P $(TEST_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --apply $(DROP_FLAGS) $(TEST_DB_NAME) < $(SCHEMA_FILE)

dry-run: dry-run-dev dry-run-test

dry-run-dev:
	docker compose run --rm -T mysqldef \
	  -h $(DEV_DB_HOST) -P $(DEV_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --dry-run $(DROP_FLAGS) $(DEV_DB_NAME) < $(SCHEMA_FILE)

dry-run-test:
	docker compose run --rm -T mysqldef \
	  -h $(TEST_DB_HOST) -P $(TEST_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --dry-run $(DROP_FLAGS) $(TEST_DB_NAME) < $(SCHEMA_FILE)

# テストDBを完全リセット（tmpfsでも、コンテナが生きてる間は状態が残るため）
reset-test-db:
	docker compose rm -sf mysql_test
	docker compose up -d mysql_test
	$(MAKE) migrate-test
