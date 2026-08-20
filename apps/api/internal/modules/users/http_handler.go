package users

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/domain"
	"github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/platform/httpx"
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

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var userID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		http.Error(w, (&domain.DatabaseError{Operation: "fetching user ID", Err: err}).Error(), http.StatusInternalServerError)
		return
	}

	page := pageFromRequest(r)
	profileData, err := h.getProfileData(userID, userID, page, 5)
	if err != nil {
		http.Error(w, (&domain.DatabaseError{Operation: "fetching profile data", Err: err}).Error(), http.StatusInternalServerError)
		return
	}

	if profileData.User.Username == "" {
		http.Redirect(w, r, "/profile/edit", http.StatusSeeOther)
		return
	}

	tmpl, err := template.ParseFiles(h.templatePath("profile.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, profileData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ProfileEdit(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		h.updateProfile(w, r, email)
		return
	}

	var user domain.User
	err := h.db.QueryRow("SELECT email, COALESCE(username, ''), COALESCE(display_name, ''), life_aspirations, things_i_like_to_do, profile_image_url, bio, bio_link FROM users WHERE email = $1", email).Scan(&user.Email, &user.Username, &user.DisplayName, &user.LifeAspirations, &user.ThingsILikeToDo, &user.ProfileImageURL, &user.Bio, &user.BioLink)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.renderProfileEditPage(w, user, "")
}

func (h *Handler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimPrefix(r.URL.Path, "/users/")
	if username == "" {
		http.NotFound(w, r)
		return
	}

	var userID int
	var isBanned bool
	err := h.db.QueryRow("SELECT id, is_banned FROM users WHERE username = $1", username).Scan(&userID, &isBanned)
	if err == sql.ErrNoRows {
		log.Printf("User not found: %s", username)
		http.NotFound(w, r)
		return
	} else if err != nil {
		log.Printf("Database error when fetching user %s: %v", username, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if isBanned && !isAdmin {
		http.Error(w, "This user's profile is not available", http.StatusForbidden)
		return
	}

	currentUserID := 0
	if isLoggedIn {
		err = h.db.QueryRow("SELECT id FROM users WHERE email = $1", currentUserEmail).Scan(&currentUserID)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Database error when fetching current user ID: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	profileData, err := h.getProfileData(userID, currentUserID, pageFromRequest(r), 5)
	if err != nil {
		log.Printf("Error fetching profile data: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(h.templatePath("public_profile.html"))
	if err != nil {
		log.Printf("Template parsing error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) BanUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("UPDATE users SET is_banned = true WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to ban user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users/"+r.FormValue("username"), http.StatusSeeOther)
}

func (h *Handler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := strconv.Atoi(r.FormValue("user_id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("UPDATE users SET is_banned = false WHERE id = $1", userID)
	if err != nil {
		http.Error(w, "Failed to unban user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/users/"+r.FormValue("username"), http.StatusSeeOther)
}

func (h *Handler) Follow(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	var followerID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&followerID)
	if err != nil {
		http.Error(w, "Failed to get follower ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var followedID int
	err = h.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&followedID)
	if err != nil {
		http.Error(w, "Failed to get followed ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if followerID == followedID {
		http.Error(w, "You cannot follow yourself", http.StatusBadRequest)
		return
	}

	_, err = h.db.Exec("INSERT INTO followers (follower_id, followed_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", followerID, followedID)
	if err != nil {
		http.Error(w, "Failed to follow user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) Unfollow(w http.ResponseWriter, r *http.Request) {
	email, ok := httpx.EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := r.FormValue("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	var followerID int
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&followerID)
	if err != nil {
		http.Error(w, "Failed to get follower ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var followedID int
	err = h.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&followedID)
	if err != nil {
		http.Error(w, "Failed to get followed ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = h.db.Exec("DELETE FROM followers WHERE follower_id = $1 AND followed_id = $2", followerID, followedID)
	if err != nil {
		http.Error(w, "Failed to unfollow user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request, email string) {
	username := strings.ToLower(r.FormValue("username"))
	displayName := r.FormValue("display_name")
	lifeAspirations := r.FormValue("life_aspirations")
	thingsILikeToDo := r.FormValue("things_i_like_to_do")
	bio := r.FormValue("bio")
	bioLink := r.FormValue("bio_link")

	if username == "" {
		h.renderProfileEditPage(w, domain.User{Email: email}, "Username cannot be empty")
		return
	}

	var currentUsername sql.NullString
	err := h.db.QueryRow("SELECT username FROM users WHERE email = $1", email).Scan(&currentUsername)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if username != currentUsername.String {
		var count int
		err := h.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 AND email != $2", username, email).Scan(&count)
		if err != nil {
			http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if count > 0 {
			h.renderProfileEditPage(w, domain.User{Email: email, Username: username}, "Username is already taken")
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
		http.Error(w, "Failed to update user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (h *Handler) renderProfileEditPage(w http.ResponseWriter, user domain.User, errorMessage string) {
	tmpl, err := template.ParseFiles(h.templatePath("profile_edit.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		User         domain.User
		ErrorMessage string
	}{
		User:         user,
		ErrorMessage: errorMessage,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	var updates []struct {
		domain.AspirationUpdate
		IsOwnPost bool
	}
	for rows.Next() {
		var update domain.AspirationUpdate
		var isOwnPost bool
		if err := rows.Scan(&update.ID, &update.Content, &update.CreatedAt, &update.LikeCount, &update.CommentCount, &update.Liked, &isOwnPost, &update.ProfileImageURL); err != nil {
			return ProfileData{}, err
		}
		updates = append(updates, struct {
			domain.AspirationUpdate
			IsOwnPost bool
		}{update, isOwnPost})
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
	User    domain.User
	Updates []struct {
		domain.AspirationUpdate
		IsOwnPost bool
	}
	IsFollowing     bool
	FollowerCount   int
	RecentFollowers []RecentFollower
	CurrentPage     int
	TotalPages      int
	PreviousPage    int
	NextPage        int
}

type RecentFollower struct {
	Username        string
	ProfileImageURL string
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
