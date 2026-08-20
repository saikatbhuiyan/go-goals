package httpx

import (
	"context"
)

type contextKey string

const emailKey contextKey = "email"

func ContextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailKey, email)
}

func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailKey).(string)
	return email, ok && email != ""
}
