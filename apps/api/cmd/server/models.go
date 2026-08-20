package main

import (
	"database/sql"
	"fmt"
	"time"
)

type contextKey string

const emailKey contextKey = "email"

type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found", e.Resource)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Validation error for %s: %s", e.Field, e.Message)
}

type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("Database error during %s: %v", e.Operation, e.Err)
}

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
