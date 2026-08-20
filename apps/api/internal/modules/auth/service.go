package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrBannedAccount      = errors.New("account is banned")
	ErrDuplicateIdentity  = errors.New("email or username is already in use")
	ErrInvalidSignup      = errors.New("email, username, and password are required")
	ErrInvalidUsername    = errors.New("username can only use letters, numbers, underscores, and hyphens")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
	ErrPasswordMismatch   = errors.New("passwords do not match")
)

type UserCredentials struct {
	PasswordHash sql.NullString
	IsBanned     bool
}

type Repository interface {
	FindCredentialsByEmail(ctx context.Context, email string) (UserCredentials, error)
	CreateUser(ctx context.Context, input CreateUserInput) error
}

type CreateUserInput struct {
	Email        string
	Username     string
	DisplayName  string
	PasswordHash string
}

type Service struct {
	repository Repository
	hasher     PasswordHasher
}

func NewService(repository Repository, hasher PasswordHasher) *Service {
	return &Service{
		repository: repository,
		hasher:     hasher,
	}
}

type SignUpInput struct {
	Email           string
	Username        string
	DisplayName     string
	Password        string
	ConfirmPassword string
}

func (s *Service) SignIn(ctx context.Context, email string, password string) error {
	email = NormalizeEmail(email)

	credentials, err := s.repository.FindCredentialsByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrInvalidCredentials
		}
		return err
	}

	if credentials.IsBanned {
		return ErrBannedAccount
	}
	if !credentials.PasswordHash.Valid || credentials.PasswordHash.String == "" {
		return ErrInvalidCredentials
	}

	ok, err := s.hasher.Verify(password, credentials.PasswordHash.String)
	if err != nil || !ok {
		return ErrInvalidCredentials
	}

	return nil
}

func (s *Service) SignUp(ctx context.Context, input SignUpInput) (string, error) {
	email := NormalizeEmail(input.Email)
	username := NormalizeUsername(input.Username)
	displayName := strings.TrimSpace(input.DisplayName)

	if email == "" || username == "" || input.Password == "" {
		return email, ErrInvalidSignup
	}
	if !IsValidUsername(username) {
		return email, ErrInvalidUsername
	}
	if len(input.Password) < 8 {
		return email, ErrWeakPassword
	}
	if input.Password != input.ConfirmPassword {
		return email, ErrPasswordMismatch
	}

	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return email, err
	}
	if displayName == "" {
		displayName = username
	}

	err = s.repository.CreateUser(ctx, CreateUserInput{
		Email:        email,
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if IsUniqueViolation(err) {
			return email, ErrDuplicateIdentity
		}
		return email, err
	}

	return email, nil
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func IsValidUsername(username string) bool {
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

func IsUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}
