package users

import (
	"database/sql"
	"fmt"
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

type updateProfileRequest struct {
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	LifeAspirations string `json:"life_aspirations"`
	ThingsILikeToDo string `json:"things_i_like_to_do"`
	Bio             string `json:"bio"`
	BioLink         string `json:"bio_link"`
}

type moderationRequest struct {
	UserID int `json:"user_id"`
}

type followRequest struct {
	Username string `json:"username"`
}

func NewHandler(db *sql.DB, store *gorillasessions.CookieStore) *Handler {
	return &Handler{db: db, store: store}
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, (&domain.DatabaseError{Operation: "fetching user ID", Err: err}).Error())
		return
	}

	page := pageFromRequest(r)
	profileData, err := h.getProfileData(userID, userID, page, 5)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, (&domain.DatabaseError{Operation: "fetching profile data", Err: err}).Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, profileData)
}

func (h *Handler) ProfileEdit(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		h.updateProfile(w, r, email)
		return
	}
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var user domain.User
	err := h.db.QueryRow("SELECT email, COALESCE(username, ''), COALESCE(display_name, ''), life_aspirations, things_i_like_to_do, profile_image_url, bio, bio_link FROM users WHERE email = $1", email).Scan(&user.Email, &user.Username, &user.DisplayName, &user.LifeAspirations, &user.ThingsILikeToDo, &user.ProfileImageURL, &user.Bio, &user.BioLink)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusOK, user)
}

