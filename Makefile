include .env
export

MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDRESS) up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDRESS) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-version
migrate-version:
	@migrate -path=$(MIGRATIONS_PATH) -database=$(DB_ADDRESS) version

.PHONY: docker-up docker-down

docker-up:
	@docker compose up -d

docker-down:
	@docker compose down

.PHONY: seed
seed: 
	@go run cmd/migrate/seed/main.go