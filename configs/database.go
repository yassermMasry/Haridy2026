package configs

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDatabase(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Africa/Cairo",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)
	gormLogger := logger.Default.LogMode(logger.Warn)
	if cfg.AppEnv == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 25))
	sqlDB.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 10))
	sqlDB.SetConnMaxLifetime(time.Duration(envInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute)
	return db, nil
}

func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(env(key, ""))
	if err != nil || value <= 0 {
		if err != nil && env(key, "") != "" {
			log.Printf("invalid %s, using %d", key, fallback)
		}
		return fallback
	}
	return value
}
