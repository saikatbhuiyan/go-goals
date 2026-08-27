package auth

import (
	"errors"
	"net/http"
	"strings"

	gorillasessions "github.com/gorilla/sessions"
)

type Handler struct {
	service   *Service
	store     *gorillasessions.CookieStore
	webOrigin string
}

func NewHandler(service *Service, store *gorillasessions.CookieStore, webOrigin string) *Handler {
	return &Handler{
		service:   service,
		store:     store,
		webOrigin: strings.TrimRight(webOrigin, "/"),
	}
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, h.webURL("/auth/signin"), http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	err := h.service.SignIn(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		case errors.Is(err, ErrBannedAccount):
			http.Error(w, "Your account has been banned", http.StatusForbidden)
		default:
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if err := h.signInUser(w, r, email); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.webURL("/"), http.StatusSeeOther)
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, h.webURL("/auth/signup"), http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	input := SignUpInput{
		Email:           r.FormValue("email"),
		Username:        r.FormValue("username"),
		DisplayName:     r.FormValue("display_name"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm_password"),
	}
	email, err := h.service.SignUp(r.Context(), input)
	if err != nil {
		http.Error(w, messageForSignUpError(err), http.StatusBadRequest)
		return
	}

	if err := h.signInUser(w, r, email); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, h.webURL("/"), http.StatusSeeOther)
}

func (h *Handler) signInUser(w http.ResponseWriter, r *http.Request, email string) error {
	session, _ := h.store.Get(r, "session-name")
	session.Values["email"] = email
	delete(session.Values, "newUser")
	return session.Save(r, w)
}

func (h *Handler) webURL(path string) string {
	if h.webOrigin == "" {
		return path
	}
	return h.webOrigin + path
}

func messageForSignUpError(err error) string {
	switch {
	case errors.Is(err, ErrDuplicateIdentity):
		return "Email or username is already in use"
	case errors.Is(err, ErrInvalidSignup):
		return "Email, username, and password are required"
	case errors.Is(err, ErrInvalidUsername):
		return "Username can only use letters, numbers, underscores, and hyphens"
	case errors.Is(err, ErrWeakPassword):
		return "Password must be at least 8 characters"
	case errors.Is(err, ErrPasswordMismatch):
		return "Passwords do not match"
	default:
		return "Failed to create user: " + err.Error()
	}
}
