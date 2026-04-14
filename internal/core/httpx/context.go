package httpx

import "context"

type contextKey string

const (
	ctxDomainKey contextKey = "request_domain"
	ctxUserIDKey contextKey = "request_user_id"
)

func WithDomain(ctx context.Context, domain string) context.Context {
	return context.WithValue(ctx, ctxDomainKey, domain)
}

func DomainFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxDomainKey).(string); ok {
		return v
	}
	return ""
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserIDKey, userID)
}

func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxUserIDKey).(string); ok {
		return v
	}
	return ""
}
