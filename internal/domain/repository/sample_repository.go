package repository

import (
	"context"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
)

// SampleRepository は Sample 集約の永続化境界を表すドメイン側のポートです。
// アダプタ層（mysql/memory 等）はこのインターフェースを実装します。
type SampleRepository interface {
	Create(ctx context.Context, in *entity.Sample) (*entity.Sample, error)
	Get(ctx context.Context, id int64) (*entity.Sample, error)
	List(ctx context.Context, p domain.ListParams) ([]*entity.Sample, error)
	Update(ctx context.Context, in *entity.Sample) (*entity.Sample, error)
	Delete(ctx context.Context, id int64) error
}
