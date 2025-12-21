package auth

import (
	"context"
)

type Principal struct {
	UserID int64
	Email  string
	Roles  []string
}

type ctxKey int

const principalKey ctxKey = 1

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func FromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey).(*Principal)
	return p, ok
}
