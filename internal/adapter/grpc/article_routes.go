package grpc

import (
    "net/http"

    articlev1connect "github.com/xiao1203/go-onion-grpc-template/gen/article/v1/articlev1connect"
    mysqlrepo "github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/mysql"
    "github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

func init() { Add(registerArticle) }

func registerArticle(mux *http.ServeMux, deps Deps) {
    repo := mysqlrepo.NewArticleRepository(deps.Gorm)
    uc := usecase.NewArticleUsecase(repo)
    h := NewArticleHandler(uc)
    path, handler := articlev1connect.NewArticleServiceHandler(h)
    mux.Handle(path, handler)
}
