package backups

import (
	"PostgresDump/internal/config"
	"PostgresDump/pkg/stree"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func CreateBackup(cfg *config.Config) (string, error) {
	cfg.Log.Info("🚀 Starting backup creation...")

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("postgres_backup_%s.dump", timestamp)
	filePath := fmt.Sprintf("%s/%s", cfg.Postgres.BackupPath, fileName)

	if _, err := os.Stat(cfg.Postgres.BackupPath); os.IsNotExist(err) {
		cfg.Log.Warn("📂 Backup directory doesn't exist, creating...", "path", cfg.Postgres.BackupPath)
		err := os.MkdirAll(cfg.Postgres.BackupPath, 0755)
		if err != nil {
			return "", err
		}
	}

	err := os.Setenv("PGPASSWORD", cfg.Postgres.Password)
	if err != nil {
		return "", err
	}

	cmd := exec.Command(
		"pg_dump",
		"-U", cfg.Postgres.User,
		"-h", cfg.Postgres.Host,
		"-p", cfg.Postgres.Port,
		"-F", "c",
		"-f", filePath,
		cfg.Postgres.Dbname,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+cfg.Postgres.Password)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("❌ Error creating backup: %v\n%s", err, string(output))
	}

	if cfg.S3Client != nil {

		fileName, err := stree.UploadFileToS3(cfg.S3Client, cfg.BucketName, filePath, cfg.Postgres.BackupPath)
		if err != nil {
			cfg.Log.Error("❌ Error uploading to S3", "error", err)
			return filePath, nil
		}

		err = os.Remove(filePath)
		if err != nil {
			cfg.Log.Warn("⚠️ Failed to delete local file", "file", filePath, "error", err)
		}

		return fileName, nil
	}

	return filePath, nil
}
