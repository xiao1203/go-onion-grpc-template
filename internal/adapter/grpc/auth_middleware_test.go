package grpc

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	iauth "github.com/xiao1203/go-onion-grpc-template/internal/auth"
)

type pingReq struct{}

func runThrough(t *testing.T, headers map[string]string) (context.Context, error) {
	t.Helper()
	req := connect.NewRequest(&pingReq{})
	for k, v := range headers {
		req.Header().Set(k, v)
	}
	var got context.Context
	next := func(ctx context.Context, r connect.AnyRequest) (connect.AnyResponse, error) {
		got = ctx
		return nil, nil
	}
	u := AuthUnaryInterceptor(nil)
	_, err := u(next)(context.Background(), req)
	return got, err
}

// NOTE: allowlistをHTTP経由で厳密に検証するには、connect-goのプロシージャ設定とProtocolヘッダが必要。
// ここではミドルウェアの代表ケース（DEV_BYPASS / HS256 / Missing）に絞る。

func TestAuth_DevBypass(t *testing.T) {
	t.Setenv("DEV_AUTH_BYPASS", "1")
	ctx, err := runThrough(t, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, ok := iauth.FromContext(ctx); !ok {
		t.Fatalf("principal missing")
	}
}

func TestAuth_HS256_OK(t *testing.T) {
	t.Setenv("AUTH_HS256_SECRET", "secret")
	// build HS256 token with sub=1
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "1",
		"email": "dev@example.com",
		"exp":   time.Now().Add(5 * time.Minute).Unix(),
	})
	s, err := token.SignedString([]byte("secret"))
	if err != nil {
		t.Fatalf("sign err: %v", err)
	}
	ctx, err := runThrough(t, map[string]string{"Authorization": "Bearer " + s})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, ok := iauth.FromContext(ctx); !ok {
		t.Fatalf("principal missing")
	}
}

func TestAuth_MissingAuth_Unauthenticated(t *testing.T) {
	t.Setenv("DEV_AUTH_BYPASS", "")
	t.Setenv("AUTH_HS256_SECRET", "")
	t.Setenv("AUTH_JWKS_URL", "")
	_, err := runThrough(t, nil)
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("want Unauthenticated, got %v", connect.CodeOf(err))
	}
}
