package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type config struct {
	HTTPAddr      string
	TemplatesDir  string
	StaticDir     string
	SessionSecret string
	WebOrigin     string
}

var appConfig config

func loadConfig() config {
	return config{
		HTTPAddr:      envOrDefault("HTTP_ADDR", ":8080"),
		TemplatesDir:  envOrDefault("TEMPLATES_DIR", "templates"),
		StaticDir:     envOrDefault("STATIC_DIR", "static"),
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

func templatePath(name string) string {
	return filepath.Join(appConfig.TemplatesDir, name)
}

func staticPath(name string) string {
	return filepath.Join(appConfig.StaticDir, name)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin == appConfig.WebOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
