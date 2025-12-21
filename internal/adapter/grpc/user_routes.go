package grpc

import (
	"net/http"

	"connectrpc.com/connect"
	userv1connect "github.com/xiao1203/go-onion-grpc-template/gen/user/v1/userv1connect"
	mysqlrepo "github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/mysql"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

func init() { Add(registerUser) }

func registerUser(mux *http.ServeMux, deps Deps) {
	repo := mysqlrepo.NewUserRepository(deps.Gorm)
	uc := usecase.NewUserUsecase(repo)
	h := NewUserHandler(uc)
	// attach auth interceptor (public allowlist currently empty)
	opts := connect.WithInterceptors(AuthUnaryInterceptor(PublicAllowlist()))
	path, handler := userv1connect.NewUserServiceHandler(h, opts)
	mux.Handle(path, handler)
}
