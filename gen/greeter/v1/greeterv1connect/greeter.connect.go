// Minimal Connect handler glue for template. DO NOT EDIT.

package greeterv1connect

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	greeterv1 "github.com/xiao1203/go-onion-grpc-template/gen/greeter/v1"
)

const GreeterServiceName = "greeter.v1.GreeterService"

type GreeterServiceHandler interface {
	Hello(ctx context.Context, req *connect.Request[greeterv1.HelloRequest]) (*connect.Response[greeterv1.HelloResponse], error)
}

func NewGreeterServiceHandler(svc GreeterServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	path := "/" + GreeterServiceName + "/Hello"
	h := connect.NewUnaryHandler(
		path,
		svc.Hello,
		opts...,
	)
	return path, h
}
