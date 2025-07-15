# Variables
APP_NAME = subtracker
DOCKER_COMPOSE = docker compose

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
	migrate -path ./migrations -database "$$(grep POSTGRES_DSN .env | cut -d '=' -f2-)" up

migrate-down:
	migrate -path ./migrations -database "$$(grep POSTGRES_DSN .env | cut -d '=' -f2-)" down

migrate-force:
	migrate -path ./migrations -database "$$(grep POSTGRES_DSN .env | cut -d '=' -f2-)" force ${version}

migrate-drop:
	migrate -path ./migrations -database "$$(grep POSTGRES_DSN .env | cut -d '=' -f2-)" drop -f

migrate-goto:
	migrate -path ./migrations -database "$$(grep POSTGRES_DSN .env | cut -d '=' -f2-)" goto ${version}

migrate-version:
	migrate -path ./migrations -database "$$(grep POSTGRES_DSN .env | cut -d '=' -f2-)" version
