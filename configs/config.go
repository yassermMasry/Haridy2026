package configs

import (
	"log/slog"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName       string
	AppEnv        string
	AppPort       string
	AppSecret     string
	JWTSecret     string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	SessionName   string
	RedisAddr     string
	RedisPassword string
	RedisDB       string
}

func Load() Config {
	_ = godotenv.Load()
	return Config{
		AppName:       env("APP_NAME", "Haridy Inventory"),
		AppEnv:        env("APP_ENV", "development"),
		AppPort:       env("APP_PORT", "8080"),
		AppSecret:     env("APP_SECRET", "change-me"),
		JWTSecret:     env("JWT_SECRET", env("APP_SECRET", "change-me")),
		DBHost:        env("DB_HOST", "localhost"),
		DBPort:        env("DB_PORT", "5432"),
		DBUser:        env("DB_USER", "postgres"),
		DBPassword:    env("DB_PASSWORD", "postgres"),
		DBName:        env("DB_NAME", "haridy_inventory"),
		DBSSLMode:     env("DB_SSLMODE", "disable"),
		SessionName:   env("SESSION_NAME", "haridy_session"),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       env("REDIS_DB", "0"),
	}
}

func (c Config) LogLevel() slog.Level {
	if strings.EqualFold(c.AppEnv, "production") {
		return slog.LevelInfo
	}
	return slog.LevelDebug
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
