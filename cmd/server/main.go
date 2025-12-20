package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xiao1203/go-onion-grpc-template/gen/greeter/v1/greeterv1connect"
	grpcadapter "github.com/xiao1203/go-onion-grpc-template/internal/adapter/grpc"
	"github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/memory"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"

	// scaffold:imports (DO NOT REMOVE)
)

func main() {
	// DI（後で wire に置き換えてもOK）
	repo := memory.NewGreeterRepository()
	uc := usecase.NewGreeterUsecase(repo)
	handler := grpcadapter.NewGreeterHandler(uc)

	mux := http.NewServeMux()
	path, h := greeterv1connect.NewGreeterServiceHandler(handler)
	mux.Handle(path, h)

	// scaffold:routes (DO NOT REMOVE)


	addr := ":8080"
	fmt.Printf("listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
