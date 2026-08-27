package server

import (
	"net/http"

	authmodule "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/modules/auth"
	updatesmodule "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/modules/updates"
	usersmodule "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/modules/users"
	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/httpx"
)

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()

	authHandler := a.newAuthHandler()
	usersHandler := usersmodule.NewHandler(a.db, a.store)
	updatesHandler := updatesmodule.NewHandler(a.db, a.store)

	a.registerPlatformRoutes(mux)
	a.registerAuthRoutes(mux, authHandler)
	a.registerUserRoutes(mux, usersHandler)
	a.registerUpdateRoutes(mux, updatesHandler)

	return httpx.WithCORS(a.cfg.WebOrigin, mux)
}

func (a *App) registerPlatformRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/api/health", healthHandler)
}

func (a *App) registerAuthRoutes(mux *http.ServeMux, handler *authmodule.Handler) {
	mux.HandleFunc("/auth/signin", handler.SignIn)
	mux.HandleFunc("/auth/signup", handler.SignUp)
	mux.HandleFunc("/auth/logout", a.logoutHandler)
}

func (a *App) registerUserRoutes(mux *http.ServeMux, handler *usersmodule.Handler) {
	mux.HandleFunc("/profile", a.authMiddleware(handler.Profile))
	mux.HandleFunc("/profile/edit", a.authMiddleware(handler.ProfileEdit))
	mux.HandleFunc("/users/", handler.PublicProfile)
	mux.HandleFunc("/follow", a.authMiddleware(handler.Follow))
	mux.HandleFunc("/unfollow", a.authMiddleware(handler.Unfollow))
	mux.HandleFunc("/admin/ban-user", a.adminAuthMiddleware(handler.BanUser))
	mux.HandleFunc("/admin/unban-user", a.adminAuthMiddleware(handler.UnbanUser))
}

func (a *App) registerUpdateRoutes(mux *http.ServeMux, handler *updatesmodule.Handler) {
	mux.HandleFunc("/aspiration-update", a.authMiddleware(handler.AspirationUpdate))
	mux.HandleFunc("/aspiration-update/edit/", a.authMiddleware(handler.EditAspirationUpdate))
	mux.HandleFunc("/aspiration-update/delete/", a.authMiddleware(handler.DeleteAspirationUpdate))
	mux.HandleFunc("/like", a.authMiddleware(handler.Like))
	mux.HandleFunc("/unlike", a.authMiddleware(handler.Unlike))
	mux.HandleFunc("/update/", handler.Permalink)
	mux.HandleFunc("/comment/add", a.authMiddleware(handler.AddComment))
}

func (a *App) newAuthHandler() *authmodule.Handler {
	authRepository := authmodule.NewPostgresRepository(a.db)
	authService := authmodule.NewService(authRepository, authmodule.PasswordHasher{})
	return authmodule.NewHandler(authService, a.store, a.cfg.WebOrigin)
}
