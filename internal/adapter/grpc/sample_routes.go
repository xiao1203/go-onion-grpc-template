package grpc

import (
	"net/http"

	samplev1connect "github.com/xiao1203/go-onion-grpc-template/gen/sample/v1/samplev1connect"
	mysqlrepo "github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/mysql"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

func init() { Add(registerSample) }

func registerSample(mux *http.ServeMux, deps Deps) {
	repo := mysqlrepo.NewSampleRepository(deps.Gorm)
	uc := usecase.NewSampleUsecase(repo)
	h := NewSampleHandler(uc)
	path, handler := samplev1connect.NewSampleServiceHandler(h)
	mux.Handle(path, handler)
}
