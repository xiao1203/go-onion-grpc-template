.PHONY: up down logs sh test \
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

migrate-dev:
	docker compose run --rm -T mysqldef \
	  -h $(DEV_DB_HOST) -P $(DEV_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --apply $(DEV_DB_NAME) < $(SCHEMA_FILE)

migrate-test:
	docker compose run --rm -T mysqldef \
	  -h $(TEST_DB_HOST) -P $(TEST_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --apply $(TEST_DB_NAME) < $(SCHEMA_FILE)

dry-run: dry-run-dev dry-run-test

dry-run-dev:
	docker compose run --rm -T mysqldef \
	  -h $(DEV_DB_HOST) -P $(DEV_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --dry-run $(DEV_DB_NAME) < $(SCHEMA_FILE)

dry-run-test:
	docker compose run --rm -T mysqldef \
	  -h $(TEST_DB_HOST) -P $(TEST_DB_PORT) -u$(DB_USER) -p$(DB_PASS) \
	  --dry-run $(TEST_DB_NAME) < $(SCHEMA_FILE)

# テストDBを完全リセット（tmpfsでも、コンテナが生きてる間は状態が残るため）
reset-test-db:
	docker compose rm -sf mysql_test
	docker compose up -d mysql_test
	$(MAKE) migrate-test
