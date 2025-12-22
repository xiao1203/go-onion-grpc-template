package usecase

import (
	"context"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
	domainrepo "github.com/xiao1203/go-onion-grpc-template/internal/domain/repository"
)

type SampleUsecase struct {
	repo domainrepo.SampleRepository
}

func NewSampleUsecase(repo domainrepo.SampleRepository) *SampleUsecase {
	return &SampleUsecase{repo: repo}
}

func (u *SampleUsecase) Create(ctx context.Context, in *entity.Sample) (*entity.Sample, error) {
	return u.repo.Create(ctx, in)
}
func (u *SampleUsecase) Get(ctx context.Context, id int64) (*entity.Sample, error) {
	return u.repo.Get(ctx, id)
}
func (u *SampleUsecase) List(ctx context.Context, p domain.ListParams) ([]*entity.Sample, error) {
	return u.repo.List(ctx, p)
}
func (u *SampleUsecase) Update(ctx context.Context, in *entity.Sample) (*entity.Sample, error) {
	return u.repo.Update(ctx, in)
}
func (u *SampleUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
