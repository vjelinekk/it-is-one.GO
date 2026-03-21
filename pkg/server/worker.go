package server

import (
	"log"
	"sort"
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

	for _, user := range users {
		loc, err := time.LoadLocation(user.Timezone)
		if err != nil || user.Timezone == "" {
			loc = time.UTC
		}
		now := time.Now().In(loc)
		today := now.Format("2006-01-02")

		// Sort schedules by time so index 0 = dose 1, index 1 = dose 2
		sort.Slice(user.Schedules, func(i, j int) bool {
			return user.Schedules[i].ScheduledTime < user.Schedules[j].ScheduledTime
		})

		for i, schedule := range user.Schedules {
			doseSlot := i + 1
			scheduledTimeStr := schedule.ScheduledTime
			if scheduledTimeStr == "" {
				continue
			}

			// Check if this specific dose was already taken today
			var takenCount int64
			db.Model(&models.IntakeLog{}).
				Where("user_id = ? AND dose_slot = ? AND date = ?", user.ID, doseSlot, today).
				Count(&takenCount)
			if takenCount > 0 {
				if user.CurrentMissedDoses > 0 {
					db.Model(&models.User{}).Where("id = ?", user.ID).Update("current_missed_doses", 0)
				}
				continue
			}

			parsed, err := time.Parse("15:04:05", scheduledTimeStr)
			if err != nil {
				parsed, err = time.Parse("15:04", scheduledTimeStr)
			}
			if err != nil {
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

			threshold := updatedUser.NotifyCaregiversAfterRetries
			if updatedUser.CurrentMissedDoses >= threshold && updatedUser.CurrentMissedDoses <= threshold+1 {
				if len(user.Caregivers) == 0 {
					log.Printf("[EMAIL] User %d: dose %d threshold met at %s but no caregivers registered",
						user.ID, doseSlot, scheduledTimeStr)
				} else {
					for _, cg := range user.Caregivers {
						email.SendMissedDoseAlert(cg.Email, user.Email, scheduledTimeStr)
					}
				}
			} else if updatedUser.CurrentMissedDoses < threshold {
				log.Printf("[NUDGE] User %d: dose %d missed at %s, retry %d/%d",
					user.ID, doseSlot, scheduledTimeStr, updatedUser.CurrentMissedDoses, threshold)
			}
		}
	}
}

func checkOfflineDevices(db *gorm.DB) {
	threshold := time.Now().Add(-60 * time.Minute)

	var offlineUsers []models.User
	db.Preload("Caregivers").
		Where("device_last_seen < ? AND device_serial IS NOT NULL AND device_offline_notified = false", threshold).
		Find(&offlineUsers)

	for _, user := range offlineUsers {
		lastSeen := "unknown"
		if user.DeviceLastSeen != nil {
			lastSeen = user.DeviceLastSeen.Format("2006-01-02 15:04:05")
		}
		log.Printf("[WATCHDOG] User %d: device OFFLINE since %s", user.ID, lastSeen)

		if len(user.Caregivers) == 0 {
			log.Printf("[WATCHDOG] User %d: no caregivers to notify", user.ID)
		} else {
			for _, cg := range user.Caregivers {
				email.SendDeviceOfflineAlert(cg.Email, user.Email, lastSeen)
			}
		}

		db.Model(&models.User{}).Where("id = ?", user.ID).Update("device_offline_notified", true)
	}
}
