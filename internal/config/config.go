package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all runtime configuration, sourced from environment variables.
// There are no baked-in defaults for values that affect what data is considered
// real vs. fake (e.g. thresholds) beyond documented, sane fallbacks for local dev.
type Config struct {
	Port                  string
	DatabaseURL           string
	OnlineThresholdSeconds int
}

func Load() (Config, error) {
	cfg := Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	thresholdStr := getEnv("ONLINE_THRESHOLD_SECONDS", "90")
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		return Config{}, fmt.Errorf("invalid ONLINE_THRESHOLD_SECONDS: %w", err)
	}
	cfg.OnlineThresholdSeconds = threshold

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
