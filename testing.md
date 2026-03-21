# API Testing Guide: Smart Pill Doser (POC)

This document provides a set of manual testing scenarios using `curl` to verify the functionality of the Smart Pill Doser API.

## Prerequisites
- The server must be running: `bash run.sh` or `go run ./cmd/server/main.go`
- Base URL: `http://localhost:8080/api/v1`

---

## 1. Authentication Headers
The API uses two types of custom headers for identification:
- **Mobile/Client:** `X-User-ID: <id>` (e.g., `X-User-ID: 1`)
- **Hardware (ESP32):** `X-Device-Serial: <serial>` (e.g., `X-Device-Serial: ESP32-ABC-123`)

---

## 2. Scenario 1: Initial User Registration (Public)
**Goal:** Create a new user in the system.

```bash
curl -X POST http://localhost:8080/api/v1/users \
     -H "Content-Type: application/json" \
     -d '{
       "full_name": "John Doe",
       "email": "john@example.com",
       "timezone": "Europe/Prague"
     }'
```
*Take note of the `"id"` in the response (e.g., `1`).*

---

## 3. Scenario 2: Mobile App Configuration (Authenticated)
**Goal:** Set up the user's environment using the `X-User-ID` header.

### 3.1 Link a Hardware Device
```bash
curl -X PUT http://localhost:8080/api/v1/users/me/device \
     -H "X-User-ID: 1" \
     -H "Content-Type: application/json" \
     -d '{"device_serial": "ESP32-ABC-123"}'
```

### 3.2 Create a Medication Schedule
*Set the time to 2 minutes from your current local time to test the Escalator worker.*
```bash
curl -X POST http://localhost:8080/api/v1/schedules \
     -H "X-User-ID: 1" \
     -H "Content-Type: application/json" \
     -d '{
       "scheduled_time": "14:30:00",
       "days_of_week": "1,2,3,4,5,6,7"
     }'
```

### 3.3 Add a Caregiver
```bash
curl -X POST http://localhost:8080/api/v1/caregivers \
     -H "X-User-ID: 1" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Jane Caregiver",
       "email": "jane@example.com"
     }'
```

### 3.4 Update User Threshold
```bash
curl -X PUT http://localhost:8080/api/v1/users/me \
     -H "X-User-ID: 1" \
     -H "Content-Type: application/json" \
     -d '{"missed_dose_threshold": 2}'
```

---

## 4. Scenario 3: Hardware Simulation (Authenticated)
**Goal:** Simulate ESP32 behavior using the `X-Device-Serial` header.

### 4.1 Device Heartbeat (I'm Alive)
```bash
curl -X POST http://localhost:8080/api/v1/device/heartbeat \
     -H "X-Device-Serial: ESP32-ABC-123" \
     -H "Content-Type: application/json" \
     -d '{"battery_level": 88}'
```

### 4.2 Log a Successful Intake
```bash
curl -X POST http://localhost:8080/api/v1/device/intake \
     -H "X-Device-Serial: ESP32-ABC-123" \
     -H "Content-Type: application/json" \
     -d '{
       "timestamp": "2026-03-21T14:30:00Z",
       "status": "taken"
     }'
```
*Note: This resets `current_missed_doses` to 0.*

---

## 5. Scenario 4: Background Worker Logic (Verification)

### 5.1 The Escalator (Missed Dose Notification)
1. Set a schedule for **1 minute from now**.
2. Do **not** send the `intake` signal.
3. Check the server logs. You should see a `[NUDGE]` log entry.

### 5.2 Threshold Escalation (Caregiver Alert)
1. Manually set `current_missed_doses` to `>= missed_dose_threshold` in the database (or via API).
2. Set a schedule for **1 minute from now**.
3. Check the server logs. You should see a `[NOTIFICATION]` log entry for the caregiver.

### 5.3 The Watchdog (Offline Alert)
1. Stop sending heartbeats for **120+ minutes** (or manually update `device_last_seen` in the DB).
2. Watch the server logs for the `[WATCHDOG]` offline alert.
