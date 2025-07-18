# ---------- [ CONFIGURATION ] ----------

APP_NAME=subtracker
DOCKER_COMPOSE=docker compose

include .env
export

# ---------- [ DB CONFIGURATION ] ----------

MIGRATIONS_PATH=./migrations

# Используем переменные из .env напрямую
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable

# ---------- [ LOCAL GO ] ----------

build:
	go build -o $(APP_NAME) ./cmd

run: build
	./$(APP_NAME)

tidy:
	go mod tidy

test:
	go test ./...

# ---------- [ DOCKER ] ----------

up:
	$(DOCKER_COMPOSE) up --build

down:
	$(DOCKER_COMPOSE) down

restart: down up

logs:
	$(DOCKER_COMPOSE) logs -f

ps:
	$(DOCKER_COMPOSE) ps

# ---------- [ MIGRATIONS ] ----------

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down

migrate-force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force ${version}

migrate-drop:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" drop -f

migrate-goto:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" goto ${version}

migrate-version:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

# ---------- [ UTILITY ] ----------

help:
	@echo "Available commands:"
	@echo "  build           - Build the Go application"
	@echo "  run             - Run the Go application"
	@echo "  tidy            - Tidy up Go modules"
	@echo "  test            - Run tests"
	@echo "  up              - Start Docker containers"
	@echo "  down            - Stop Docker containers"
	@echo "  restart         - Restart Docker containers"
	@echo "  logs            - Show Docker logs"
	@echo "  ps              - List Docker containers"
	@echo "  migrate-up      - Apply database migrations"
	@echo "  migrate-down    - Rollback database migrations"
	@echo "  migrate-force   - Force a migration version"
	@echo "  migrate-drop    - Drop the database schema"
	@echo "  migrate-goto    - Go to a specific migration version"
	@echo "  migrate-version - Show current migration version"
	@echo "  help            - Show this help message"
