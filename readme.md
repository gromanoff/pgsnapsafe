# 🚀 PgSnapSafe

![GitHub stars](https://img.shields.io/github/stars/gromanoff/pgsnapsafe?style=social)
![GitHub forks](https://img.shields.io/github/forks/gromanoff/pgsnapsafe?style=social)
![GitHub license](https://img.shields.io/github/license/gromanoff/pgsnapsafe)
![GitHub release](https://img.shields.io/github/v/release/gromanoff/pgsnapsafe)

[English](README.md) | [Русский](README.ru.md)


**PgSnapSafe** is an automated PostgreSQL backup service with S3 storage support and email notifications. It allows flexible scheduling of backups and efficient management of stored copies via a simple `.yml` configuration file.

## 📌 Features

✅ Automated PostgreSQL backup  
✅ Flexible backup scheduling (via YAML)  
✅ Retention policy for stored copies  
✅ **AWS S3 / MinIO** support  
✅ Email notifications for backup status  
✅ Built-in **health check**

## 🛠 Installation

### 1️⃣ Clone the repository
```bash
git clone https://github.com/yourusername/PgSnapSafe.git
cd PgSnapSafe
```

### 2️⃣ Configure environment variables
Create a `.env` file and add your database and S3 storage settings:

```ini
# PostgreSQL
POSTGRESQL_HOST=your-postgresql-host
POSTGRESQL_PORT=5432
POSTGRESQL_USER=your-user
POSTGRESQL_PASSWORD=your-password
POSTGRESQL_DBNAME=your-database

# Backup directory
DIRECTORY_BACKUP_PATH=/app/db_backups

# S3 (if enabled)
S3_BUCKET_NAME=your-bucket
S3_REGION=your-region
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_ENDPOINT=your-s3-endpoint

# Email notifications (if enabled)
SMTP_HOST=smtp.your-email.com
SMTP_PORT=587
SMTP_USER=your-email@example.com
SMTP_PASS=your-password
SMTP_SENDER_SIGN=PgSnapSafe
EMAIL_DELIVERY=your-notify-email@example.com
```

### 3️⃣ Configure backup schedule
Open `config.yml` and specify the backup times and number of stored copies:

```yaml
backup:
  times:
    - "02:00"
    - "14:00"
  keep_copies: 5

s3: true        # Enable S3 storage  
smtp: true      # Enable email notifications  
health_check: true  # Perform health check on startup  
```

### 4️⃣ Start with Docker
```bash
docker-compose up -d
```

## 🚀 Usage

### Run a backup manually
```bash
docker exec -it postgresdump /usr/local/bin/postgresdump
```

### Stop the service
```bash
docker-compose down
```

## 📜 License

This project is distributed under the **MIT License**. Feel free to use and contribute!

## 💚 Support the project with a donation 💚
USDT TRC20
```bash
TMpytFwhc5BaWpUeBkB9BdJndoHb7nFWkU
```
