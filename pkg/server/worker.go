package server

import (
	"log"
	"time"

	"github.com/vjelinekk/it-is-one.GO/pkg/email"
	"github.com/vjelinekk/it-is-one.GO/pkg/models"
	"gorm.io/gorm"
)

// StartEscalator runs every 1 minute to check for missed doses
func StartEscalator(db *gorm.DB) {
	log.Println("Starting Escalator worker...")
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			checkSchedules(db)
		}
	}()
}

// StartWatchdog runs every 5 minutes to check for offline devices
func StartWatchdog(db *gorm.DB) {
	log.Println("Starting Watchdog worker...")
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			checkOfflineDevices(db)
		}
	}()
}

func checkSchedules(db *gorm.DB) {
	var users []models.User
	if err := db.Preload("Schedules").Preload("Caregivers").Find(&users).Error; err != nil {
		log.Printf("Worker Error: Failed to fetch users: %v", err)
		return
	}

	log.Printf("[WORKER] checking %d users", len(users))

	for _, user := range users {
		loc, err := time.LoadLocation(user.Timezone)
		if err != nil || user.Timezone == "" {
			loc = time.UTC
		}
		now := time.Now().In(loc)
		startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

		// Check if dose was taken today
		var takenCount int64
		db.Model(&models.IntakeLog{}).
			Where("user_id = ? AND taken_at >= ?", user.ID, startOfDay).
			Count(&takenCount)

		if takenCount > 0 {
			// Dose taken today — reset missed doses if needed
			if user.CurrentMissedDoses > 0 {
				db.Model(&models.User{}).Where("id = ?", user.ID).Update("current_missed_doses", 0)
			}
			continue
		}

		log.Printf("[WORKER] user %d has %d schedules, now=%s", user.ID, len(user.Schedules), now.Format("15:04:05"))

		for _, schedule := range user.Schedules {
			for _, scheduledTimeStr := range []string{schedule.ScheduledTime} {
				if scheduledTimeStr == "" {
					continue
				}

				log.Printf("[WORKER] user %d schedule time raw: %q", user.ID, scheduledTimeStr)

				parsed, err := time.Parse("15:04:05", scheduledTimeStr)
				if err != nil {
					parsed, err = time.Parse("15:04", scheduledTimeStr)
				}
				if err != nil {
					log.Printf("[WORKER] user %d parse error for %q: %v", user.ID, scheduledTimeStr, err)
					continue
				}

				scheduledTime := time.Date(now.Year(), now.Month(), now.Day(),
					parsed.Hour(), parsed.Minute(), parsed.Second(), 0, loc)

				minutesSince := int(now.Sub(scheduledTime).Minutes())

				// Skip if not due yet or too old (> 2 hours)
				if minutesSince < 0 || minutesSince > 120 {
					continue
				}

				// Only trigger at each NotifyAfterMinutes interval
				notifyInterval := user.NotifyAfterMinutes
				if notifyInterval <= 0 {
					notifyInterval = models.DefaultNotifyAfterMinutes
				}
				if minutesSince%notifyInterval != 0 {
					continue
				}

				// Increment missed doses
				db.Model(&models.User{}).Where("id = ?", user.ID).
					Update("current_missed_doses", gorm.Expr("current_missed_doses + 1"))

				// Re-fetch updated count
				var updatedUser models.User
				db.Select("current_missed_doses, notify_caregivers_after_retries").
					First(&updatedUser, user.ID)

				if updatedUser.CurrentMissedDoses >= updatedUser.NotifyCaregiversAfterRetries {
					if len(user.Caregivers) == 0 {
						log.Printf("[EMAIL] User %d: threshold met at %s but no caregivers registered",
							user.ID, scheduledTimeStr)
					} else {
						for _, cg := range user.Caregivers {
							email.SendMissedDoseAlert(cg.Email, user.Email, scheduledTimeStr)
						}
					}
				} else {
					log.Printf("[NUDGE] User %d: dose missed at %s, retry %d/%d",
						user.ID, scheduledTimeStr, updatedUser.CurrentMissedDoses, updatedUser.NotifyCaregiversAfterRetries)
				}
			}
		}
	}
}

func checkOfflineDevices(db *gorm.DB) {
	threshold := time.Now().Add(-120 * time.Minute)

	var offlineUsers []models.User
	db.Where("device_last_seen < ? AND device_serial IS NOT NULL", threshold).Find(&offlineUsers)

	for _, user := range offlineUsers {
		log.Printf("[WATCHDOG] User %d: device is OFFLINE (Last seen: %v)",
			user.ID, user.DeviceLastSeen)
	}
}
