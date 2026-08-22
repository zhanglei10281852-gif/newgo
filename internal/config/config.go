package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr, DatabasePath, BusinessTimezone, LogLevel     string
	SessionTTL, PermitTTL, WorkerInterval, ShutdownTimeout time.Duration
	WorkerBatchSize                                        int
}

func Load() (Config, error) {
	c := Config{HTTPAddr: env("HTTP_ADDR", ":8080"), DatabasePath: env("DATABASE_PATH", "./data/geothermal.db"), BusinessTimezone: env("BUSINESS_TIMEZONE", "Asia/Shanghai"), LogLevel: env("LOG_LEVEL", "info"), WorkerBatchSize: intEnv("WORKER_BATCH_SIZE", 50)}
	var err error
	if c.SessionTTL, err = durationEnv("SESSION_TTL", "12h"); err != nil {
		return Config{}, err
	}
	if c.PermitTTL, err = durationEnv("PERMIT_TTL", "30m"); err != nil {
		return Config{}, err
	}
	if c.WorkerInterval, err = durationEnv("WORKER_INTERVAL", "2s"); err != nil {
		return Config{}, err
	}
	if c.ShutdownTimeout, err = durationEnv("SHUTDOWN_TIMEOUT", "10s"); err != nil {
		return Config{}, err
	}
	if c.WorkerBatchSize < 1 {
		return Config{}, fmt.Errorf("WORKER_BATCH_SIZE must be positive")
	}
	if _, err := time.LoadLocation(c.BusinessTimezone); err != nil {
		return Config{}, fmt.Errorf("business timezone: %w", err)
	}
	return c, nil
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func intEnv(key string, fallback int) int {
	var value int
	if _, err := fmt.Sscanf(os.Getenv(key), "%d", &value); err != nil {
		return fallback
	}
	return value
}
func durationEnv(key, fallback string) (time.Duration, error) {
	value, err := time.ParseDuration(env(key, fallback))
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}
