package config

import (
	"os"
	"strings"
)

type Config struct {
	HTTPAddr      string
	SessionSecret string
	WebOrigin     string
}

func Load() Config {
	return Config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		SessionSecret: envOrDefault("SESSION_SECRET", "development-session-secret"),
		WebOrigin:     envOrDefault("WEB_ORIGIN", "http://localhost:3000"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
