package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

const (
	passwordHashIterations = 210000
	passwordSaltBytes      = 16
	passwordKeyBytes       = 32
)

type authPageData struct {
	Email        string
	Username     string
	DisplayName  string
	ErrorMessage string
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderAuthPage(w, "signin.html", authPageData{})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := normalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")
	data := authPageData{Email: email}

	var passwordHash sql.NullString
	var isBanned bool
	err := db.QueryRow("SELECT password_hash, is_banned FROM users WHERE email = $1", email).Scan(&passwordHash, &isBanned)
	if err != nil {
		if err == sql.ErrNoRows {
			renderAuthPage(w, "signin.html", withAuthError(data, "Invalid email or password"))
			return
		}
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isBanned {
		http.Error(w, "Your account has been banned", http.StatusForbidden)
		return
	}

	if !passwordHash.Valid || passwordHash.String == "" {
		renderAuthPage(w, "signin.html", withAuthError(data, "Invalid email or password"))
		return
	}

	ok, err := verifyPassword(password, passwordHash.String)
	if err != nil || !ok {
		renderAuthPage(w, "signin.html", withAuthError(data, "Invalid email or password"))
		return
	}

	if err := signInUser(w, r, email); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderAuthPage(w, "signup.html", authPageData{})
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := normalizeEmail(r.FormValue("email"))
	username := normalizeUsername(r.FormValue("username"))
	displayName := strings.TrimSpace(r.FormValue("display_name"))
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	data := authPageData{
		Email:       email,
		Username:    username,
		DisplayName: displayName,
	}

	if email == "" || username == "" || password == "" {
		renderAuthPage(w, "signup.html", withAuthError(data, "Email, username, and password are required"))
		return
	}
	if !isValidUsername(username) {
		renderAuthPage(w, "signup.html", withAuthError(data, "Username can only use letters, numbers, underscores, and hyphens"))
		return
	}
	if len(password) < 8 {
		renderAuthPage(w, "signup.html", withAuthError(data, "Password must be at least 8 characters"))
		return
	}
	if password != confirmPassword {
		renderAuthPage(w, "signup.html", withAuthError(data, "Passwords do not match"))
		return
	}

	passwordHash, err := hashPassword(password)
	if err != nil {
		http.Error(w, "Failed to hash password: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if displayName == "" {
		displayName = username
	}

	_, err = db.Exec(
		"INSERT INTO users (email, username, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		email,
		username,
		displayName,
		passwordHash,
	)
	if err != nil {
		if isUniqueViolation(err) {
			renderAuthPage(w, "signup.html", withAuthError(data, "Email or username is already in use"))
			return
		}
		http.Error(w, "Failed to create user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := signInUser(w, r, email); err != nil {
		http.Error(w, "Failed to save session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/profile/edit", http.StatusSeeOther)
}

func renderAuthPage(w http.ResponseWriter, templateName string, data authPageData) {
	tmpl, err := template.ParseFiles("templates/" + templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func withAuthError(data authPageData, message string) authPageData {
	data.ErrorMessage = message
	return data
}

func signInUser(w http.ResponseWriter, r *http.Request, email string) error {
	session, _ := store.Get(r, "session-name")
	session.Values["email"] = email
	delete(session.Values, "newUser")
	return session.Save(r, w)
}

func hashPassword(password string) (string, error) {
	salt := make([]byte, passwordSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, passwordHashIterations, passwordKeyBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(
		"pbkdf2-sha256$%d$%s$%s",
		passwordHashIterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

func verifyPassword(password string, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false, errors.New("unsupported password hash")
	}

	iterations, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false, err
	}

	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}

	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, len(expectedKey))
	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(key, expectedKey) == 1, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func isValidUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}

	for _, char := range username {
		if char >= 'a' && char <= 'z' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char == '_' || char == '-' {
			continue
		}
		return false
	}

	return true
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
