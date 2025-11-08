# 📦 Subscription Tracker API

**A REST service for aggregating data about users' online subscriptions, implemented in Go.**  
This project is a solution for the Junior Golang Developer technical assignment.

---

## ✨ Features

- ✅ Full CRUDL API for managing user subscriptions  
- 💰 Cost Calculation Endpoint with flexible filtering  
- 🛡️ Advanced Request Validation using `go-playground/validator`  
- 🧱 Clean Architecture (Handler, Service, Repository)  
- 🐳 Single-Command Startup with Docker and Docker Compose  
- 📘 Interactive Swagger API Documentation  
- 📄 Structured Logging with `uber-go/zap`  
- 🔁 Graceful Shutdown support  

---

## 🚀 Quick Start

> **Requirements:**  
> Docker + Docker Compose installed

### 1. Clone the repository

```bash
git clone https://github.com/adal4ik/subtracker.git
cd subtracker
```
### 2. Create the .env file
Copy the example environment file. The default values are already configured for local launch.
```bash
cp .env.example .env
```
### 3. Start the service
This command will build the application's Docker image, start all required containers (app, database, Swagger UI), and automatically apply database migrations.
```bash
make up
```
### 4. Stop the service
To stop and remove all running containers for this project:
```bash
make down
```
The service is now running and ready for use!

    API Base URL: http://localhost:8080

    Interactive Swagger UI: http://localhost:8081

## API Documentation

The interactive Swagger documentation is the best way to explore and test the API.

Simply open the following URL in your browser:
http://localhost:8081

The Swagger UI is served in a separate container and is pre-configured to display the documentation for this API.

## 🧪 Tests

To run all unit tests for the project, execute the following command:
```bash
make test
```
Or, if you do not have make installed:
```bash
go test ./... -v
```
## Makefile Commands

A Makefile is included to simplify common development tasks. To see all available commands and their descriptions, run:
```bash
make help
```

## Project structure
```
.
├── cmd/app/                # Application entry point (main.go)
├── internal/
│   ├── config/             # Configuration loading (`cleanenv`)
│   ├── domain/             # Core domain models (DTOs, DAOs)
│   ├── handler/            # HTTP handlers (Controllers)
│   ├── mapper/             # Data mapping functions
│   ├── repository/         # Database access layer (`sqlx`, `squirrel`)
│   └── service/            # Business logic layer
├── migrations/             # SQL database migrations (`golang-migrate`)
├── pkg/                    # Reusable packages (apperrors, logger, validator)
├── docs/                   # Generated Swagger documentation
├── Dockerfile              # Multi-stage Dockerfile for building the app
├── docker-compose.yml      # Docker Compose configuration for all services
├── Makefile                # Automation for common tasks
└── README.md
```

---

## ✨ Features

✅ **Implemented**
- CRUD operations for subscriptions  
- Filtering, pagination, and status tracking  
- PostgreSQL + Clean Architecture  
- Admin routes for service management  

🚧 **In Progress**
- Renewal scheduler (auto-renewal of subscriptions)  
- CSV import/export support  

🔮 **Planned**
- JWT authentication for users  
- Email notifications for renewals  
- Usage tracking and statistics dashboard
---

**Author:**  
adal4ik
