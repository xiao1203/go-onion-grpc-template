package mysql_test

import (
	"context"
	"testing"

	"github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/mysql"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
	"github.com/xiao1203/go-onion-grpc-template/util/testhelper"
	"gorm.io/gorm"
)

func TestArticleRepository_Create(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testDB := testhelper.OpenGormTestDB(t)

	ctx := context.Background()

	tests := []struct {
		name    string
		db      *gorm.DB
		in      *usecase.Article
		want    *usecase.Article
		wantErr bool
	}{
		{
			name: "create succeeds and persists to mysql_test",
			db:   testDB,
			in:   &usecase.Article{Name: "sample", Content: "content"},
			want: &usecase.Article{Name: "sample", Content: "content"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mysql.NewArticleRepository(tt.db)
			got, gotErr := r.Create(ctx, tt.in)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Create() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Create() succeeded unexpectedly")
			}
			if got == nil || got.ID == 0 {
				t.Fatalf("Create() returned invalid object: %+v", got)
			}
			if got.Name != tt.want.Name || got.Content != tt.want.Content {
				t.Errorf("Create() mismatch: got name=%q content=%q, want name=%q content=%q", got.Name, got.Content, tt.want.Name, tt.want.Content)
			}
			// Verify persistence by reading back from mysql_test
			back, err := r.Get(context.Background(), got.ID)
			if err != nil {
				t.Fatalf("Get() error: %v", err)
			}
			if back == nil || back.ID != got.ID || back.Name != tt.want.Name || back.Content != tt.want.Content {
				t.Errorf("Get() mismatch: got=%+v, want id=%d name=%q content=%q", back, got.ID, tt.want.Name, tt.want.Content)
			}
		})
	}
}
