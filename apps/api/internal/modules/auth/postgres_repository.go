package auth

import (
	"context"
	"database/sql"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) FindCredentialsByEmail(ctx context.Context, email string) (UserCredentials, error) {
	var credentials UserCredentials
	err := r.db.QueryRowContext(
		ctx,
		"SELECT password_hash, is_banned FROM users WHERE email = $1",
		email,
	).Scan(&credentials.PasswordHash, &credentials.IsBanned)
	if err != nil {
		return UserCredentials{}, err
	}
	return credentials, nil
}

func (r *PostgresRepository) CreateUser(ctx context.Context, input CreateUserInput) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO users (email, username, display_name, password_hash) VALUES ($1, $2, $3, $4)",
		input.Email,
		input.Username,
		input.DisplayName,
		input.PasswordHash,
	)
	return err
}
