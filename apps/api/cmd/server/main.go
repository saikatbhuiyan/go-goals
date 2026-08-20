package main

import (
	"database/sql"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	authmodule "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/modules/auth"
	pagesmodule "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/modules/pages"
	updatesmodule "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/modules/updates"
	usersmodule "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/modules/users"
	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/config"
	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/httpx"
	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/postgres"
	apisessions "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/sessions"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var (
	db    *sql.DB
	store *sessions.CookieStore
	cfg   config.Config
)

func main() {
	cfg = config.Load()
	db = postgres.ConnectFromEnv()
	defer db.Close()

	store = apisessions.NewCookieStore(cfg.SessionSecret)
	startServer(setupRoutes())
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()

	authHandler := newAuthHandler()
	pagesHandler := pagesmodule.NewHandler(db, store, templatePath)
	usersHandler := usersmodule.NewHandler(db, store, templatePath)
	updatesHandler := updatesmodule.NewHandler(db, store, templatePath)

	registerPlatformRoutes(mux)
	registerPageRoutes(mux, pagesHandler)
	registerAuthRoutes(mux, authHandler)
	registerUserRoutes(mux, usersHandler)
	registerUpdateRoutes(mux, updatesHandler)

	return httpx.WithCORS(cfg.WebOrigin, mux)
}

func registerPlatformRoutes(mux *http.ServeMux) {
	fs := http.FileServer(http.Dir(cfg.StaticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := staticPath(r.URL.Path)
		contentType := mime.TypeByExtension(filepath.Ext(path))
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		fs.ServeHTTP(w, r)
	})))
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/api/health", healthHandler)
}

func registerPageRoutes(mux *http.ServeMux, handler *pagesmodule.Handler) {
	mux.HandleFunc("/", handler.HomePage)
	mux.HandleFunc("/browse", handler.Browse)
	mux.HandleFunc("/terms", handler.StaticPage("pages/doc_terms.html"))
	mux.HandleFunc("/privacy", handler.StaticPage("pages/doc_privacy.html"))
	mux.HandleFunc("/community-guidelines", handler.StaticPage("pages/doc_community_guidelines.html"))
}

func registerAuthRoutes(mux *http.ServeMux, handler *authmodule.Handler) {
	mux.HandleFunc("/auth/signin", handler.SignIn)
	mux.HandleFunc("/auth/signup", handler.SignUp)
	mux.HandleFunc("/auth/logout", logoutHandler)
}

func registerUserRoutes(mux *http.ServeMux, handler *usersmodule.Handler) {
	mux.HandleFunc("/profile", authMiddleware(handler.Profile))
	mux.HandleFunc("/profile/edit", authMiddleware(handler.ProfileEdit))
	mux.HandleFunc("/users/", handler.PublicProfile)
	mux.HandleFunc("/follow", authMiddleware(handler.Follow))
	mux.HandleFunc("/unfollow", authMiddleware(handler.Unfollow))
	mux.HandleFunc("/admin/ban-user", adminAuthMiddleware(handler.BanUser))
	mux.HandleFunc("/admin/unban-user", adminAuthMiddleware(handler.UnbanUser))
}

func registerUpdateRoutes(mux *http.ServeMux, handler *updatesmodule.Handler) {
	mux.HandleFunc("/aspiration-update", authMiddleware(handler.AspirationUpdate))
	mux.HandleFunc("/aspiration-update/edit/", authMiddleware(handler.EditAspirationUpdate))
	mux.HandleFunc("/aspiration-update/delete/", authMiddleware(handler.DeleteAspirationUpdate))
	mux.HandleFunc("/like", authMiddleware(handler.Like))
	mux.HandleFunc("/unlike", authMiddleware(handler.Unlike))
	mux.HandleFunc("/update/", handler.Permalink)
	mux.HandleFunc("/comment/add", authMiddleware(handler.AddComment))
}

func newAuthHandler() *authmodule.Handler {
	authRepository := authmodule.NewPostgresRepository(db)
	authService := authmodule.NewService(authRepository, authmodule.PasswordHasher{})
	return authmodule.NewHandler(authService, store, templatePath)
}

func startServer(handler http.Handler) {
	log.Printf("Server starting on %s", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, handler))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name")
		if err != nil {
			log.Printf("Error getting session: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		email, ok := session.Values["email"].(string)
		if !ok || email == "" {
			log.Printf("No valid email in session")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		var isBanned bool
		err = db.QueryRow("SELECT is_banned FROM users WHERE email = $1", email).Scan(&isBanned)
		if err != nil {
			log.Printf("Error checking user ban status: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if isBanned {
			http.Error(w, "Your account has been banned", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r.WithContext(httpx.ContextWithEmail(r.Context(), email)))
	}
}

func adminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		email, ok := session.Values["email"].(string)
		if !ok || email == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM administrators WHERE email = $1", email).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(httpx.ContextWithEmail(r.Context(), email)))
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["email"] = nil
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
