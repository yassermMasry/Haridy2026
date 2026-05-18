package jobs

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"haridy2026/internal/models"
	"haridy2026/internal/services"

	"gorm.io/gorm"
)

func Start(ctx context.Context, db *gorm.DB) {
	erp := services.NewERPService(db)
	go schedule(ctx, 30*time.Minute, func() {
		if err := erp.GenerateNotifications(); err != nil {
			slog.Error("notification job failed", "error", err)
		}
	})
	go schedule(ctx, 6*time.Hour, func() {
		cleanup(ctx, db)
	})
	go schedule(ctx, 3*time.Hour, func() {
		if err := services.RunReconciliation(db, "all", nil); err != nil {
			slog.Error("reconciliation job failed", "error", err)
		}
	})
	go schedule(ctx, 24*time.Hour, func() {
		writeBackupMarker(db)
	})
}

func schedule(ctx context.Context, interval time.Duration, fn func()) {
	fn()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fn()
		}
	}
}

func cleanup(ctx context.Context, db *gorm.DB) {
	_ = ctx
	db.Where("created_at < ? AND success = ?", time.Now().Add(-7*24*time.Hour), true).Delete(&models.LoginAttempt{})
}

func writeBackupMarker(db *gorm.DB) {
	dir := filepath.Join("storage", "backups")
	if err := os.MkdirAll(dir, 0755); err != nil {
		slog.Error("backup dir failed", "error", err)
		return
	}
	name := filepath.Join(dir, time.Now().Format("20060102-150405")+".backup.txt")
	if err := os.WriteFile(name, []byte("Configure pg_dump in production scheduler for full database backups.\n"), 0644); err != nil {
		slog.Error("backup marker failed", "error", err)
		_ = db.Create(&models.BackupVerification{BackupName: filepath.Base(name), StorageURI: name, Status: "failed", Details: err.Error()}).Error
		return
	}
	completed := time.Now()
	_ = db.Create(&models.BackupVerification{BackupName: filepath.Base(name), StorageURI: name, Status: "placeholder", Details: "pg_dump backup script must run in production", CompletedAt: &completed}).Error
}
