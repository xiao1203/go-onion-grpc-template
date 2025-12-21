package usecase

import (
	"context"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*domain.User, error)
	UpdateProfile(ctx context.Context, id int64, displayName, pictureURL string) (*domain.User, error)
}

type UserUsecase struct {
	repo UserRepository
}

func NewUserUsecase(repo UserRepository) *UserUsecase { return &UserUsecase{repo: repo} }

func (u *UserUsecase) GetMe(ctx context.Context, id int64) (*domain.User, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *UserUsecase) UpdateMyProfile(ctx context.Context, id int64, displayName, pictureURL string) (*domain.User, error) {
	return u.repo.UpdateProfile(ctx, id, displayName, pictureURL)
}
