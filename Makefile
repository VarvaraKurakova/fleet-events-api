POSTGRES_DSN=postgres://fleet:fleet@localhost:5432/fleet_events?sslmode=disable

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

migrate-up:
	migrate -path migrations -database "$(POSTGRES_DSN)" up

migrate-down:
	migrate -path migrations -database "$(POSTGRES_DSN)" down 1

migrate-force:
	migrate -path migrations -database "$(POSTGRES_DSN)" force $(VERSION)

test:
	go test ./...

test-race:
	go test -race ./...

tidy:
	go mod tidy
