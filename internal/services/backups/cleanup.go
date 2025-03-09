package backups

import (
	"PostgresDump/internal/config"
	"PostgresDump/pkg/stree"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func CleanupOldBackups(cfg *config.Config) error {
	cfg.Log.Info("🔄 Starting cleanup of old backups...")

	if cfg.S3Client != nil {
		return cleanupOldBackupsFromS3(cfg)
	}

	files, err := filepath.Glob(cfg.Postgres.BackupPath + "/*.dump")
	if err != nil {
		cfg.Log.Error("❌ Error getting list of local files", "error", err)
		return err
	}

	if len(files) == 0 {
		cfg.Log.Warn("📂 No local backups to clean up")
		return nil
	}

	if cfg.Backup.KeepCopies <= 0 {
		cfg.Log.Warn("⚠️ KeepCopies parameter <= 0, cleanup not performed", "keepCopies", cfg.Backup.KeepCopies)
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		if errI != nil || errJ != nil {
			cfg.Log.Warn("⚠️ Error getting file date, comparing by name", "fileI", files[i], "fileJ", files[j])
			return files[i] < files[j]
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	toDelete := len(files) - cfg.Backup.KeepCopies
	if toDelete <= 0 {
		cfg.Log.Info("✅ Number of backups within limit, deletion not required")
		return nil
	}

	for i := 0; i < toDelete; i++ {
		err := os.Remove(files[i])
		if err != nil {
			cfg.Log.Warn("⚠️ Error deleting old local backup", "file", files[i], "error", err)
			continue
		}
		cfg.Log.Info("✅ Successfully deleted old local backup", "file", files[i])
	}

	cfg.Log.Info("🧹 Local backups cleanup completed")

	return nil
}

func cleanupOldBackupsFromS3(cfg *config.Config) error {

	files, err := stree.ListFilesInS3Directory(cfg.S3Client, cfg.BucketName, cfg.Postgres.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to get list of backups from S3: %w", err)
	}

	if len(files) == 0 {
		cfg.Log.Warn("📂 No backups in S3 to delete")
		return nil
	}

	if len(files) <= cfg.Backup.KeepCopies {
		cfg.Log.Info("✅ Number of backups within limit, deletion not required",
			"keepCopies", cfg.Backup.KeepCopies, "totalFiles", len(files))
		return nil
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})

	toDelete := len(files) - cfg.Backup.KeepCopies

	for i := 0; i < toDelete; i++ {
		fileToDelete := files[i]

		err := stree.DeleteFileFromS3(cfg.S3Client, cfg.BucketName, fileToDelete)
		if err != nil {
			cfg.Log.Warn("⚠️ Error deleting backup from S3", "file", fileToDelete, "error", err)
			continue
		}

	}

	return nil
}
