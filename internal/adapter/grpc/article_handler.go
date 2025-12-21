package grpc

import (
	"context"

	"connectrpc.com/connect"
	articlev1 "github.com/xiao1203/go-onion-grpc-template/gen/article/v1"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

type ArticleHandler struct {
	uc *usecase.ArticleUsecase
}

func NewArticleHandler(uc *usecase.ArticleUsecase) *ArticleHandler {
	return &ArticleHandler{uc: uc}
}

func (h *ArticleHandler) CreateArticle(
	ctx context.Context,
	req *connect.Request[articlev1.CreateArticleRequest],
) (*connect.Response[articlev1.CreateArticleResponse], error) {
	in := &usecase.Article{
		Name:    req.Msg.GetName(),
		Content: req.Msg.GetContent(),
	}
	out, err := h.uc.Create(ctx, in)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	res := connect.NewResponse(&articlev1.CreateArticleResponse{
		Article: toProtoArticle(out),
	})
	return res, nil
}

func (h *ArticleHandler) GetArticle(
	ctx context.Context,
	req *connect.Request[articlev1.GetArticleRequest],
) (*connect.Response[articlev1.GetArticleResponse], error) {
	out, err := h.uc.Get(ctx, req.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&articlev1.GetArticleResponse{Article: toProtoArticle(out)}), nil
}

func (h *ArticleHandler) ListArticles(
	ctx context.Context,
	req *connect.Request[articlev1.ListArticlesRequest],
) (*connect.Response[articlev1.ListArticlesResponse], error) {
	items, err := h.uc.List(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	out := make([]*articlev1.Article, 0, len(items))
	for _, it := range items {
		out = append(out, toProtoArticle(it))
	}
	return connect.NewResponse(&articlev1.ListArticlesResponse{Articles: out}), nil
}

func (h *ArticleHandler) UpdateArticle(
	ctx context.Context,
	req *connect.Request[articlev1.UpdateArticleRequest],
) (*connect.Response[articlev1.UpdateArticleResponse], error) {
	in := &usecase.Article{
		ID:      req.Msg.GetId(),
		Name:    req.Msg.GetName(),
		Content: req.Msg.GetContent(),
	}
	out, err := h.uc.Update(ctx, in)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&articlev1.UpdateArticleResponse{Article: toProtoArticle(out)}), nil
}

func (h *ArticleHandler) DeleteArticle(
	ctx context.Context,
	req *connect.Request[articlev1.DeleteArticleRequest],
) (*connect.Response[articlev1.DeleteArticleResponse], error) {
	if err := h.uc.Delete(ctx, req.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&articlev1.DeleteArticleResponse{}), nil
}

func toProtoArticle(in *usecase.Article) *articlev1.Article {
	if in == nil {
		return nil
	}
	return &articlev1.Article{
		Id:      in.ID,
		Name:    in.Name,
		Content: in.Content,
	}
}
