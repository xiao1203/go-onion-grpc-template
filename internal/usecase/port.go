package usecase

import (
	"context"

	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
)

type GreeterRepository interface {
	BuildGreeting(ctx context.Context, name string) (*domain.Greeting, error)
}
