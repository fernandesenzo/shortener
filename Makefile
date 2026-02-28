ifneq (,$(wildcard ./.env))
    include .env
    export
endif

APP_NAME=shortener_app
DB_CONTAINER=shortener_db
CACHE_CONTAINER=shortener_cache
COMPOSE_FILE=docker-compose.dev.yml

.PHONY: run up down restart logs redis-keys db-keys

run: up
	@echo "⌛ waitng for  postgres at localhost:5432..."
	@until nc -z localhost 5432; do printf '.'; sleep 1; done
	go run ./cmd/api

up:
	docker compose -f $(COMPOSE_FILE) up -d

down:
	docker compose -f $(COMPOSE_FILE) down

restart:
	docker compose -f $(COMPOSE_FILE) up -d --build

redis-keys:
	docker exec -it $(CACHE_CONTAINER) redis-cli -a $(REDIS_PASSWORD) KEYS "*"

db-keys:
	docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d shortener -c "SELECT * FROM links;"

logs:
	docker compose -f $(COMPOSE_FILE) logs -f