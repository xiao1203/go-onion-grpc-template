package usecase_test

import (
	"context"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	mysqlRepositoryImpl "github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/mysql"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
	"github.com/xiao1203/go-onion-grpc-template/util/testhelper"
)

func TestSampleUsecase_Create(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testDB := testhelper.OpenGormTestDB(t)
	ctx := context.Background()
	repository := mysqlRepositoryImpl.NewSampleRepository(testDB)
	usecase := usecase.NewSampleUsecase(repository)

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		in      *entity.Sample
		want    *entity.Sample
		wantErr bool
	}{
		{
			name: "正常系: Sampleの作成に成功すること",
			in: &entity.Sample{
				Name:    "sample name",
				Content: "sample content",
				Count:   10,
			},
			want: &entity.Sample{
				ID:      1, // 無視するので値は何でも良い
				Name:    "sample name",
				Content: "sample content",
				Count:   10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := usecase.Create(ctx, tt.in)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Create() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Create() succeeded unexpectedly")
			} else {
				opts := cmp.Options{
					cmpopts.IgnoreFields(entity.Sample{}, "ID"),
				}
				if diff := cmp.Diff(tt.want, got, opts); diff != "" {
					t.Errorf("Create() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSampleUsecase_Get(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	// use repository fixtures from adapter/mysql
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	ctx := context.Background()
	repository := mysqlRepositoryImpl.NewSampleRepository(testDB)
	usecase := usecase.NewSampleUsecase(repository)

	tests := []struct {
		name string
		id   int64
		want *entity.Sample
	}{
		{
			name: "正常系: IDに対応するSampleデータを取得できること",
			id:   1,
			want: &entity.Sample{ID: 1, Name: "test_name_1", Content: "test_content_1", Count: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := usecase.Get(ctx, tt.id)
			if gotErr != nil {
				t.Fatalf("Get() failed: %v", gotErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Get() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSampleUsecase_List(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	repo := mysqlRepositoryImpl.NewSampleRepository(testDB)
	u := usecase.NewSampleUsecase(repo)
	ctx := context.Background()

	tests := []struct {
		name string
		p    domain.ListParams
		want []*entity.Sample
	}{
		{
			name: "正常系: データが存在する場合、全件取得できること",
			p:    domain.ListParams{Offset: 0, Limit: 100},
			want: []*entity.Sample{
				{ID: 3, Name: "test_name_3", Content: "test_content_3", Count: 3},
				{ID: 2, Name: "test_name_2", Content: "test_content_2", Count: 2},
				{ID: 1, Name: "test_name_1", Content: "test_content_1", Count: 1},
			},
		},
		{
			name: "正常系: offset/limit指定で該当件数を取得できること",
			p:    domain.ListParams{Offset: 1, Limit: 1},
			want: []*entity.Sample{
				{ID: 2, Name: "test_name_2", Content: "test_content_2", Count: 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := u.List(ctx, tt.p)
			if err != nil {
				t.Fatalf("List() failed: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("List() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestSampleUsecase_Update(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	repo := mysqlRepositoryImpl.NewSampleRepository(testDB)
	u := usecase.NewSampleUsecase(repo)

	tt := struct {
		name string
		in   *entity.Sample
		want *entity.Sample
	}{
		name: "正常系: 指定したIDのSampleレコードの更新に成功すること",
		in:   &entity.Sample{ID: 1, Name: "updated_name", Content: "updated_content", Count: 100},
		want: &entity.Sample{ID: 1, Name: "updated_name", Content: "updated_content", Count: 100},
	}
	t.Run(tt.name, func(t *testing.T) {
		got, err := u.Update(context.Background(), tt.in)
		if err != nil {
			t.Fatalf("Update() failed: %v", err)
		}
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("Update() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestSampleUsecase_Delete(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	repo := mysqlRepositoryImpl.NewSampleRepository(testDB)
	u := usecase.NewSampleUsecase(repo)

	tt := struct {
		name string
		id   int64
	}{
		name: "正常系: 指定したIDのSampleレコードの削除に成功すること",
		id:   1,
	}
	t.Run(tt.name, func(t *testing.T) {
		if err := u.Delete(context.Background(), tt.id); err != nil {
			t.Fatalf("Delete() failed: %v", err)
		}
		// ensure it’s gone
		if got, err := u.Get(context.Background(), tt.id); err != nil {
			t.Fatalf("Get() after delete failed: %v", err)
		} else if got != nil {
			t.Fatalf("Get() after delete = %#v, want nil", got)
		}
	})
}
