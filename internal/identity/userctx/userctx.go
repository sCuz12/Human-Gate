package userctx

import "context"

type User struct {
	UserID string
	Email  string
	Role   string
}

type contextKey string

const userKey contextKey = "authenticated_user"

func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func FromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey).(User)
	return user, ok
}
