# API Design: Smart Pill Doser (POC)

## Authentication (POC Level)
- **Mobile/Client Requests:** Authenticated via the header `X-User-ID: <user_id>`.
- **Hardware (ESP32) Requests:** Authenticated via the header `X-Device-Serial: <serial_number>`.

---

## 1. Hardware (ESP32) API

These endpoints are called by the ESP32 (either via Wi-Fi directly, or proxied through the Mobile App via BLE).

### 1.1 Device Heartbeat
**`POST /api/v1/device/heartbeat`**
- **Description:** Sends an "I'm alive" signal. Updates the `device_last_seen` timestamp and `device_battery` level in the linked User record. Must be called every 60 minutes.
- **Headers:** `X-Device-Serial: ESP32-XYZ`
- **Request Body:**
  ```json
  {
    "battery_level": 85
  }
  ```
- **Response:** `200 OK`

### 1.2 Log Intake
**`POST /api/v1/device/intake`**
- **Description:** Called when the hardware detects a pill room is emptied. This creates an `intake_logs` entry and should reset `current_missed_doses` to 0.
- **Headers:** `X-Device-Serial: ESP32-XYZ`
- **Request Body:**
  ```json
  {
    "timestamp": "2026-03-20T08:05:00Z", 
    "status": "taken" 
  }
  ```
- **Response:** `201 Created`

---

## 2. Mobile App (Client) API

### 2.1 User & Device Linking
**`GET /api/v1/users/me`**
- **Description:** Get user profile, threshold settings, and linked device status.

**`PUT /api/v1/users/me`**
- **Description:** Update user profile (e.g., `full_name`, `timezone`, `missed_dose_threshold`).

**`PUT /api/v1/users/me/device`**
- **Description:** Link an ESP32 device to the user by its hardcoded serial number.

### 2.2 Schedule Management
**`POST /api/v1/schedules`**
- **Description:** Create a new medication schedule (e.g. 08:00 AM).

**`GET /api/v1/schedules`**
- **Description:** List all schedules for the user.

**`DELETE /api/v1/schedules/{id}`**
- **Description:** Delete a schedule.

### 2.3 Caregiver Management
**`POST /api/v1/caregivers`**
- **Description:** Add a caregiver to be notified when the threshold is met.

**`GET /api/v1/caregivers`**
- **Description:** List all caregivers for the user.

**`DELETE /api/v1/caregivers/{id}`**
- **Description:** Remove a caregiver.

### 2.4 Intake Logs (History)
**`GET /api/v1/intake-logs`**
- **Description:** Retrieve the history of taken/missed medications.

---

## 3. Background Workers
1. **The Watchdog:** Runs every 5 minutes. Checks `users.device_last_seen`. If `now() - device_last_seen > 120 minutes`, it triggers an email notification to the user: *"Your pill doser is offline."*
2. **The Escalator:** Runs every 1 minute. Checks `schedules` against current time and `intake_logs`. 
    - At $T$ (Scheduled Time): If `current_missed_doses < threshold`, notify User via email. 
    - At $T$ (Scheduled Time): If `current_missed_doses >= threshold`, notify User + Caregivers via email.
    - If a dose is missed (T + X hours), it increments `current_missed_doses`.
