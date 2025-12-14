package grpc

import (
	"context"

	"connectrpc.com/connect"
	greeterv1 "github.com/xiao1203/go-onion-grpc-template/gen/greeter/v1"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

type GreeterHandler struct {
	uc *usecase.GreeterUsecase
}

func NewGreeterHandler(uc *usecase.GreeterUsecase) *GreeterHandler {
	return &GreeterHandler{uc: uc}
}

func (h *GreeterHandler) Hello(
	ctx context.Context,
	req *connect.Request[greeterv1.HelloRequest],
) (*connect.Response[greeterv1.HelloResponse], error) {
	g, err := h.uc.Hello(ctx, req.Msg.GetName())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	res := connect.NewResponse(&greeterv1.HelloResponse{Message: g.Message})
	return res, nil
}
