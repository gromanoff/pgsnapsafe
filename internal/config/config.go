package config

import (
	"PostgresDump/pkg/email"
	"PostgresDump/pkg/slogger"
	"PostgresDump/pkg/stree"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	v "github.com/spf13/viper"
	"log"
	"log/slog"
	"strconv"
)

type Config struct {
	Log        *slog.Logger
	Postgres   *Postgres
	Backup     *BackupConfig
	S3Client   *s3.Client
	BucketName string
	SMTPClient *email.SMTPClient
}

type BackupConfig struct {
	Times      []string `mapstructure:"times"`
	KeepCopies int      `mapstructure:"keep_copies"`
	HeathCheck bool     `bool:"health_check"`
}

type Postgres struct {
	Host       string
	Port       string
	User       string
	Password   string
	Dbname     string
	BackupPath string
}

func (p *Postgres) new() *Postgres {
	return &Postgres{
		Host:       v.GetString("POSTGRESQL_HOST"),
		Port:       v.GetString("POSTGRESQL_PORT"),
		User:       v.GetString("POSTGRESQL_USER"),
		Password:   v.GetString("POSTGRESQL_PASSWORD"),
		Dbname:     v.GetString("POSTGRESQL_DBNAME"),
		BackupPath: v.GetString("DIRECTORY_BACKUP_PATH"),
	}
}

func Init() *Config {
	var cfg Config
	cfg.Log = initLogger()

	requiredVars := []string{
		"POSTGRESQL_HOST",
		"POSTGRESQL_PORT",
		"POSTGRESQL_USER",
		"POSTGRESQL_PASSWORD",
		"POSTGRESQL_DBNAME",

		"DIRECTORY_BACKUP_PATH",
	}

	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err == nil {
		log.Println("✅ .env file loaded!")
	} else {
		log.Println("⚠️ .env file not found, using system variables only.")
	}

	v.AutomaticEnv()

	if err := checkEnv(requiredVars); err != nil {
		log.Fatalf("Error checking environment variables: %v", err)
	}

	if v.GetString("S3_BUCKET_NAME") != "" {
		client := initS3Client()
		cfg.S3Client = client
	}
	if v.GetString("SMTP_HOST") != "" {
		client := initSMTPClient()
		cfg.SMTPClient = client
	}

	cfg.Postgres = cfg.Postgres.new()
	cfg.BucketName = v.GetString("S3_BUCKET_NAME")
	cfg.Backup = loadConfigBackup()

	if !v.GetBool("s3") {
		cfg.S3Client = nil
	}
	if !v.GetBool("smtp") {
		cfg.SMTPClient = nil
	}

	cfg.Log.Info("Environment initialization completed. ✅")
	return &cfg
}

func initLogger() *slog.Logger {
	return slogger.SetupLogger()
}

func initS3Client() *s3.Client {
	bucketName := v.GetString("S3_BUCKET_NAME")
	region := v.GetString("S3_REGION")
	accessKey := v.GetString("S3_ACCESS_KEY")
	secretKey := v.GetString("S3_SECRET_KEY")
	endpoint := v.GetString("S3_ENDPOINT")

	client, err := stree.InitS3Client(bucketName, region, accessKey, secretKey, endpoint)
	if err != nil {
		log.Fatalf("Error creating S3 client: %v", err)
	}
	return client
}

func initSMTPClient() *email.SMTPClient {
	port, err := strconv.Atoi(v.GetString("SMTP_PORT"))
	if err != nil {
		log.Fatalf("Error converting SMTP port: %v", err)
	}
	client := email.NewSMTPClient(
		v.GetString("SMTP_HOST"),
		port,
		v.GetString("SMTP_USER"),
		v.GetString("SMTP_PASS"),
		v.GetString("SMTP_SENDER_SIGN"),
	)
	return client
}

func loadConfigBackup() *BackupConfig {
	v.SetConfigFile("config.yml")
	err := v.ReadInConfig()
	if err != nil {
		log.Fatalf("❌ Error reading config.yml: %v", err)
	}

	backupCfg := v.Sub("backup")
	if backupCfg == nil {
		log.Fatalf("❌ Error: 'backup' section not found in config.yml")
	}

	var cfg BackupConfig
	err = backupCfg.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("❌ Error processing config: %v", err)
	}

	log.Printf("📌 Loaded backup settings: %+v\n", cfg)
	return &cfg
}

func checkEnv(requiredVars []string) error {
	var missingVars []string

	for _, key := range requiredVars {
		if !v.IsSet(key) || v.GetString(key) == "" {
			missingVars = append(missingVars, key)
		}
	}

	if len(missingVars) > 0 {
		return fmt.Errorf("missing or empty environment variables: %v", missingVars)
	}

	return nil
}
