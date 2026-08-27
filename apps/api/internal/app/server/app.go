package server

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/config"
	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/postgres"
	apisessions "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/sessions"
	gorillasessions "github.com/gorilla/sessions"
)

type App struct {
	cfg   config.Config
	db    *sql.DB
	store *gorillasessions.CookieStore
}

func Run() error {
	cfg := config.Load()
	db := postgres.ConnectFromEnv()
	defer db.Close()

	app := New(cfg, db, apisessions.NewCookieStore(cfg.SessionSecret))
	return app.Start()
}

func New(cfg config.Config, db *sql.DB, store *gorillasessions.CookieStore) *App {
	return &App{cfg: cfg, db: db, store: store}
}

func (a *App) Start() error {
	log.Printf("Server starting on %s", a.cfg.HTTPAddr)
	return http.ListenAndServe(a.cfg.HTTPAddr, a.Handler())
}

func (a *App) webURL(path string) string {
	webOrigin := strings.TrimRight(a.cfg.WebOrigin, "/")
	if webOrigin == "" {
		return path
	}
	return webOrigin + path
}
