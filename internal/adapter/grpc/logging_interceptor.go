package grpc

import (
    "context"
    "fmt"
    "log/slog"

    "connectrpc.com/connect"
    "github.com/newmo-oss/ergo"
)

// LoggingUnaryInterceptor logs errors with ergo stacktrace when available.
func LoggingUnaryInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
    if logger == nil {
        logger = slog.Default()
    }
    return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            res, err := next(ctx, req)
            if err != nil {
                st := ergo.StackTraceOf(err)
                logger.Error("rpc error",
                    slog.String("procedure", req.Spec().Procedure),
                    slog.String("code", connect.CodeOf(err).String()),
                    slog.String("error", fmt.Sprintf("%+v", err)),
                    slog.String("stack", fmt.Sprintf("%v", st)),
                )
            }
            return res, err
        }
    })
}

