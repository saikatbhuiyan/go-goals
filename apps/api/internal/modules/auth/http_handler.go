package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/httpx"
	gorillasessions "github.com/gorilla/sessions"
)

type Handler struct {
	service   *Service
	store     *gorillasessions.CookieStore
	webOrigin string
}

type signInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpRequest struct {
	Email           string `json:"email"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type authResponse struct {
	Status string `json:"status"`
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
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request signInRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	email := NormalizeEmail(request.Email)
	password := request.Password

	err := h.service.SignIn(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			httpx.WriteError(w, http.StatusUnauthorized, "Invalid email or password")
		case errors.Is(err, ErrBannedAccount):
			httpx.WriteError(w, http.StatusForbidden, "Your account has been banned")
		default:
			httpx.WriteError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		}
		return
	}

	if err := h.signInUser(w, r, email); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to save session: "+err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusOK, authResponse{Status: "signed_in"})
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.Redirect(w, r, h.webURL("/auth/signup"), http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request signUpRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	input := SignUpInput{
		Email:           request.Email,
		Username:        request.Username,
		DisplayName:     request.DisplayName,
		Password:        request.Password,
		ConfirmPassword: request.ConfirmPassword,
	}
	email, err := h.service.SignUp(r.Context(), input)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, messageForSignUpError(err))
		return
	}

	if err := h.signInUser(w, r, email); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to save session: "+err.Error())
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, authResponse{Status: "signed_up"})
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
