DB_URL=postgres://postgres:postgres@localhost:5432/shortener?sslmode=disable

.PHONY: run db-up migrate-up migrate-down create-migration setup

setup:
	@which migrate > /dev/null || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

db-up:
	docker compose up -d postgres

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

create-migration:
	@if [ -z "$(name)" ]; then echo "Usage: make create-migration name=my_migration"; exit 1; fi
	migrate create -ext sql -dir db/migrations -seq $(name)

run: db-up
	@sleep 3
	@make migrate-up
	go run ./cmd/api