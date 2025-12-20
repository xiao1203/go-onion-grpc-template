package mysql

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

type ArticleModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;not null"`
	Content   string    `gorm:"column:content;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ArticleModel) TableName() string { return "articles" }

type ArticleRepository struct{ db *gorm.DB }

func NewArticleRepository(db *gorm.DB) *ArticleRepository { return &ArticleRepository{db: db} }

func (r *ArticleRepository) Create(ctx context.Context, in *usecase.Article) (*usecase.Article, error) {
	m := ArticleModel{
		Name:    in.Name,
		Content: in.Content,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	out := *in
	out.ID = m.ID
	return &out, nil
}

func (r *ArticleRepository) Get(ctx context.Context, id int64) (*usecase.Article, error) {
	var m ArticleModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &usecase.Article{
		ID:      m.ID,
		Name:    m.Name,
		Content: m.Content,
	}, nil
}

func (r *ArticleRepository) List(ctx context.Context) ([]*usecase.Article, error) {
	var rows []ArticleModel
	if err := r.db.WithContext(ctx).Order("id DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*usecase.Article, 0, len(rows))
	for _, m := range rows {
		it := usecase.Article{
			ID:      m.ID,
			Name:    m.Name,
			Content: m.Content,
		}
		out = append(out, &it)
	}
	return out, nil
}

func (r *ArticleRepository) Update(ctx context.Context, in *usecase.Article) (*usecase.Article, error) {
	updates := map[string]interface{}{
		"name":       in.Name,
		"content":    in.Content,
		"updated_at": time.Now(),
	}
	if err := r.db.WithContext(ctx).Model(&ArticleModel{}).Where("id = ?", in.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, in.ID)
}

func (r *ArticleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&ArticleModel{}, id).Error
}
