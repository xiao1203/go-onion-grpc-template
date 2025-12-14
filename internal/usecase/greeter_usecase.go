package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
)

type GreeterUsecase struct {
	repo GreeterRepository
}

func NewGreeterUsecase(repo GreeterRepository) *GreeterUsecase {
	return &GreeterUsecase{repo: repo}
}

func (u *GreeterUsecase) Hello(ctx context.Context, name string) (*domain.Greeting, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name is required")
	}
	return u.repo.BuildGreeting(ctx, name)
}
