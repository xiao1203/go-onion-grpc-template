package apperr

import (
    "connectrpc.com/connect"
    "github.com/newmo-oss/ergo"
)

// アプリ内で共通利用するエラーコード定義
var (
    // 認証関連
    Unauthenticated   = ergo.NewCode("Unauthenticated", "unauthenticated")
    PermissionDenied  = ergo.NewCode("PermissionDenied", "permission denied")

    // バリデーション/入力
    InvalidArgument   = ergo.NewCode("InvalidArgument", "invalid argument")

    // リソース関連
    NotFound          = ergo.NewCode("NotFound", "not found")
    Conflict          = ergo.NewCode("Conflict", "conflict")

    // その他
    Internal          = ergo.NewCode("Internal", "internal error")
)

// Connectのステータスコードに変換
func ToConnect(err error) error {
    if err == nil {
        return nil
    }
    code := ergo.CodeOf(err)
    if code.IsZero() {
        return connect.NewError(connect.CodeInternal, err)
    }

    switch code {
    case Unauthenticated:
        return connect.NewError(connect.CodeUnauthenticated, err)
    case PermissionDenied:
        return connect.NewError(connect.CodePermissionDenied, err)
    case InvalidArgument:
        return connect.NewError(connect.CodeInvalidArgument, err)
    case NotFound:
        return connect.NewError(connect.CodeNotFound, err)
    case Conflict:
        return connect.NewError(connect.CodeAlreadyExists, err)
    default:
        return connect.NewError(connect.CodeInternal, err)
    }
}

