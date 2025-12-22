package repository

import (
	"context"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
)

// UserRepository は User 集約の永続化境界を表すドメイン側のポートです。
type UserRepository interface {
	FindByID(ctx context.Context, id int64) (*entity.User, error)
	UpdateProfile(ctx context.Context, id int64, displayName, pictureURL string) (*entity.User, error)
}
