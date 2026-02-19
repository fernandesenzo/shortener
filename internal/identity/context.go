package identity

import "context"

type contextKey string

const userIDKey contextKey = "userID"

func GetUserID(ctx context.Context) (string, bool) {
	ctxValue := ctx.Value(userIDKey)
	id, ok := ctxValue.(string)
	return id, ok
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
