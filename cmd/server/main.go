package main

import (
    "fmt"
    "log"
    "net/http"

    grpcadapter "github.com/xiao1203/go-onion-grpc-template/internal/adapter/grpc"
    inframysql "github.com/xiao1203/go-onion-grpc-template/internal/infra/mysql"
)

func main() {
    mux := http.NewServeMux()

	// Registry-based DI: open shared DB (GORM) and register all generated routes
	db, err := inframysql.OpenGormFromEnv("")
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		defer sqlDB.Close()
	}
	grpcadapter.RegisterAll(mux, grpcadapter.Deps{Gorm: db})

	addr := ":8080"
	fmt.Printf("listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
