package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xiao1203/go-onion-grpc-template/gen/greeter/v1/greeterv1connect"
	grpcadapter "github.com/xiao1203/go-onion-grpc-template/internal/adapter/grpc"
	"github.com/xiao1203/go-onion-grpc-template/internal/adapter/repository/memory"
	"github.com/xiao1203/go-onion-grpc-template/internal/infra/mysql"
	"github.com/xiao1203/go-onion-grpc-template/internal/usecase"
)

func main() {
	// DI（後で wire に置き換えてもOK）
	repo := memory.NewGreeterRepository()
	uc := usecase.NewGreeterUsecase(repo)
	handler := grpcadapter.NewGreeterHandler(uc)

	mux := http.NewServeMux()
	path, h := greeterv1connect.NewGreeterServiceHandler(handler)
	mux.Handle(path, h)

	// Registry-based DI: open shared DB and register all generated routes
	db, err := mysql.OpenFromEnv("")
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()
	grpcadapter.RegisterAll(mux, grpcadapter.Deps{MySQL: db})

	addr := ":8080"
	fmt.Printf("listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
