# Functional Specification: Smart Pill Doser API (POC 2026)

## 1. Core Objective
The API acts as the central coordinator for the Smart Pill Doser ecosystem. It is responsible for tracking medication adherence, monitoring hardware health, and executing a persistent, escalating notification system to ensure patients take their medication and caregivers are alerted to failures.

---

## 2. Communication & Connectivity Logic
The system follows a strict hierarchical communication model to ensure data delivery to the API under varying environmental conditions.

### 2.1 Dual-Path Connectivity (Exclusive)
1.  **Primary Path (Bluetooth Low Energy):** The ESP32 attempts to pair with the user's mobile phone. If successful, the mobile app acts as the internet gateway for the API.
2.  **Fallback Path (Wi-Fi):** If a Bluetooth connection is unavailable, the ESP32 activates its Wi-Fi module to connect directly to the API via a local network.
3.  **Conflict Prevention:** The hardware uses only one path at a time. If BLE is active, Wi-Fi is disabled to prevent duplicate data streams.

---

## 3. API Functional Responsibilities

### 3.1 Identity & Hardware Mapping
* **User-to-Device Linking:** The API maintains a 1-to-1 relationship between a `User ID` and a physical `Device Serial Number`.
* **Caretaker Management:** The API stores a list of caretaker contact details (Email/Push Tokens) associated with each patient.
* **POC Authentication:** Requests are identified via custom headers: `X-User-ID` for the mobile app and `X-Device-Serial` for the hardware.

### 3.2 Schedule & Intake Management
* **Schedule Storage:** Users define pill times (e.g., 08:00 AM) and frequency (e.g., Daily) via the API.
* **Intake Logging:** When the hardware detects a pill room is emptied, it sends a signal. The API marks the corresponding scheduled dose as **"Taken"** and records the `actual_intake_time`.

### 3.3 Device Health Monitoring ("I'm Alive" Heartbeat)
* **Periodic Heartbeat:** The hardware sends a status ping every **60 minutes**.
* **Offline Watchdog:** The API monitors the time since the last heartbeat.
* **Alert Trigger:** If no signal (Heartbeat or Pill Taken) is received for **120 minutes (2 hours)**, the API notifies the **User only**: *"Your pill doser is offline. Please check the battery or Wi-Fi."*

---

## 4. Escalation Notification Logic (The "Nudge")
If a pill is not marked as "Taken" by the scheduled time ($T$), the API initiates a reminder cycle. The behavior depends on the user's `missed_dose_threshold`.

### 4.1 Condition: Below Threshold
If `current_missed_doses < missed_dose_threshold`:
- **$T$ + 0 min**: Initial "Time for your medication" alert to **User Only**.
- **Every 10 mins**: Reminder to **User Only** until "Taken" signal is received.

### 4.2 Condition: At or Above Threshold
If `current_missed_doses >= missed_dose_threshold`:
- **$T$ + 0 min**: Immediate "Medication Due" alert to **User AND Caregivers**.
- **Every 10 mins**: Persistent reminders to **User AND Caregivers** until "Taken" signal is received.

---

## 5. Missed Dose Definition
A "Missed Dose" is recorded if a scheduled dose is not taken by the time of the *next* scheduled dose, or at the end of the day (as defined by the system). When a dose is missed:
1. `current_missed_doses` is incremented.
2. If the user successfully takes a dose later, `current_missed_doses` is reset to 0.