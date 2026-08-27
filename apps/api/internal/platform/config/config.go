package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	HTTPAddr      string
	TemplatesDir  string
	StaticDir     string
	SessionSecret string
	WebOrigin     string
}

func Load() Config {
	return Config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		TemplatesDir:  envOrDefault("TEMPLATES_DIR", "apps/api/templates"),
		StaticDir:     envOrDefault("STATIC_DIR", "apps/api/static"),
		SessionSecret: envOrDefault("SESSION_SECRET", "development-session-secret"),
		WebOrigin:     envOrDefault("WEB_ORIGIN", "http://localhost:3000"),
	}
}

func (c Config) TemplatePath(name string) string {
	return filepath.Join(c.TemplatesDir, name)
}

func (c Config) StaticPath(name string) string {
	return filepath.Join(c.StaticDir, name)
}

func envOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
