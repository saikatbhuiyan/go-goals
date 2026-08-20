package main

import "github.com/ALT-F4-LLC/fem-fd-service/apps/api/internal/domain"

type contextKey string

const emailKey contextKey = "email"

type NotFoundError = domain.NotFoundError
type ValidationError = domain.ValidationError
type DatabaseError = domain.DatabaseError

type User = domain.User
type AspirationUpdate = domain.AspirationUpdate
type Comment = domain.Comment
type CommentContext = domain.CommentContext
type Administrator = domain.Administrator
