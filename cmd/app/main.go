package main

import (
	"PostgresDump/internal/config"
	"PostgresDump/internal/processor"
	healthcheck "PostgresDump/internal/services/healthCheck"
	v "github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.Init()
	cfg.Log.Info("🚀 Starting backup script...")

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	if _, err := os.Stat(cfg.Postgres.BackupPath); os.IsNotExist(err) {
		err := os.Mkdir(cfg.Postgres.BackupPath, 0755)
		if err != nil {
			log.Fatalf("❌ Error creating backup folder: %v", err)
		}
	}

	if v.GetBool("health_check") {
		err := healthcheck.HealthCheck(cfg)
		if err != nil {
			log.Fatalf("❌ Error checking service health: %v", err)
		}
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				cfg.Log.Error("⚠️ Critical error in backup process", "error", r)
			}
		}()
		processor.Run(cfg)
	}()

	<-stopChan

	cfg.Log.Info("🛑 Shutting down...")
}
