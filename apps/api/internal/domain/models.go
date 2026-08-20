package domain

import (
	"database/sql"
	"time"
)

type User struct {
	ID              int
	Email           string
	Username        string
	DisplayName     string
	LifeAspirations sql.NullString
	ThingsILikeToDo sql.NullString
	ProfileImageURL sql.NullString
	Bio             sql.NullString
	BioLink         sql.NullString
	IsBanned        bool
}

type AspirationUpdate struct {
	ID              int
	Username        string
	DisplayName     string
	Content         string
	CreatedAt       time.Time
	LikeCount       int
	CommentCount    int
	Liked           bool
	IsOwnPost       bool
	ProfileImageURL sql.NullString
}

type Comment struct {
	ID              int
	UpdateID        int
	UserID          int
	ParentID        sql.NullInt64
	Content         string
	CreatedAt       time.Time
	Username        string
	DisplayName     string
	ProfileImageURL string
	Replies         []*Comment
}

type CommentContext struct {
	Root            interface{}
	Comment         *Comment
	UpdateID        int
	IsAuthenticated bool
}

type Administrator struct {
	ID       int
	Email    string
	Username string
}
