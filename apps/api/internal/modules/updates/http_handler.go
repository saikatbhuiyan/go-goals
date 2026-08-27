package updates

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	gorillasessions "github.com/gorilla/sessions"
	"github.com/saikatbhuiyan/go-goals/internal/domain"
	"github.com/saikatbhuiyan/go-goals/internal/platform/httpx"
)

type Handler struct {
	db    *sql.DB
	store *gorillasessions.CookieStore
}

type upsertUpdateRequest struct {
	Content  string `json:"content"`
	UpdateID string `json:"update_id"`
}

type updateContentRequest struct {
	Content string `json:"content"`
}

type updateReactionRequest struct {
	UpdateID string `json:"update_id"`
}

type addCommentRequest struct {
	UpdateID string `json:"update_id"`
	ParentID string `json:"parent_id"`
	Content  string `json:"content"`
}

type recentUser struct {
	ID              int    `json:"id"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
}

type feedResponse struct {
	Users   []recentUser              `json:"users"`
	Updates []domain.AspirationUpdate `json:"updates"`
}

func NewHandler(db *sql.DB, store *gorillasessions.CookieStore) *Handler {
	return &Handler{db: db, store: store}
}

func (h *Handler) Browse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	users, err := h.fetchRecentUsers()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch recent users")
		return
	}

	updates, err := h.fetchRecentUpdates(h.currentSessionUserID(r), 20)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch recent updates")
		return
	}

	httpx.WriteJSON(w, http.StatusOK, feedResponse{
		Users:   users,
		Updates: updates,
	})
}

func (h *Handler) AspirationUpdate(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request upsertUpdateRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	content := strings.TrimSpace(request.Content)
	updateID := strings.TrimSpace(request.UpdateID)
	if content == "" {
		httpx.WriteError(w, http.StatusBadRequest, (&domain.ValidationError{Field: "content", Message: "Content cannot be empty"}).Error())
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, (&domain.DatabaseError{Operation: "fetching user ID", Err: err}).Error())
		return
	}

	if updateID != "" {
		_, err = h.db.Exec("UPDATE aspiration_updates SET content = $1 WHERE id = $2 AND user_id = $3", content, updateID, userID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, (&domain.DatabaseError{Operation: "updating aspiration update", Err: err}).Error())
			return
		}
	} else {
		_, err = h.db.Exec("INSERT INTO aspiration_updates (user_id, content) VALUES ($1, $2)", userID, content)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, (&domain.DatabaseError{Operation: "creating aspiration update", Err: err}).Error())
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) EditAspirationUpdate(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	updateID := strings.TrimPrefix(r.URL.Path, "/api/aspiration-updates/edit/")
	if updateID == "" {
		httpx.WriteError(w, http.StatusNotFound, "Aspiration update not found")
		return
	}

	var update domain.AspirationUpdate
	err := h.db.QueryRow("SELECT id, content FROM aspiration_updates WHERE id = $1 AND user_id = (SELECT id FROM users WHERE email = $2)", updateID, email).Scan(&update.ID, &update.Content)
	if err == sql.ErrNoRows {
		httpx.WriteError(w, http.StatusNotFound, "Aspiration update not found")
		return
	} else if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to fetch aspiration update: "+err.Error())
		return
	}

	if r.Method == http.MethodGet {
		httpx.WriteJSON(w, http.StatusOK, update)
		return
	}
	if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request updateContentRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	content := strings.TrimSpace(request.Content)
	if content == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Content cannot be empty")
		return
	}

	_, err = h.db.Exec("UPDATE aspiration_updates SET content = $1 WHERE id = $2", content, updateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to update aspiration update: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) DeleteAspirationUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	updateID := strings.TrimPrefix(r.URL.Path, "/api/aspiration-updates/delete/")
	if updateID == "" {
		httpx.WriteError(w, http.StatusNotFound, "Aspiration update not found")
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to start transaction: "+err.Error())
		return
	}
	defer tx.Rollback()

	if _, err = tx.Exec("DELETE FROM comments WHERE update_id = $1", updateID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to delete associated comments: "+err.Error())
		return
	}
	if _, err = tx.Exec("DELETE FROM likes WHERE update_id = $1", updateID); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to delete associated likes: "+err.Error())
		return
	}
	if _, err = tx.Exec("DELETE FROM aspiration_updates WHERE id = $1 AND user_id = (SELECT id FROM users WHERE email = $2)", updateID, email); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to delete aspiration update: "+err.Error())
		return
	}
	if err = tx.Commit(); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to commit transaction: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Like(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var request updateReactionRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	updateID := strings.TrimSpace(request.UpdateID)
	if updateID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Missing update_id")
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get user ID: "+err.Error())
		return
	}

	var postOwnerID int
	err = h.db.QueryRow("SELECT user_id FROM aspiration_updates WHERE id = $1", updateID).Scan(&postOwnerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get post owner: "+err.Error())
		return
	}
	if userID == postOwnerID {
		httpx.WriteError(w, http.StatusBadRequest, "You cannot like your own post")
		return
	}

	_, err = h.db.Exec("INSERT INTO likes (user_id, update_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", userID, updateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to add like: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Unlike(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var request updateReactionRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	updateID := strings.TrimSpace(request.UpdateID)
	if updateID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Missing update_id")
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get user ID: "+err.Error())
		return
	}

	_, err = h.db.Exec("DELETE FROM likes WHERE user_id = $1 AND update_id = $2", userID, updateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to remove like: "+err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Permalink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	updateID := strings.TrimPrefix(r.URL.Path, "/api/updates/")
	log.Printf("Handling update permalink for ID: %s", updateID)
	if updateID == "" {
		httpx.WriteError(w, http.StatusNotFound, "Update not found")
		return
	}

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
			httpx.WriteError(w, http.StatusNotFound, "Update not found")
		} else {
			log.Printf("Error fetching update: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}
	if isBanned {
		httpx.WriteError(w, http.StatusForbidden, "This update is not available")
		return
	}

	comments, err := h.getComments(updateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, (&domain.DatabaseError{Operation: "fetching comments", Err: err}).Error())
		return
	}
	h.modifyBannedUserComments(comments)

	data := UpdatePageData{
		Update:          update,
		Comments:        comments,
		TotalComments:   countTotalComments(comments),
		IsAuthenticated: currentUserID != 0,
	}

	httpx.WriteJSON(w, http.StatusOK, data)
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request addCommentRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	updateID := strings.TrimSpace(request.UpdateID)
	parentID := strings.TrimSpace(request.ParentID)
	content := strings.TrimSpace(request.Content)
	if updateID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Missing update_id")
		return
	}
	if content == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Comment content cannot be empty")
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
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to add comment")
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, struct {
		ID       int64  `json:"id"`
		UpdateID string `json:"update_id"`
	}{
		ID:       commentID,
		UpdateID: updateID,
	})
}

type UpdatePageData struct {
	Update          domain.AspirationUpdate `json:"update"`
	Comments        []*domain.Comment       `json:"comments"`
	TotalComments   int                     `json:"total_comments"`
	IsAuthenticated bool                    `json:"is_authenticated"`
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

func (h *Handler) currentSessionUserID(r *http.Request) int {
	session, _ := h.store.Get(r, "session-name")
	email, ok := session.Values["email"].(string)
	if !ok || email == "" {
		return 0
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		log.Printf("Error getting session user ID: %v", err)
		return 0
	}
	return userID
}

func (h *Handler) fetchRecentUsers() ([]recentUser, error) {
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

	var users []recentUser
	for rows.Next() {
		var user recentUser
		if err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.ProfileImageURL); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (h *Handler) fetchRecentUpdates(currentUserID int, limit int) ([]domain.AspirationUpdate, error) {
	query := `
		WITH ranked_updates AS (
			SELECT au.id, u.username, u.display_name, u.profile_image_url, au.content, au.created_at,
				COUNT(DISTINCT l.id) AS like_count,
				COUNT(DISTINCT c.id) AS comment_count,
				CASE WHEN $1 > 0 AND EXISTS (SELECT 1 FROM likes WHERE user_id = $1 AND update_id = au.id) THEN TRUE ELSE FALSE END AS liked,
				CASE WHEN $1 > 0 AND au.user_id = $1 THEN TRUE ELSE FALSE END AS is_own_post,
				ROW_NUMBER() OVER (PARTITION BY au.user_id ORDER BY au.created_at DESC) AS rn
			FROM aspiration_updates au
			JOIN users u ON au.user_id = u.id
			LEFT JOIN likes l ON au.id = l.update_id
			LEFT JOIN comments c ON au.id = c.update_id
			WHERE u.is_banned = false
			GROUP BY au.id, u.username, u.display_name, u.profile_image_url, au.content, au.created_at, au.user_id
		)
		SELECT id, COALESCE(username, ''), COALESCE(display_name, ''), COALESCE(profile_image_url, ''), content, created_at,
			like_count, comment_count, liked, is_own_post
		FROM ranked_updates
		WHERE rn = 1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := h.db.Query(query, currentUserID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updates []domain.AspirationUpdate
	for rows.Next() {
		var update domain.AspirationUpdate
		if err := rows.Scan(
			&update.ID,
			&update.Username,
			&update.DisplayName,
			&update.ProfileImageURL,
			&update.Content,
			&update.CreatedAt,
			&update.LikeCount,
			&update.CommentCount,
			&update.Liked,
			&update.IsOwnPost,
		); err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}
	return updates, rows.Err()
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
