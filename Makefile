run:
	go run ./cmd/server/main.go

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