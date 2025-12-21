package usecase

import (
	"context"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
)

type SampleRepository interface {
	Create(ctx context.Context, in *domain.Sample) (*domain.Sample, error)
	Get(ctx context.Context, id int64) (*domain.Sample, error)
	List(ctx context.Context, p domain.ListParams) ([]*domain.Sample, error)
	Update(ctx context.Context, in *domain.Sample) (*domain.Sample, error)
	Delete(ctx context.Context, id int64) error
}

type SampleUsecase struct {
	repo SampleRepository
}

func NewSampleUsecase(repo SampleRepository) *SampleUsecase {
	return &SampleUsecase{repo: repo}
}

func (u *SampleUsecase) Create(ctx context.Context, in *domain.Sample) (*domain.Sample, error) {
	return u.repo.Create(ctx, in)
}
func (u *SampleUsecase) Get(ctx context.Context, id int64) (*domain.Sample, error) {
	return u.repo.Get(ctx, id)
}
func (u *SampleUsecase) List(ctx context.Context, p domain.ListParams) ([]*domain.Sample, error) {
	return u.repo.List(ctx, p)
}
func (u *SampleUsecase) Update(ctx context.Context, in *domain.Sample) (*domain.Sample, error) {
	return u.repo.Update(ctx, in)
}
func (u *SampleUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
