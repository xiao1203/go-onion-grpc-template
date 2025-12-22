package mysql

// scaffold実行のサンプルコードです。不要であれば削除してください。
import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
	domainrepo "github.com/xiao1203/go-onion-grpc-template/internal/domain/repository"
)

type SampleModel struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;not null"`
	Content   string    `gorm:"column:content;not null"`
	Count     uint32    `gorm:"column:count;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SampleModel) TableName() string { return "samples" }

type SampleRepository struct{ db *gorm.DB }

func NewSampleRepository(db *gorm.DB) domainrepo.SampleRepository { return &SampleRepository{db: db} }

func (r *SampleRepository) Create(ctx context.Context, in *entity.Sample) (*entity.Sample, error) {
	m := SampleModel{
		Name:    in.Name,
		Content: in.Content,
		Count:   in.Count,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	out := *in
	out.ID = m.ID
	return &out, nil
}

func (r *SampleRepository) Get(ctx context.Context, id int64) (*entity.Sample, error) {
	var m SampleModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity.Sample{
		ID:      m.ID,
		Name:    m.Name,
		Content: m.Content,
		Count:   m.Count,
	}, nil
}

func (r *SampleRepository) List(ctx context.Context, p domain.ListParams) ([]*entity.Sample, error) {
	var rows []SampleModel
	p = p.Sanitize()
	q := r.db.WithContext(ctx).Order("id DESC").Offset(p.Offset).Limit(p.Limit)
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*entity.Sample, 0, len(rows))
	for _, m := range rows {
		it := entity.Sample{
			ID:      m.ID,
			Name:    m.Name,
			Content: m.Content,
			Count:   m.Count,
		}
		out = append(out, &it)
	}
	return out, nil
}

func (r *SampleRepository) Update(ctx context.Context, in *entity.Sample) (*entity.Sample, error) {
	updates := map[string]interface{}{
		"name":       in.Name,
		"content":    in.Content,
		"count":      in.Count,
		"updated_at": time.Now(),
	}
	if err := r.db.WithContext(ctx).Model(&SampleModel{}).Where("id = ?", in.ID).Updates(updates).Error; err != nil {
		return nil, err
	}
	return r.Get(ctx, in.ID)
}

func (r *SampleRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&SampleModel{}, id).Error
}
