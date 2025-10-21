
package backup

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)
func BackupDatabase() error {
	log.Println("🚀 Starting database backup process...")
	host := viper.GetString("DB_HOST")
	port := viper.GetString("DB_PORT")
	user := viper.GetString("DB_USER")
	password := viper.GetString("DB_PASSWORD")
	dbname := viper.GetString("DATABASE_NAME")
	if host == "" || port == "" || user == "" || dbname == "" {
		return fmt.Errorf("database configuration is incomplete")
	}
	backupDir := "./backups"
	if err := os.MkdirAll(backupDir, os.ModePerm); err != nil {
		log.Printf("❌ Failed to create backup directory: %v", err)
		return err
	}
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	fileName := fmt.Sprintf("%s/backup_%s_%s.dump", backupDir, dbname, timestamp)
	os.Setenv("PGPASSWORD", password)
	cmd := exec.Command("pg_dump",
		"-h", host,
		"-p", port,
		"-U", user,
		"-d", dbname,
		"-F", "c", // Custom format (compressed, recommended)
		"-b", // Include large objects
		"-v", // Verbose mode for better logging
		"-f", fileName,
	)

	// Capture command output for logging
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Backup failed. Command output:\n%s", string(output))
		return fmt.Errorf("pg_dump execution failed: %w", err)
	}
	log.Printf("✅ Database backup successful! File saved to: %s", fileName)
	log.Printf("pg_dump output:\n%s", string(output))

	return nil
}
func StartAutoBackupScheduler() {
	log.Println("⚙️ Initializing automatic backup scheduler...")
	loc, _ := time.LoadLocation("Asia/Bangkok")
	scheduler := cron.New(cron.WithLocation(loc))
	spec := "0 0 * * *" // This means at 00:00 every day
	_, err := scheduler.AddFunc(spec, func() {
		if err := BackupDatabase(); err != nil {
			log.Printf("‼️ Automatic backup failed: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("❌ Could not start backup scheduler: %v", err)
	}
	scheduler.Start()
	log.Printf("🎉 Auto backup scheduler started. Backups will run daily at midnight (Bangkok time).")
}
