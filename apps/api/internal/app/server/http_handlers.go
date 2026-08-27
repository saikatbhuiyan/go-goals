package server

import (
	"net/http"

	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/httpx"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (a *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	session, _ := a.store.Get(r, "session-name")
	session.Values["email"] = nil
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to save session: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
