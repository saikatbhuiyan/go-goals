package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeRepository struct {
	createErr error
	created   CreateUserInput
}

func (r *fakeRepository) FindCredentialsByEmail(context.Context, string) (UserCredentials, error) {
	return UserCredentials{}, errors.New("not implemented")
}

func (r *fakeRepository) CreateUser(_ context.Context, input CreateUserInput) error {
	r.created = input
	return r.createErr
}

func TestSignUpValidatesRequiredFields(t *testing.T) {
	service := NewService(&fakeRepository{}, PasswordHasher{})

	_, err := service.SignUp(context.Background(), SignUpInput{})
	if !errors.Is(err, ErrInvalidSignup) {
		t.Fatalf("expected ErrInvalidSignup, got %v", err)
	}
}

func TestSignUpValidatesPasswordMatch(t *testing.T) {
	service := NewService(&fakeRepository{}, PasswordHasher{})

	_, err := service.SignUp(context.Background(), SignUpInput{
		Email:           "person@example.com",
		Username:        "person",
		Password:        "password-one",
		ConfirmPassword: "password-two",
	})
	if !errors.Is(err, ErrPasswordMismatch) {
		t.Fatalf("expected ErrPasswordMismatch, got %v", err)
	}
}

func TestSignUpNormalizesIdentityBeforeCreatingUser(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, PasswordHasher{})

	email, err := service.SignUp(context.Background(), SignUpInput{
		Email:           " Person@Example.COM ",
		Username:        " Person_1 ",
		DisplayName:     "",
		Password:        "strong-password",
		ConfirmPassword: "strong-password",
	})
	if err != nil {
		t.Fatalf("expected signup to succeed, got %v", err)
	}
	if email != "person@example.com" {
		t.Fatalf("expected normalized email, got %q", email)
	}
	if repository.created.Username != "person_1" {
		t.Fatalf("expected normalized username, got %q", repository.created.Username)
	}
	if repository.created.DisplayName != "person_1" {
		t.Fatalf("expected default display name, got %q", repository.created.DisplayName)
	}
	if repository.created.PasswordHash == "" {
		t.Fatal("expected password hash to be stored")
	}
}

func TestSignUpMapsDuplicateIdentity(t *testing.T) {
	repository := &fakeRepository{
		createErr: errors.New("pq: duplicate key value violates unique constraint"),
	}
	service := NewService(repository, PasswordHasher{})

	_, err := service.SignUp(context.Background(), SignUpInput{
		Email:           "person@example.com",
		Username:        "person",
		Password:        "strong-password",
		ConfirmPassword: "strong-password",
	})
	if !errors.Is(err, ErrDuplicateIdentity) {
		t.Fatalf("expected ErrDuplicateIdentity, got %v", err)
	}
}
