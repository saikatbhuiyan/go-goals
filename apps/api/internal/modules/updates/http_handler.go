package updates

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/domain"
	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/httpx"
	gorillasessions "github.com/gorilla/sessions"
)

type Handler struct {
	db    *sql.DB
	store *gorillasessions.CookieStore
}

func NewHandler(db *sql.DB, store *gorillasessions.CookieStore) *Handler {
	return &Handler{db: db, store: store}
}

func (h *Handler) AspirationUpdate(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := r.FormValue("content")
	updateID := r.FormValue("update_id")

	if content == "" {
		http.Error(w, (&domain.ValidationError{Field: "content", Message: "Content cannot be empty"}).Error(), http.StatusBadRequest)
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		http.Error(w, (&domain.DatabaseError{Operation: "fetching user ID", Err: err}).Error(), http.StatusInternalServerError)
		return
	}

	if updateID != "" {
		_, err = h.db.Exec("UPDATE aspiration_updates SET content = $1 WHERE id = $2 AND user_id = $3", content, updateID, userID)
		if err != nil {
			http.Error(w, (&domain.DatabaseError{Operation: "updating aspiration update", Err: err}).Error(), http.StatusInternalServerError)
			return
		}
	} else {
		_, err = h.db.Exec("INSERT INTO aspiration_updates (user_id, content) VALUES ($1, $2)", userID, content)
		if err != nil {
			http.Error(w, (&domain.DatabaseError{Operation: "creating aspiration update", Err: err}).Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) EditAspirationUpdate(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updateID := strings.TrimPrefix(r.URL.Path, "/aspiration-update/edit/")
	if updateID == "" {
		http.NotFound(w, r)
		return
	}

	var update domain.AspirationUpdate
	err := h.db.QueryRow("SELECT id, content FROM aspiration_updates WHERE id = $1 AND user_id = (SELECT id FROM users WHERE email = $2)", updateID, email).Scan(&update.ID, &update.Content)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Failed to fetch aspiration update: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, update)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Content cannot be empty", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("UPDATE aspiration_updates SET content = $1 WHERE id = $2", content, updateID)
	if err != nil {
		http.Error(w, "Failed to update aspiration update: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAspirationUpdate(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updateID := strings.TrimPrefix(r.URL.Path, "/aspiration-update/delete/")
	if updateID == "" {
		http.NotFound(w, r)
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM comments WHERE update_id = $1", updateID); err != nil {
		http.Error(w, "Failed to delete associated comments: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err = tx.Exec("DELETE FROM likes WHERE update_id = $1", updateID); err != nil {
		http.Error(w, "Failed to delete associated likes: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err = tx.Exec("DELETE FROM aspiration_updates WHERE id = $1 AND user_id = (SELECT id FROM users WHERE email = $2)", updateID, email); err != nil {
		http.Error(w, "Failed to delete aspiration update: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if err = tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Like(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updateID := r.FormValue("update_id")
	if updateID == "" {
		http.Error(w, "Missing update_id", http.StatusBadRequest)
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var postOwnerID int
	err = h.db.QueryRow("SELECT user_id FROM aspiration_updates WHERE id = $1", updateID).Scan(&postOwnerID)
	if err != nil {
		http.Error(w, "Failed to get post owner: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if userID == postOwnerID {
		http.Error(w, "You cannot like your own post", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("INSERT INTO likes (user_id, update_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, updateID)
	if err != nil {
		http.Error(w, "Failed to add like: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Unlike(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	updateID := r.FormValue("update_id")
	if updateID == "" {
		http.Error(w, "Missing update_id", http.StatusBadRequest)
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		http.Error(w, "Failed to get user ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec("DELETE FROM likes WHERE user_id = $1 AND update_id = $2", userID, updateID)
	if err != nil {
		http.Error(w, "Failed to remove like: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Permalink(w http.ResponseWriter, r *http.Request) {
	updateID := strings.TrimPrefix(r.URL.Path, "/update/")
	log.Printf("Handling update permalink for ID: %s", updateID)

	var update domain.AspirationUpdate
	var currentUserID int
	var isBanned bool

	session, _ := h.store.Get(r, "session-name")
	if email, ok := session.Values["email"].(string); ok {
		err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&currentUserID)
		if err != nil {
			log.Printf("Error getting user ID: %v", err)
		}
	}

	query := `
        SELECT au.id, u.username, COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, ''), au.content, au.created_at, 
               COUNT(l.id) as like_count,
               CASE WHEN $1 > 0 AND EXISTS (SELECT 1 FROM likes WHERE user_id = $1 AND update_id = au.id) THEN TRUE ELSE FALSE END as liked,
               CASE WHEN $1 > 0 AND au.user_id = $1 THEN TRUE ELSE FALSE END as is_own_post,
               u.is_banned
        FROM aspiration_updates au
        JOIN users u ON au.user_id = u.id
        LEFT JOIN likes l ON au.id = l.update_id
        WHERE au.id = $2
        GROUP BY au.id, u.username, u.display_name, u.profile_image_url, au.content, au.created_at, au.user_id, u.is_banned
    `

	err := h.db.QueryRow(query, currentUserID, updateID).Scan(
		&update.ID, &update.Username, &update.DisplayName, &update.ProfileImageURL, &update.Content,
		&update.CreatedAt, &update.LikeCount, &update.Liked, &update.IsOwnPost, &isBanned,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Update not found: %s", updateID)
			http.Error(w, "Update not found", http.StatusNotFound)
		} else {
			log.Printf("Error fetching update: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	if isBanned {
		http.Error(w, "This update is not available", http.StatusForbidden)
		return
	}

	comments, err := h.getComments(updateID)
	if err != nil {
		http.Error(w, (&domain.DatabaseError{Operation: "fetching comments", Err: err}).Error(), http.StatusInternalServerError)
		return
	}
	h.modifyBannedUserComments(comments)

	data := UpdatePageData{
		Update:          update,
		Comments:        comments,
		TotalComments:   countTotalComments(comments),
		IsAuthenticated: currentUserID != 0,
	}

	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	updateID := r.FormValue("update_id")
	parentID := r.FormValue("parent_id")
	content := r.FormValue("content")
	if content == "" {
		http.Error(w, "Comment content cannot be empty", http.StatusBadRequest)
		return
	}

	userID := h.currentUserID(r)

	var err error
	var commentID int64
	if parentID == "" {
		err = h.db.QueryRow("INSERT INTO comments (update_id, user_id, content) VALUES ($1, $2, $3) RETURNING id",
			updateID, userID, content).Scan(&commentID)
	} else {
		err = h.db.QueryRow("INSERT INTO comments (update_id, user_id, parent_id, content) VALUES ($1, $2, $3, $4) RETURNING id",
			updateID, userID, parentID, content).Scan(&commentID)
	}
	if err != nil {
		log.Printf("Failed to add comment: %v", err)
		http.Error(w, "Failed to add comment", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, struct {
		ID       int64  `json:"id"`
		UpdateID string `json:"update_id"`
	}{
		ID:       commentID,
		UpdateID: updateID,
	})
}

type UpdatePageData struct {
	Update          domain.AspirationUpdate
	Comments        []*domain.Comment
	TotalComments   int
	IsAuthenticated bool
}

func (h *Handler) currentUserID(r *http.Request) int {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		return 0
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		log.Printf("Error getting user ID: %v", err)
		return 0
	}
	return userID
}

func (h *Handler) getComments(updateID string) ([]*domain.Comment, error) {
	rows, err := h.db.Query(`
        SELECT c.id, c.user_id, c.parent_id, c.content, c.created_at, 
               u.username, COALESCE(u.display_name, ''), COALESCE(u.profile_image_url, '')
        FROM comments c
        JOIN users u ON c.user_id = u.id
        WHERE c.update_id = $1
        ORDER BY c.created_at ASC
    `, updateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allComments []*domain.Comment
	commentMap := make(map[int]*domain.Comment)

	for rows.Next() {
		var comment domain.Comment
		err := rows.Scan(&comment.ID, &comment.UserID, &comment.ParentID, &comment.Content, &comment.CreatedAt,
			&comment.Username, &comment.DisplayName, &comment.ProfileImageURL)
		if err != nil {
			return nil, err
		}
		commentMap[comment.ID] = &comment
		allComments = append(allComments, &comment)
	}

	var rootComments []*domain.Comment
	for _, comment := range allComments {
		if !comment.ParentID.Valid {
			rootComments = append(rootComments, comment)
		} else {
			parentID := int(comment.ParentID.Int64)
			if parent, ok := commentMap[parentID]; ok {
				parent.Replies = append(parent.Replies, comment)
			}
		}
	}

	log.Printf("Root comments for update %s: %+v", updateID, rootComments)
	return rootComments, nil
}

func (h *Handler) modifyBannedUserComments(comments []*domain.Comment) {
	for _, comment := range comments {
		var isBanned bool
		err := h.db.QueryRow("SELECT is_banned FROM users WHERE id = $1", comment.UserID).Scan(&isBanned)
		if err != nil {
			log.Printf("Error checking user ban status: %v", err)
			continue
		}
		if isBanned {
			comment.Content = "(user banned)"
		}
		h.modifyBannedUserComments(comment.Replies)
	}
}

func countTotalComments(comments []*domain.Comment) int {
	total := len(comments)
	for _, comment := range comments {
		total += countTotalComments(comment.Replies)
	}
	return total
}

func writeJSON(w http.ResponseWriter, statusCode int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}