func (h *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	username := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if username == "" {
		httpx.WriteError(w, http.StatusNotFound, "User not found")
		return
	}

	var userID int
	var isBanned bool
	err := h.db.QueryRow("SELECT id, is_banned FROM users WHERE username = $1", username).Scan(&userID, &isBanned)
	if err == sql.ErrNoRows {
		log.Printf("User not found: %s", username)
		httpx.WriteError(w, http.StatusNotFound, "User not found")
		return
	} else if err != nil {
		log.Printf("Database error when fetching user %s: %v", username, err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	session, _ := h.store.Get(r, "session-name")
	currentUserEmail, _ := session.Values["email"].(string)
	isLoggedIn := currentUserEmail != ""

	isAdmin := false
	if isLoggedIn {
		err = h.db.QueryRow("SELECT COUNT(*) > 0 FROM administrators WHERE email = $1", currentUserEmail).Scan(&isAdmin)
		if err != nil {
			log.Printf("Error checking admin status: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	if isBanned && !isAdmin {
		httpx.WriteError(w, http.StatusForbidden, "This user's profile is not available")
		return
	}

	currentUserID := 0
	if isLoggedIn {
		err = h.db.QueryRow("SELECT id FROM users WHERE email = $1", currentUserEmail).Scan(&currentUserID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Database error when fetching current user ID: %v", err)
			httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
	}

	profileData, err := h.getProfileData(userID, currentUserID, pageFromRequest(r), 5)
	if err != nil {
		log.Printf("Error fetching profile data: %v", err)
		httpx.WriteError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	data := struct {
		ProfileData
		IsOwnProfile bool
		IsLoggedIn   bool
		IsAdmin      bool
		IsBanned     bool
	}{
		ProfileData:  profileData,
		IsOwnProfile: profileData.User.Email == currentUserEmail,
		IsLoggedIn:   isLoggedIn,
		IsAdmin:      isAdmin,
		IsBanned:     isBanned,
	}

	httpx.WriteJSON(w, http.StatusOK, data)
}

func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request moderationRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}
	if request.UserID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	_, err := h.db.Exec("UPDATE users SET is_banned = true WHERE id = $1", request.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to ban user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var request moderationRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}
	if request.UserID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	_, err := h.db.Exec("UPDATE users SET is_banned = false WHERE id = $1", request.UserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to unban user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var request followRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	username := strings.TrimSpace(request.Username)
	if username == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Missing username")
		return
	}

	var followerID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&followerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get follower ID: "+err.Error())
		return
	}

	var followedID int
	err = h.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&followedID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get followed ID: "+err.Error())
		return
	}

	if followerID == followedID {
		httpx.WriteError(w, http.StatusBadRequest, "You cannot follow yourself")
		return
	}

	_, err = h.db.Exec("INSERT INTO followers (follower_id, followed_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", followerID, followedID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to follow user: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var request followRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	username := strings.TrimSpace(request.Username)
	if username == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Missing username")
		return
	}

	var followerID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&followerID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get follower ID: "+err.Error())
		return
	}

	var followedID int
	err = h.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&followedID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to get followed ID: "+err.Error())
		return
	}

	_, err = h.db.Exec("DELETE FROM followers WHERE follower_id = $1 AND followed_id = $2", followerID, followedID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to unfollow user: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request, email string) {
	var request updateProfileRequest
	if err := httpx.ReadJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "Invalid JSON request body")
		return
	}

	username := strings.ToLower(strings.TrimSpace(request.Username))
	displayName := strings.TrimSpace(request.DisplayName)
	lifeAspirations := strings.TrimSpace(request.LifeAspirations)
	thingsILikeToDo := strings.TrimSpace(request.ThingsILikeToDo)
	bio := strings.TrimSpace(request.Bio)
	bioLink := strings.TrimSpace(request.BioLink)

	if username == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Username cannot be empty")
		return
	}

	var currentUsername sql.NullString
	err := h.db.QueryRow("SELECT username FROM users WHERE email = $1", email).Scan(&currentUsername)
	if err != nil && err != sql.ErrNoRows {
		httpx.WriteError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	if username != currentUsername.String {
		var count int
		err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 AND email != $2", username, email).Scan(&count)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Database error: "+err.Error())
			return
		}
		if count > 0 {
			httpx.WriteError(w, http.StatusConflict, "Username is already taken")
			return
		}
	}

	_, err = h.db.Exec("UPDATE users SET username = $1, display_name = $2, life_aspirations = $3, things_i_like_to_do = $4, bio = $5, bio_link = $6 WHERE email = $7",
		username,
		displayName,
		sql.NullString{String: lifeAspirations, Valid: lifeAspirations != ""},
		sql.NullString{String: thingsILikeToDo, Valid: thingsILikeToDo != ""},
		sql.NullString{String: bio, Valid: bio != ""},
		sql.NullString{String: bioLink, Valid: bioLink != ""},
		email)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Failed to update user: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) getProfileData(userID int, currentUserID int, page int, pageSize int) (ProfileData, error) {
	offset := (page - 1) * pageSize

	var user domain.User
	err := h.db.QueryRow(`
		SELECT id, email, COALESCE(username, ''), COALESCE(display_name, ''), 
		life_aspirations, things_i_like_to_do, profile_image_url, bio, bio_link, is_banned 
		FROM users WHERE id = $1`, userID).Scan(
		&user.ID, &user.Email, &user.Username, &user.DisplayName, &user.LifeAspirations,
		&user.ThingsILikeToDo, &user.ProfileImageURL, &user.Bio, &user.BioLink, &user.IsBanned)
	if err != nil {
		return ProfileData{}, err
	}

	rows, err := h.db.Query(`
		SELECT au.id, au.content, au.created_at, COUNT(DISTINCT l.id) as like_count,
			   COUNT(DISTINCT c.id) as comment_count,
			   CASE WHEN EXISTS (SELECT 1 FROM likes WHERE user_id = $1 AND update_id = au.id) THEN TRUE ELSE FALSE END as liked,
			   au.user_id = $1 as is_own_post,
			   u.profile_image_url
		FROM aspiration_updates au
		JOIN users u ON au.user_id = u.id
		LEFT JOIN likes l ON au.id = l.update_id
		LEFT JOIN comments c ON au.id = c.update_id
		WHERE au.user_id = $2
		GROUP BY au.id, au.content, au.created_at, au.user_id, u.profile_image_url
		ORDER BY au.created_at DESC
		LIMIT $3 OFFSET $4
	`, currentUserID, userID, pageSize, offset)
	if err != nil {
		return ProfileData{}, err
	}
	defer rows.Close()

	var updates []ProfileUpdate
	for rows.Next() {
		var update domain.AspirationUpdate
		var isOwnPost bool
		if err := rows.Scan(&update.ID, &update.Content, &update.CreatedAt, &update.LikeCount, &update.CommentCount, &update.Liked, &isOwnPost, &update.ProfileImageURL); err != nil {
			return ProfileData{}, err
		}
		updates = append(updates, ProfileUpdate{AspirationUpdate: update, IsOwnPost: isOwnPost})
	}

	var isFollowing bool
	if currentUserID != 0 {
		err = h.db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM followers 
				WHERE follower_id = $1 AND followed_id = $2
			)
		`, currentUserID, userID).Scan(&isFollowing)
		if err != nil {
			return ProfileData{}, err
		}
	}

	var followerCount int
	err = h.db.QueryRow("SELECT COUNT(*) FROM followers WHERE followed_id = $1", userID).Scan(&followerCount)
	if err != nil {
		return ProfileData{}, err
	}

	rows, err = h.db.Query(`
		SELECT u.username, COALESCE(u.profile_image_url, '') as profile_image_url
		FROM followers f
		JOIN users u ON f.follower_id = u.id
		WHERE f.followed_id = $1
		ORDER BY f.created_at DESC
		LIMIT 10
	`, userID)
	if err != nil {
		return ProfileData{}, err
	}
	defer rows.Close()

	var recentFollowers []RecentFollower
	for rows.Next() {
		var follower RecentFollower
		if err := rows.Scan(&follower.Username, &follower.ProfileImageURL); err != nil {
			return ProfileData{}, err
		}
		recentFollowers = append(recentFollowers, follower)
	}

	var totalUpdates int
	err = h.db.QueryRow("SELECT COUNT(*) FROM aspiration_updates WHERE user_id = $1", userID).Scan(&totalUpdates)
	if err != nil {
		return ProfileData{}, err
	}
	totalPages := (totalUpdates + pageSize - 1) / pageSize

	return ProfileData{
		User:            user,
		Updates:         updates,
		IsFollowing:     isFollowing,
		FollowerCount:   followerCount,
		RecentFollowers: recentFollowers,
		CurrentPage:     page,
		TotalPages:      totalPages,
		PreviousPage:    page - 1,
		NextPage:        page + 1,
	}, nil
}

type ProfileData struct {
	User            domain.User      `json:"user"`
	Updates         []ProfileUpdate  `json:"updates"`
	IsFollowing     bool             `json:"is_following"`
	FollowerCount   int              `json:"follower_count"`
	RecentFollowers []RecentFollower `json:"recent_followers"`
	CurrentPage     int              `json:"current_page"`
	TotalPages      int              `json:"total_pages"`
	PreviousPage    int              `json:"previous_page"`
	NextPage        int              `json:"next_page"`
}

type ProfileUpdate struct {
	domain.AspirationUpdate
	IsOwnPost bool `json:"is_own_post"`
}

type RecentFollower struct {
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url"`
}

func pageFromRequest(r *http.Request) int {
	page := 1
	if pageParam := r.URL.Query().Get("page"); pageParam != "" {
		if _, err := fmt.Sscanf(pageParam, "%d", &page); err != nil || page < 1 {
			page = 1
		}
	}
	return page
}
