package usecase

import "context"

type Article struct {
	ID      int64
	Name    string
	Content string
}

type ArticleRepository interface {
	Create(ctx context.Context, in *Article) (*Article, error)
	Get(ctx context.Context, id int64) (*Article, error)
	List(ctx context.Context) ([]*Article, error)
	Update(ctx context.Context, in *Article) (*Article, error)
	Delete(ctx context.Context, id int64) error
}

type ArticleUsecase struct {
	repo ArticleRepository
}

func NewArticleUsecase(repo ArticleRepository) *ArticleUsecase {
	return &ArticleUsecase{repo: repo}
}

func (u *ArticleUsecase) Create(ctx context.Context, in *Article) (*Article, error) {
	return u.repo.Create(ctx, in)
}
func (u *ArticleUsecase) Get(ctx context.Context, id int64) (*Article, error) {
	return u.repo.Get(ctx, id)
}
func (u *ArticleUsecase) List(ctx context.Context) ([]*Article, error) {
	return u.repo.List(ctx)
}
func (u *ArticleUsecase) Update(ctx context.Context, in *Article) (*Article, error) {
	return u.repo.Update(ctx, in)
}
func (u *ArticleUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}
