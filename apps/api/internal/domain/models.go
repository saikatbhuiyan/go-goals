package domain

import (
	"database/sql"
	"time"
)

type User struct {
	ID              int            `json:"id"`
	Email           string         `json:"email"`
	Username        string         `json:"username"`
	DisplayName     string         `json:"display_name"`
	LifeAspirations sql.NullString `json:"life_aspirations"`
	ThingsILikeToDo sql.NullString `json:"things_i_like_to_do"`
	ProfileImageURL sql.NullString `json:"profile_image_url"`
	Bio             sql.NullString `json:"bio"`
	BioLink         sql.NullString `json:"bio_link"`
	IsBanned        bool           `json:"is_banned"`
}

type AspirationUpdate struct {
	ID              int            `json:"id"`
	Username        string         `json:"username"`
	DisplayName     string         `json:"display_name"`
	Content         string         `json:"content"`
	CreatedAt       time.Time      `json:"created_at"`
	LikeCount       int            `json:"like_count"`
	CommentCount    int            `json:"comment_count"`
	Liked           bool           `json:"liked"`
	IsOwnPost       bool           `json:"is_own_post"`
	ProfileImageURL sql.NullString `json:"profile_image_url"`
}

type Comment struct {
	ID              int           `json:"id"`
	UpdateID        int           `json:"update_id"`
	UserID          int           `json:"user_id"`
	ParentID        sql.NullInt64 `json:"parent_id"`
	Content         string        `json:"content"`
	CreatedAt       time.Time     `json:"created_at"`
	Username        string        `json:"username"`
	DisplayName     string        `json:"display_name"`
	ProfileImageURL string        `json:"profile_image_url"`
	Replies         []*Comment    `json:"replies"`
}

type CommentContext struct {
	Root            interface{} `json:"-"`
	Comment         *Comment    `json:"comment"`
	UpdateID        int         `json:"update_id"`
	IsAuthenticated bool        `json:"is_authenticated"`
}

type Administrator struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}
