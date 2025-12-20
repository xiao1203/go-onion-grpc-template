package grpc

import (
    "context"
    "errors"
    "os"
    "strconv"
    "strings"
    "time"
    "sync"

    "connectrpc.com/connect"
    "github.com/golang-jwt/jwt/v5"
    "github.com/xiao1203/go-onion-grpc-template/internal/auth"
)

// AuthUnaryInterceptor enforces auth unless the method is allowlisted.
func AuthUnaryInterceptor(allowlist map[string]struct{}) connect.UnaryInterceptorFunc {
    return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            if _, ok := allowlist[req.Spec().Procedure]; ok {
                return next(ctx, req)
            }
            if os.Getenv("DEV_AUTH_BYPASS") == "1" {
                uid := int64(1)
                if s := os.Getenv("DEV_USER_ID"); s != "" {
                    if v, err := strconv.ParseInt(s, 10, 64); err == nil { uid = v }
                }
                p := &auth.Principal{UserID: uid, Email: "dev@example.com", Roles: []string{"admin","user"}}
                return next(auth.WithPrincipal(ctx, p), req)
            }
            // Prefer JWKS (OIDC) if configured
            if jwksURL := os.Getenv("AUTH_JWKS_URL"); jwksURL != "" {
                ctx2, err := withJWTFromHeader(ctx, req, func(token *jwt.Token) (any, error) {
                    // cache per process
                    ttl := 5 * time.Minute
                    if s := os.Getenv("AUTH_JWKS_TTL"); s != "" {
                        if d, err := time.ParseDuration(s); err == nil { ttl = d }
                    }
                    c := jwksCacheSingleton(jwksURL, ttl)
                    kid, _ := token.Header["kid"].(string)
                    return c.KeyFor(kid)
                }, verifyStandardClaims())
                if err != nil { return nil, err }
                return next(ctx2, req)
            }

            authz := req.Header().Get("Authorization")
            if authz == "" { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("missing Authorization")) }
            parts := strings.SplitN(authz, " ", 2)
            if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
                return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid Authorization"))
            }
            tokenString := parts[1]
            hs := os.Getenv("AUTH_HS256_SECRET")
            if hs == "" {
                return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no verifier configured"))
            }
            token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
                if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                    return nil, errors.New("invalid signing method")
                }
                return []byte(hs), nil
            })
            if err != nil || !token.Valid {
                return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
            }
            claims, ok := token.Claims.(jwt.MapClaims)
            if !ok { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid claims")) }
            if exp, ok := claims["exp"].(float64); ok {
                if time.Now().Unix() > int64(exp) { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token expired")) }
            }
            var uid int64
            if sub, ok := claims["sub"].(string); ok {
                if v, err := strconv.ParseInt(sub, 10, 64); err == nil { uid = v }
            }
            email, _ := claims["email"].(string)
            var roles []string
            if rr, ok := claims["roles"].([]any); ok {
                for _, r := range rr { if s, ok := r.(string); ok { roles = append(roles, s) } }
            }
            p := &auth.Principal{UserID: uid, Email: email, Roles: roles}
            return next(auth.WithPrincipal(ctx, p), req)
        }
    })
}

func PublicAllowlist() map[string]struct{} { return map[string]struct{}{} }

// helpers
var (
    _jwksMu sync.Mutex
    _jwks   *auth.JWKSCache
    _jwksURL string
    _jwksTTL time.Duration
)

func jwksCacheSingleton(url string, ttl time.Duration) *auth.JWKSCache {
    _jwksMu.Lock()
    defer _jwksMu.Unlock()
    if _jwks == nil || _jwksURL != url || _jwksTTL != ttl {
        _jwks = auth.NewJWKSCache(url, ttl)
        _jwksURL = url
        _jwksTTL = ttl
    }
    return _jwks
}

func withJWTFromHeader(ctx context.Context, req connect.AnyRequest, keyfunc jwt.Keyfunc, claimCheck func(jwt.MapClaims) error) (context.Context, error) {
    authz := req.Header().Get("Authorization")
    if authz == "" { return ctx, connect.NewError(connect.CodeUnauthenticated, errors.New("missing Authorization")) }
    parts := strings.SplitN(authz, " ", 2)
    if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
        return ctx, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid Authorization"))
    }
    tokenString := parts[1]
    token, err := jwt.Parse(tokenString, keyfunc)
    if err != nil || !token.Valid {
        return ctx, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid token"))
    }
    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok { return ctx, connect.NewError(connect.CodeUnauthenticated, errors.New("invalid claims")) }
    if err := claimCheck(claims); err != nil {
        return ctx, connect.NewError(connect.CodeUnauthenticated, err)
    }
    var uid int64
    if sub, ok := claims["sub"].(string); ok { if v, err := strconv.ParseInt(sub, 10, 64); err == nil { uid = v } }
    email, _ := claims["email"].(string)
    var roles []string
    if rr, ok := claims["roles"].([]any); ok { for _, r := range rr { if s, ok := r.(string); ok { roles = append(roles, s) } } }
    p := &auth.Principal{UserID: uid, Email: email, Roles: roles}
    return auth.WithPrincipal(ctx, p), nil
}

func verifyStandardClaims() func(jwt.MapClaims) error {
    iss := os.Getenv("AUTH_ISSUER")
    audWant := os.Getenv("AUTH_AUDIENCE")
    skew := 60 * time.Second
    if s := os.Getenv("AUTH_CLOCK_SKEW"); s != "" { if d, err := time.ParseDuration(s); err == nil { skew = d } }
    return func(c jwt.MapClaims) error {
        now := time.Now()
        if iss != "" {
            if v, _ := c["iss"].(string); v != iss { return errors.New("issuer mismatch") }
        }
        if audWant != "" {
            switch aud := c["aud"].(type) {
            case string:
                if aud != audWant { return errors.New("audience mismatch") }
            case []any:
                ok := false
                for _, a := range aud { if s, _ := a.(string); s == audWant { ok = true; break } }
                if !ok { return errors.New("audience mismatch") }
            }
        }
        if exp, ok := c["exp"].(float64); ok { if now.After(time.Unix(int64(exp),0).Add(skew)) { return errors.New("token expired") } }
        if nbf, ok := c["nbf"].(float64); ok { if now.Before(time.Unix(int64(nbf),0).Add(-skew)) { return errors.New("token not yet valid") } }
        return nil
    }
}
