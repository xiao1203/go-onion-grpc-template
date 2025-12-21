package usecase

import "context"

type User struct {
	ID          int64
	Email       string
	DisplayName string
	PictureURL  string
	Roles       []string
}

type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*User, error)
	UpdateProfile(ctx context.Context, id int64, displayName, pictureURL string) (*User, error)
}

type UserUsecase struct {
	repo UserRepository
}

func NewUserUsecase(repo UserRepository) *UserUsecase { return &UserUsecase{repo: repo} }

func (u *UserUsecase) GetMe(ctx context.Context, id int64) (*User, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *UserUsecase) UpdateMyProfile(ctx context.Context, id int64, displayName, pictureURL string) (*User, error) {
	return u.repo.UpdateProfile(ctx, id, displayName, pictureURL)
}
