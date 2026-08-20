package auth

import (
	"errors"
	"html/template"
	"net/http"

	gorillasessions "github.com/gorilla/sessions"
)

type TemplatePathFunc func(string) string

type Handler struct {
	service      *Service
	store        *gorillasessions.CookieStore
	templatePath TemplatePathFunc
}

type PageData struct {
	Email        string
	Username     string
	DisplayName  string
	ErrorMessage string
}

func NewHandler(service *Service, store *gorillasessions.CookieStore, templatePath TemplatePathFunc) *Handler {
	return &Handler{
		service:      service,
		store:        store,
		templatePath: templatePath,
	}
}

func (h *Handler) SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.renderAuthPage(w, "signin.html", PageData{})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")
	data := PageData{Email: email}

	err := h.service.SignIn(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			h.renderAuthPage(w, "signin.html", withError(data, "Invalid email or password"))
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
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.renderAuthPage(w, "signup.html", PageData{})
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
	data := PageData{
		Email:       NormalizeEmail(input.Email),
		Username:    NormalizeUsername(input.Username),
		DisplayName: input.DisplayName,
	}

	email, err := h.service.SignUp(r.Context(), input)
	if err != nil {
		h.renderAuthPage(w, "signup.html", withError(data, messageForSignUpError(err)))
		return
	}

	if err := h.signInUser(w, r, email); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/profile/edit", http.StatusSeeOther)
}

func (h *Handler) renderAuthPage(w http.ResponseWriter, templateName string, data PageData) {
	tmpl, err := template.ParseFiles(h.templatePath(templateName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) signInUser(w http.ResponseWriter, r *http.Request, email string) error {
	session, _ := h.store.Get(r, "session-name")
	session.Values["email"] = email
	delete(session.Values, "newUser")
	return session.Save(r, w)
}

func withError(data PageData, message string) PageData {
	data.ErrorMessage = message
	return data
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
