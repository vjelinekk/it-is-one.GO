# GEMINI.md - Context for github.com/vjelinekk/it-is-one.GO

## Project Overview
This project is a simple Go-based API server using the `chi` router and `GORM` with a "pure Go" SQLite driver. It follows a clean package structure separating the entry point (`cmd`), the server setup (`pkg/server`), the API handlers (`pkg/api`), and the database models (`pkg/models`).

- **Technologies:** Go 1.26.1, `github.com/go-chi/chi/v5`, `gorm.io/gorm`, `github.com/glebarez/sqlite`
- **Architecture:** 
  - `cmd/server/main.go`: Entry point that initializes the database, runs migrations, and starts the server on port 8080.
  - `pkg/db/db.go`: Database connection and initialization logic.
  - `pkg/models/`: Database model definitions (e.g., `User`).
  - `pkg/server/server.go`: Core server logic, including middleware (Logger, Recoverer, etc.) and routing definitions with DB injection.
  - `pkg/api/health.go`: JSON-based healthcheck endpoint implementation.

## Building and Running
The following commands can be used for development:

- **Local Development:** `bash run.sh` or `go run ./cmd/server/main.go`
- **Build the binary:** `go build -o bin/server ./cmd/server/main.go`
- **Docker Management:** A `docker-manage.sh` script is provided:
  - `bash docker-manage.sh build`: Build the Docker image.
  - `bash docker-manage.sh up`: Start the container in the background.
  - `bash docker-manage.sh down`: Stop and remove the container.
  - `bash docker-manage.sh logs`: View real-time logs.
  - `bash docker-manage.sh status`: Check if the container is running.
- **Docker Compose:** Alternatively, use `docker compose up -d` or `docker compose down`.

## API Documentation
- **Swagger UI:** Once the server is running, visit [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) to view the interactive API documentation and test endpoints.

## Development Conventions
- **Routing:** Uses `go-chi` for routing and middleware.
- **Project Structure:**
  - `cmd/`: Command-line entry points.
  - `pkg/`: Reusable library code.
  - `bin/`: Compiled binaries.
- **Port:** Defaults to `:8080`.
- **Healthcheck:** Available at `/health` returning JSON with status and timestamp.
- **Main Entry:** Available at `/` returning a plain text welcome message.
