package pages

import (
	"database/sql"
	"html/template"
	"net/http"

	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/domain"
	gorillasessions "github.com/gorilla/sessions"
)

type TemplatePathFunc func(string) string

type Handler struct {
	db           *sql.DB
	store        *gorillasessions.CookieStore
	templatePath TemplatePathFunc
}

func NewHandler(db *sql.DB, store *gorillasessions.CookieStore, templatePath TemplatePathFunc) *Handler {
	return &Handler{db: db, store: store, templatePath: templatePath}
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(h.templatePath("homepage.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := h.store.Get(r, "session-name")
	userLoggedIn := session.Values["email"] != nil

	var currentUserID int
	if userLoggedIn {
		email := session.Values["email"].(string)
		err = h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&currentUserID)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	recentUsers, err := h.fetchRecentUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recentUpdates, err := h.fetchRecentUpdates(currentUserID, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		UserLoggedIn  bool
		RecentUsers   []RecentUser
		RecentUpdates []domain.AspirationUpdate
	}{
		UserLoggedIn:  userLoggedIn,
		RecentUsers:   recentUsers,
		RecentUpdates: recentUpdates,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) Browse(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles(h.templatePath("browse.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := h.store.Get(r, "session-name")
	userLoggedIn := session.Values["email"] != nil

	var currentUserID int
	if userLoggedIn {
		email := session.Values["email"].(string)
		err = h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&currentUserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	users, err := h.fetchRecentUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recentUpdates, err := h.fetchRecentUpdates(currentUserID, 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		UserLoggedIn  bool
		RecentUpdates []domain.AspirationUpdate
		Users         []RecentUser
	}{
		UserLoggedIn:  userLoggedIn,
		RecentUpdates: recentUpdates,
		Users:         users,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) StaticPage(templateName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(h.templatePath(templateName))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

type RecentUser struct {
	ID              int
	Username        string
	DisplayName     string
	ProfileImageURL string
}

func (h *Handler) fetchRecentUsers() ([]RecentUser, error) {
	rows, err := h.db.Query(`
		SELECT id, COALESCE(username, ''), COALESCE(display_name, ''), COALESCE(profile_image_url, '')
		FROM users
		WHERE is_banned = false
		ORDER BY id DESC
		LIMIT 50
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recentUsers []RecentUser
	for rows.Next() {
		var user RecentUser
		if err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.ProfileImageURL); err != nil {
			return nil, err
		}
		recentUsers = append(recentUsers, user)
	}
	return recentUsers, nil
}

func (h *Handler) fetchRecentUpdates(currentUserID int, limit int) ([]domain.AspirationUpdate, error) {
	query := `
        WITH RankedUpdates AS (
            SELECT au.id, u.username, u.display_name, u.profile_image_url, au.content, au.created_at, 
                   COUNT(DISTINCT l.id) as like_count,
                   COUNT(DISTINCT c.id) as comment_count,
                   CASE WHEN $1 > 0 AND EXISTS (SELECT 1 FROM likes WHERE user_id = $1 AND update_id = au.id) THEN TRUE ELSE FALSE END as liked,
                   CASE WHEN $1 > 0 AND au.user_id = $1 THEN TRUE ELSE FALSE END as is_own_post,
                   ROW_NUMBER() OVER (PARTITION BY au.user_id ORDER BY au.created_at DESC) as rn
            FROM aspiration_updates au
            JOIN users u ON au.user_id = u.id
            LEFT JOIN likes l ON au.id = l.update_id
            LEFT JOIN comments c ON au.id = c.update_id
            WHERE u.is_banned = false
            GROUP BY au.id, u.username, u.display_name, u.profile_image_url, au.content, au.created_at, au.user_id
        )
        SELECT id, COALESCE(username, ''), COALESCE(display_name, ''), COALESCE(profile_image_url, ''), content, created_at, 
               like_count, comment_count, liked, is_own_post
        FROM RankedUpdates
        WHERE rn = 1
        ORDER BY created_at DESC
        LIMIT $2
    `
	rows, err := h.db.Query(query, currentUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recentUpdates []domain.AspirationUpdate
	for rows.Next() {
		var update domain.AspirationUpdate
		if err := rows.Scan(&update.ID, &update.Username, &update.DisplayName, &update.ProfileImageURL, &update.Content, &update.CreatedAt, &update.LikeCount, &update.CommentCount, &update.Liked, &update.IsOwnPost); err != nil {
			return nil, err
		}
		recentUpdates = append(recentUpdates, update)
	}
	return recentUpdates, nil
}
