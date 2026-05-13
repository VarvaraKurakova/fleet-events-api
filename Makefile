run-api:
	go run ./cmd/api

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

test:
	go test ./...

test-race:
	go test -race ./...

tidy:
	go mod tidy
