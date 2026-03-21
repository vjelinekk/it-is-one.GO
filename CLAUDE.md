# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the server locally
go run ./cmd/server/main.go
# or
./run.sh

# Build the binary
go build -o bin/server ./cmd/server/main.go

# Docker
docker-compose up --build
docker-compose down

# Swagger docs are generated with swaggo — regenerate after changing doc comments:
~/go/bin/swag init -g cmd/server/main.go -o docs
```

There are no automated tests in this codebase. `testing.md` contains manual curl-based test cases.

## Architecture

The API is a medication adherence backend for a Smart Pill Doser device (ESP32 hardware + mobile app).

**Layer structure:**
- `cmd/server/main.go` — entry point: init DB, run migrations, start server + workers
- `pkg/server/server.go` — chi router, middleware groups, route registration
- `pkg/api/` — HTTP handlers split by domain (`user.go`, `mobile.go`, `hardware.go`, `health.go`, `middleware.go`)
- `pkg/models/` — GORM models with SQLite persistence (`data.db`)
- `pkg/db/db.go` — DB initialization using pure-Go SQLite driver
- `pkg/server/worker.go` — two background goroutines

**Authentication** is header-based (no tokens/JWT):
- Mobile endpoints: `X-User-ID` header (parsed as uint, injected into context via `MobileAuthMiddleware`)
- Hardware endpoints: `X-Device-Serial` header (injected via `HardwareAuthMiddleware`)
- `POST /users` is the only public endpoint

**Background workers** in `pkg/server/worker.go`:
- **Escalator** (every 1 min): Checks all users' schedules; if a dose is overdue and not taken, logs `[NUDGE]` for user and `[NOTIFICATION]` for caregivers if `CurrentMissedDoses >= MissedDoseThreshold`. This is a POC — notifications are only `log.Printf` calls, no real push notifications.
- **Watchdog** (every 5 min): Logs `[WATCHDOG]` for devices not seen in 120+ minutes.

**Key domain logic:**
- A `User` has `DeviceSerial` (1-to-1 link to an ESP32), `MissedDoseThreshold`, and `CurrentMissedDoses`
- `Schedule` stores `ScheduledTime` (HH:MM:SS) and `DaysOfWeek` (comma-separated ints)
- When hardware calls `POST /device/intake`, `CurrentMissedDoses` resets to 0
- `IntakeLog` records both `PlannedAt` and `ActualAt` timestamps with status `"taken"` or `"missed"`
- Caregivers are linked to a patient via `PatientID` and receive escalation notifications above threshold
