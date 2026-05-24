package configs

import (
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName       string
	AppEnv        string
	AppPort       string
	AppSecret     string
	AppSecrets    []string
	JWTSecret     string
	JWTSecrets    []string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	SessionName   string
	SessionMaxAge int
	SessionSecure bool
	RedisAddr     string
	RedisPassword string
	RedisDB       string
}

func Load() Config {
	_ = godotenv.Load()
	appSecret := env("APP_SECRET", "change-me")
	jwtSecret := env("JWT_SECRET", appSecret)
	return Config{
		AppName:       env("APP_NAME", "Haridy Inventory"),
		AppEnv:        env("APP_ENV", "development"),
		AppPort:       env("APP_PORT", "8080"),
		AppSecret:     appSecret,
		AppSecrets:    splitSecrets(env("APP_SECRETS", appSecret)),
		JWTSecret:     jwtSecret,
		JWTSecrets:    splitSecrets(env("JWT_SECRETS", jwtSecret)),
		DBHost:        env("DB_HOST", "localhost"),
		DBPort:        env("DB_PORT", "5432"),
		DBUser:        env("DB_USER", "postgres"),
		DBPassword:    env("DB_PASSWORD", "postgres"),
		DBName:        env("DB_NAME", "haridy_inventory"),
		DBSSLMode:     env("DB_SSLMODE", "disable"),
		SessionName:   env("SESSION_NAME", "haridy_session"),
		SessionMaxAge: intEnv("SESSION_MAX_AGE", 28800),
		SessionSecure: boolEnv("SESSION_SECURE", false),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       env("REDIS_DB", "0"),
	}
}

func boolEnv(key string, fallback bool) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func splitSecrets(value string) []string {
	parts := strings.Split(value, ",")
	secrets := make([]string, 0, len(parts))
	for _, part := range parts {
		if secret := strings.TrimSpace(part); secret != "" {
			secrets = append(secrets, secret)
		}
	}
	if len(secrets) == 0 {
		return []string{"change-me"}
	}
	return secrets
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
