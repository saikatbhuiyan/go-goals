package server

import (
	"database/sql"
	"log"
	"net/http"

	gorillasessions "github.com/gorilla/sessions"
	"github.com/saikatbhuiyan/go-goals/internal/platform/config"
	"github.com/saikatbhuiyan/go-goals/internal/platform/postgres"
	apisessions "github.com/saikatbhuiyan/go-goals/internal/platform/sessions"
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
