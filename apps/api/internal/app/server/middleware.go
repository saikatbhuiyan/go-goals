package server

import (
	"log"
	"net/http"

	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/httpx"
)

func (a *App) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := a.store.Get(r, "session-name")
		if err != nil {
			log.Printf("Error getting session: %v", err)
			http.Redirect(w, r, a.webURL("/auth/signin"), http.StatusSeeOther)
			return
		}

		email, ok := session.Values["email"].(string)
		if !ok || email == "" {
			log.Printf("No valid email in session")
			http.Redirect(w, r, a.webURL("/auth/signin"), http.StatusSeeOther)
			return
		}

		var isBanned bool
		err = a.db.QueryRow("SELECT is_banned FROM users WHERE email = $1", email).Scan(&isBanned)
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

func (a *App) adminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := a.store.Get(r, "session-name")
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
		err = a.db.QueryRow("SELECT COUNT(*) FROM administrators WHERE email = $1", email).Scan(&count)
		if err != nil || count == 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(httpx.ContextWithEmail(r.Context(), email)))
	}
}
