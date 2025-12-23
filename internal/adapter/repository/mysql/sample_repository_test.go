package mysql_test

import (
	"context"
	"testing"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/mysql"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
	"github.com/xiao1203/go-onion-grpc-template/internal/domain/entity"
	"github.com/xiao1203/go-onion-grpc-template/util/testhelper"
	"gorm.io/gorm"
)

func TestSampleRepository_Create(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testDB := testhelper.OpenGormTestDB(t)
	repository := mysql.NewSampleRepository(testDB)
	ctx := context.Background()

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
			got, gotErr := repository.Create(ctx, tt.in)
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

func TestSampleRepository_Get(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	repository := mysql.NewSampleRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		id      int64
		want    *entity.Sample
		wantErr bool
	}{
		{
			name: "正常系: IDに対応するSampleデータを取得できること",
			id:   1,
			want: &entity.Sample{
				ID:      1,
				Name:    "test_name_1",
				Content: "test_content_1",
				Count:   1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := repository.Get(ctx, tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Get() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Get() succeeded unexpectedly")
			} else {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("Get() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSampleRepository_List(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	repository := mysql.NewSampleRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		p       domain.ListParams
		want    []*entity.Sample
		wantErr bool
	}{
		{
			name: "正常系: データが存在する場合、全件取得できること",
			p: domain.ListParams{
				Offset: 0,
				Limit:  100,
			},
			want: []*entity.Sample{
				{
					ID:      3,
					Name:    "test_name_3",
					Content: "test_content_3",
					Count:   3,
				},
				{
					ID:      2,
					Name:    "test_name_2",
					Content: "test_content_2",
					Count:   2,
				},
				{
					ID:      1,
					Name:    "test_name_1",
					Content: "test_content_1",
					Count:   1,
				},
			},
			wantErr: false,
		},
		{
			name: "正常系: データが存在する場合でoffset、limitを指定した場合、該当件数分取得できること",
			p: domain.ListParams{
				Offset: 1,
				Limit:  1,
			},
			want: []*entity.Sample{
				{
					ID:      2,
					Name:    "test_name_2",
					Content: "test_content_2",
					Count:   2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := repository.List(ctx, tt.p)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("List() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("List() succeeded unexpectedly")
			} else {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("List() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSampleRepository_Update(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	repository := mysql.NewSampleRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		in      *entity.Sample
		want    *entity.Sample
		wantErr bool
	}{
		{
			name: "正常系: 指定したIDのSampleレコードの更新に成功すること",
			in: &entity.Sample{
				ID:      1,
				Name:    "updated_name",
				Content: "updated_content",
				Count:   100,
			},
			want: &entity.Sample{
				ID:      1,
				Name:    "updated_name",
				Content: "updated_content",
				Count:   100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := repository.Update(ctx, tt.in)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Update() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Update() succeeded unexpectedly")
			} else {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("Update() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSampleRepository_Delete(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	repository := mysql.NewSampleRepository(testDB)
	ctx := context.Background()

	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		db *gorm.DB
		// Named input parameters for target function.
		id      int64
		wantErr bool
	}{
		{
			name:    "正常系: 指定したIDのSampleレコードの削除に成功すること",
			db:      testDB,
			id:      1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := repository.Delete(ctx, tt.id)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Delete() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Delete() succeeded unexpectedly")
			}
		})
	}
}
