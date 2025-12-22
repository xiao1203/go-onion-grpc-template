package usecase

import (
	"context"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
	domainrepo "github.com/xiao1203/go-onion-grpc-template/internal/domain/repository"
)

type UserUsecase struct {
	repo domainrepo.UserRepository
}

func NewUserUsecase(repo domainrepo.UserRepository) *UserUsecase { return &UserUsecase{repo: repo} }

func (u *UserUsecase) GetMe(ctx context.Context, id int64) (*entity.User, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *UserUsecase) UpdateMyProfile(ctx context.Context, id int64, displayName, pictureURL string) (*entity.User, error) {
	return u.repo.UpdateProfile(ctx, id, displayName, pictureURL)
}
