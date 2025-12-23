package grpc_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/google/go-cmp/cmp"
	samplev1 "github.com/xiao1203/go-onion-grpc-template/gen/sample/v1"
	"github.com/xiao1203/go-onion-grpc-template/gen/sample/v1/samplev1connect"
	"github.com/xiao1203/go-onion-grpc-template/internal/adapter/grpc"
	mysqlRepositoryImpl "github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/mysql"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
	"github.com/xiao1203/go-onion-grpc-template/util/testhelper"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestSampleHandler_CreateSample(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	// testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))
	testDB := testhelper.OpenGormTestDB(t)
	ctx := context.Background()
	mux := http.NewServeMux()

	repository := mysqlRepositoryImpl.NewSampleRepository(testDB)
	usecase := usecase.NewSampleUsecase(repository)
	h := grpc.NewSampleHandler(usecase)
	path, handler := samplev1connect.NewSampleServiceHandler(h)
	mux.Handle(path, handler)

	// grpc.NewSampleHandler(usecase), connect.WithInterceptors()
	server := httptest.NewServer(mux)
	t.Cleanup(func() { server.Close() })

	client := samplev1connect.NewSampleServiceClient(server.Client(), server.URL)

	type args struct {
		req *connect.Request[samplev1.CreateSampleRequest]
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "正常系: Sampleの作成に成功すること",
			args: args{
				req: connect.NewRequest(&samplev1.CreateSampleRequest{
					Name:    "sample name",
					Content: "sample content",
					Count:   10,
				}),
			},
			wantErr: false,
		},
		{
			name: "異常系: nameが長すぎる場合はInternalエラーになること",
			args: args{
				req: func() *connect.Request[samplev1.CreateSampleRequest] {
					long := strings.Repeat("x", 300)
					return connect.NewRequest(&samplev1.CreateSampleRequest{Name: long, Content: "ok", Count: 1})
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := client.CreateSample(ctx, tt.args.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateSample() failed: %v", gotErr)
				}
				// error path: check code is Internal
				if connect.CodeOf(gotErr) != connect.CodeInternal {
					t.Fatalf("want CodeInternal, got %v", connect.CodeOf(gotErr))
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateSample() succeeded unexpectedly")
			} else {
				wantSample := &samplev1.Sample{ // ignore id
					Name:    "sample name",
					Content: "sample content",
					Count:   10,
				}
				if diff := cmp.Diff(wantSample, got.Msg.GetSample(), protocmp.Transform(), protocmp.IgnoreFields(&samplev1.Sample{}, "id")); diff != "" {
					t.Errorf("CreateSample mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSampleHandler_GetSample(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory("testdata/fixture/sample"))

	testDB := testhelper.OpenGormTestDB(t)
	ctx := context.Background()
	mux := http.NewServeMux()

	repository := mysqlRepositoryImpl.NewSampleRepository(testDB)
	uc := usecase.NewSampleUsecase(repository)
	h := grpc.NewSampleHandler(uc)
	path, handler := samplev1connect.NewSampleServiceHandler(h)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(func() { server.Close() })

	client := samplev1connect.NewSampleServiceClient(server.Client(), server.URL)

	req := connect.NewRequest(&samplev1.GetSampleRequest{Id: 1})
	got, err := client.GetSample(ctx, req)
	if err != nil {
		t.Fatalf("GetSample() error = %v", err)
	}
	want := &samplev1.Sample{Id: 1, Name: "test_name_1", Content: "test_content_1", Count: 1}
	if diff := cmp.Diff(want, got.Msg.GetSample(), protocmp.Transform()); diff != "" {
		t.Errorf("GetSample mismatch (-want +got):\n%s", diff)
	}
}

func TestSampleHandler_ListSamples(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory(testhelper.FixturePath("internal/adapter/repository/mysql/testdata/fixture/sample")))

	testDB := testhelper.OpenGormTestDB(t)
	ctx := context.Background()
	mux := http.NewServeMux()

	repo := mysqlRepositoryImpl.NewSampleRepository(testDB)
	uc := usecase.NewSampleUsecase(repo)
	h := grpc.NewSampleHandler(uc)
	path, handler := samplev1connect.NewSampleServiceHandler(h)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(func() { server.Close() })

	client := samplev1connect.NewSampleServiceClient(server.Client(), server.URL)

	type args struct {
		req *connect.Request[samplev1.ListSamplesRequest]
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "正常系: データが存在する場合、全件取得できること",
			args:    args{req: connect.NewRequest(&samplev1.ListSamplesRequest{})},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := client.ListSamples(ctx, tt.args.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ListSamples() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ListSamples() succeeded unexpectedly")
			} else {
				want := []*samplev1.Sample{
					{Id: 3, Name: "test_name_3", Content: "test_content_3", Count: 3},
					{Id: 2, Name: "test_name_2", Content: "test_content_2", Count: 2},
					{Id: 1, Name: "test_name_1", Content: "test_content_1", Count: 1},
				}
				if diff := cmp.Diff(want, got.Msg.GetSamples(), protocmp.Transform()); diff != "" {
					t.Errorf("ListSamples mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSampleHandler_UpdateSample(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory(testhelper.FixturePath("internal/adapter/repository/mysql/testdata/fixture/sample")))

	testDB := testhelper.OpenGormTestDB(t)
	ctx := context.Background()
	mux := http.NewServeMux()

	repo := mysqlRepositoryImpl.NewSampleRepository(testDB)
	uc := usecase.NewSampleUsecase(repo)
	h := grpc.NewSampleHandler(uc)
	path, handler := samplev1connect.NewSampleServiceHandler(h)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(func() { server.Close() })

	client := samplev1connect.NewSampleServiceClient(server.Client(), server.URL)

	type args struct {
		req *connect.Request[samplev1.UpdateSampleRequest]
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "正常系: 指定したIDのSampleレコードの更新に成功すること",
			args:    args{req: connect.NewRequest(&samplev1.UpdateSampleRequest{Id: 1, Name: "updated_name", Content: "updated_content", Count: 100})},
			wantErr: false,
		},
		{
			name: "異常系: nameが長すぎる場合はInternalエラーになること",
			args: args{req: func() *connect.Request[samplev1.UpdateSampleRequest] {
				long := strings.Repeat("y", 300)
				return connect.NewRequest(&samplev1.UpdateSampleRequest{Id: 1, Name: long, Content: "ok", Count: 1})
			}()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := client.UpdateSample(ctx, tt.args.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("UpdateSample() failed: %v", gotErr)
				}
				if connect.CodeOf(gotErr) != connect.CodeInternal {
					t.Fatalf("want CodeInternal, got %v", connect.CodeOf(gotErr))
				}
				return
			}
			if tt.wantErr {
				t.Fatal("UpdateSample() succeeded unexpectedly")
			} else {
				want := &samplev1.Sample{Id: 1, Name: "updated_name", Content: "updated_content", Count: 100}
				if diff := cmp.Diff(want, got.Msg.GetSample(), protocmp.Transform()); diff != "" {
					t.Errorf("UpdateSample mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestSampleHandler_DeleteSample(t *testing.T) {
	testhelper.Lock(t)
	testhelper.EnsureTestDBEnv(t)
	testhelper.LoadTestFixtures(t, testfixtures.Directory(testhelper.FixturePath("internal/adapter/repository/mysql/testdata/fixture/sample")))

	testDB := testhelper.OpenGormTestDB(t)
	ctx := context.Background()
	mux := http.NewServeMux()

	repo := mysqlRepositoryImpl.NewSampleRepository(testDB)
	uc := usecase.NewSampleUsecase(repo)
	h := grpc.NewSampleHandler(uc)
	path, handler := samplev1connect.NewSampleServiceHandler(h)
	mux.Handle(path, handler)

	server := httptest.NewServer(mux)
	t.Cleanup(func() { server.Close() })

	client := samplev1connect.NewSampleServiceClient(server.Client(), server.URL)

	type args struct {
		req *connect.Request[samplev1.DeleteSampleRequest]
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "正常系: 指定したIDのSampleレコードの削除に成功すること",
			args:    args{req: connect.NewRequest(&samplev1.DeleteSampleRequest{Id: 1})},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotErr := client.DeleteSample(ctx, tt.args.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeleteSample() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeleteSample() succeeded unexpectedly")
			} else {
				// ensure deleted
				getReq := connect.NewRequest(&samplev1.GetSampleRequest{Id: tt.args.req.Msg.GetId()})
				got, err := client.GetSample(ctx, getReq)
				if err != nil {
					t.Fatalf("GetSample() after delete error = %v", err)
				}
				if got.Msg.GetSample() != nil {
					t.Fatalf("GetSample() after delete = %#v, want nil Sample", got.Msg.GetSample())
				}
			}
		})
	}
}
