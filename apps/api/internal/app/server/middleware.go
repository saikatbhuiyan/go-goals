package server

import (
	"log"
	"net/http"

	"github.com/saikatbhuiyan/go-goals/internal/platform/httpx"
)

func (a *App) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := a.store.Get(r, "session-name")
		if err != nil {
			log.Printf("Error getting session: %v", err)
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		email, ok := session.Values["email"].(string)
		if !ok || email == "" {
			log.Printf("No valid email in session")
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var isBanned bool
		err = a.db.QueryRow("SELECT is_banned FROM users WHERE email = $1", email).Scan(&isBanned)
		if err != nil {
			log.Printf("Error checking user ban status: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		if isBanned {
			httpx.WriteError(w, http.StatusForbidden, "Your account has been banned")
			return
		}

		next.ServeHTTP(w, r.WithContext(httpx.ContextWithEmail(r.Context(), email)))
	}
}

func (a *App) adminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := a.store.Get(r, "session-name")
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		email, ok := session.Values["email"].(string)
		if !ok || email == "" {
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var count int
		err = a.db.QueryRow("SELECT COUNT(*) FROM administrators WHERE email = $1", email).Scan(&count)
		if err != nil || count == 0 {
			httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		next.ServeHTTP(w, r.WithContext(httpx.ContextWithEmail(r.Context(), email)))
	}
}
