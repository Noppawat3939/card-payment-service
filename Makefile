include .env.local
export


# path file
MAIN_PATH=./cmd/server/main.go
MIGRATION_PATH=./migrations

# database
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable


run:
	go run $(MAIN_PATH)

dev:
	air

tidy:
	go mod tidy

compose-up:
	docker compose --env-file .env.local up -d

compose-down:
	docker compose down

compose-logs:
	docker compose logs -f

db-reset:
	docker compose down -v
	docker compose up -d

migrate-up-one:
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" up 1

migrate-down-one:
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" down 1

migrate-version:
	migrate -path $(MIGRATION_PATH) -database "$(DB_URL)" version
