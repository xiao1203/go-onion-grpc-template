package memory

import (
	"context"
	"fmt"

	"github.com/xiao1203/go-onion-grpc-template/internal/domain"
)

type GreeterRepository struct{}

func NewGreeterRepository() *GreeterRepository { return &GreeterRepository{} }

func (r *GreeterRepository) BuildGreeting(ctx context.Context, name string) (*domain.Greeting, error) {
	return &domain.Greeting{Message: fmt.Sprintf("hello, %s", name)}, nil
}
