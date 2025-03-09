package healthcheck

import (
	"PostgresDump/internal/config"
	"PostgresDump/internal/services/backups"
	"PostgresDump/pkg/email"
	"PostgresDump/pkg/stree"
	"context"
	"database/sql"
	"fmt"
	v "github.com/spf13/viper"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/lib/pq"
)

// HealthCheck performs connection and operation checks
func HealthCheck(cfg *config.Config) error {
	log.Println("🩺 Starting service health check...")

	// 1️⃣ Check PostgreSQL connection
	log.Println("🛢 Checking PostgreSQL connection...")
	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Dbname,
	))
	if err != nil {
		return fmt.Errorf("❌ Error connecting to PostgreSQL: %w", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("❌ PostgreSQL is unavailable: %w", err)
	}
	log.Println("✅ PostgreSQL connection successful")

	// 2️⃣ Check S3 connection
	if cfg.S3Client != nil {
		log.Println("☁️ Checking S3 connection...")
		_, err := cfg.S3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
		if err != nil {
			return fmt.Errorf("❌ Error connecting to S3: %w", err)
		}
		log.Println("✅ S3 connection successful")
	} else {
		log.Println("⚠️ S3 not initialized, skipping test")
	}

	// 3️⃣ Create a test backup
	log.Println("🛠 Creating test backup...")
	testBackup, err := backups.CreateBackup(cfg)
	if err != nil {
		return fmt.Errorf("❌ Error creating test backup: %w", err)
	}
	log.Println("✅ Test backup successfully created:", testBackup)

	// Wait a couple of seconds to ensure the backup is uploaded to S3
	time.Sleep(30 * time.Second)

	// 4️⃣ Check if the backup appeared in S3
	if cfg.S3Client != nil {
		log.Println("🔍 Checking for test backup in S3...")
		files, err := stree.ListFilesInS3Directory(cfg.S3Client, cfg.BucketName, cfg.Postgres.BackupPath)
		if err != nil {
			return fmt.Errorf("❌ Error checking backup in S3: %w", err)
		}

		found := false
		for _, file := range files {
			if file == testBackup {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("❌ Test backup not found in S3")
		}
		log.Println("✅ Test backup found in S3")
	}

	// 5️⃣ Delete test backup from S3
	if cfg.S3Client != nil {
		log.Println("🗑 Deleting test backup from S3...")
		err := stree.DeleteFileFromS3(cfg.S3Client, cfg.BucketName, testBackup)
		if err != nil {
			return fmt.Errorf("❌ Error deleting test backup from S3: %w", err)
		}
		log.Println("✅ Test backup successfully deleted from S3")
	}

	if v.GetBool("smtp") {
		if err = email.SendEmail(cfg.SMTPClient, v.GetString("email_delivery"), testBackup); err != nil {
			cfg.Log.Error("Error sending email", "error", err)
		}
	}

	log.Println("🎉 All checks passed successfully!")
	return nil
}
