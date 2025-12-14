.PHONY: up down logs sh test

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
