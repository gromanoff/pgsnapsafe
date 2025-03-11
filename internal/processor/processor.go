package processor

import (
	"PostgresDump/internal/config"
	"PostgresDump/internal/services/backups"
	"PostgresDump/pkg/email"
	v "github.com/spf13/viper"
	"time"
)

func Run(cfg *config.Config) {
	cfg.Log.Info("🔄 Starting backup cycle...")
	cfg.Log.Info("📋 Backup schedule", "times", cfg.Backup.Times)

	lastRun := make(map[string]time.Time) // Храним последнее время выполнения бэкапа

	for {
		now := time.Now()

		for _, t := range cfg.Backup.Times {
			backupTimeParsed, err := time.Parse("15:04", t)
			if err != nil {
				cfg.Log.Error("Error parsing time", "time", t, "error", err)
				continue
			}

			backupTime := time.Date(now.Year(), now.Month(), now.Day(),
				backupTimeParsed.Hour(), backupTimeParsed.Minute(), 0, 0, now.Location())

			// Проверяем, запускался ли уже бэкап сегодня
			if lastRunTime, exists := lastRun[t]; exists {
				if lastRunTime.Day() == now.Day() {
					continue // Бэкап уже запускался сегодня, пропускаем
				}
			}

			if abs(now.Sub(backupTime).Seconds()) < 30 {
				cfg.Log.Info("🕒 Backup time!", "time", t)

				filePath, err := backups.CreateBackup(cfg)
				if err != nil {
					cfg.Log.Error("❌ Error creating backup", "error", err)
				} else {
					cfg.Log.Info("✅ Backup created successfully", "file", filePath)
					lastRun[t] = now // Запоминаем, что бэкап уже был выполнен
				}

				if err := backups.CleanupOldBackups(cfg); err != nil {
					cfg.Log.Error("🚨 Error cleaning up old backups", "error", err)
				} else {
					cfg.Log.Info("🧹 Old backups cleanup completed")
				}

				if v.GetBool("smtp") {
					if err = email.SendEmail(cfg.SMTPClient, v.GetString("email_delivery"), filePath); err != nil {
						cfg.Log.Error("Error sending email", "error", err)
					}
				}
			}
		}

		time.Sleep(60 * time.Second) // Проверяем каждую минуту
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
