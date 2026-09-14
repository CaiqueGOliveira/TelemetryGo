package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort              string
	DBMaxOpenConns          int
	DBMaxIdleConns          int
	DBConnMaxLifetime       time.Duration
	JWTRefreshExpiration    time.Duration
	JWTAccessExpiration     time.Duration
	PasswordResetExpiration time.Duration
}

func Load() Config {
	return Config{
		ServerPort:              envOr("SERVER_PORT", "8080"),
		DBMaxOpenConns:          envIntOr("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:          envIntOr("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime:       envDurationOr("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		JWTRefreshExpiration:    envDurationOr("JWT_REFRESH_EXPIRATION", 168*time.Hour),
		JWTAccessExpiration:     envDurationOr("JWT_ACCESS_EXPIRATION", 10*time.Minute),
		PasswordResetExpiration: envDurationOr("PASSWORD_RESET_EXPIRATION", 30*time.Minute),
	}
}

func envOr(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOr(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func envDurationOr(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
