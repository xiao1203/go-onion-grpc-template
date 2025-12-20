# 認証の使い方（テンプレート活用ガイド）

このドキュメントは、本テンプレートから新規アプリケーションを作成したあとに、認証（OIDC/JWT）の導入・設定・公開/保護エンドポイントの出し分けを進めるための手引きです。

テンプレートには以下の“土台”が含まれています。
- 認証インターセプタ（connect-go用）: `internal/adapter/grpc/auth_middleware.go`
- Principal（ユーザー情報の受け渡し）: `internal/auth/principal.go`
- JWKSキャッシュ（OIDCの公開鍵取得）: `internal/auth/jwks.go`
- User自己参照API（最小）: `proto/user/v1/user.proto` + 実装（usecase/repository/handler/routes）

以降は「環境変数の設定 → 動作確認 → 公開/保護の切り替え → OIDC本番運用」までを段階的に説明します。

---

## 1. 環境変数（認証に関する設定）

apiサービス（Docker Compose）の environment に設定します。ローカル検証・本番運用で使い分けてください。

- 開発用（どれか1つ）
  - `DEV_AUTH_BYPASS` … 1を設定すると認証をバイパスし、開発用Principalを注入します（本番では絶対に使わない）
    - `DEV_USER_ID` … バイパス時のユーザーID（既定1）
  - `AUTH_HS256_SECRET` … HS256署名JWTの検証に使用（ローカル簡易検証向け）

- OIDC（本番運用/Keycloak/Cognitoなど）
  - `AUTH_JWKS_URL` … JWKSのURL（`/.well-known/jwks.json`）
  - `AUTH_ISSUER` … issの期待値（任意）
  - `AUTH_AUDIENCE` … audの期待値（任意）
  - `AUTH_JWKS_TTL` … JWKSキャッシュTTL（例: `5m`）
  - `AUTH_CLOCK_SKEW` … 時計ズレ許容（例: `60s`）

推奨: テンプレートのdocker-compose.ymlはデフォルトでは**DEV_AUTH_BYPASSを無効に**し、必要時に各プロジェクトで有効化してください。

---

## 2. まずは動かす（クイックスタート）

1) User自己参照APIの疎通（開発用バイパス）
- `DEV_AUTH_BYPASS=1` を有効化 → API再起動（make restart）
- 任意で開発ユーザーの作成（id=1 など）
  ```sql
  INSERT INTO users(email, display_name, created_at, updated_at)
  VALUES ('dev@example.com', 'Dev User', NOW(6), NOW(6));
  ```
- User APIを叩く
  ```bash
  curl -sS -X POST -H 'Content-Type: application/json' \
    -d '{}' http://127.0.0.1:8080/user.v1.UserService/GetMe
  
  curl -sS -X POST -H 'Content-Type: application/json' \
    -d '{"display_name":"New Name","picture_url":"https://example.com/me.png"}' \
    http://127.0.0.1:8080/user.v1.UserService/UpdateMyProfile
  ```

2) HS256での検証（ローカル簡易）
- `AUTH_HS256_SECRET=devsecret` を設定し、`DEV_AUTH_BYPASS` を無効化
- subを文字列"1"にしたHS256トークンを作り、Authorizationヘッダで送付
  ```go
  // 例: Goでトークン作成
  token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
    "sub": "1",
    "email": "dev@example.com",
    "exp": time.Now().Add(5 * time.Minute).Unix(),
  })
  s, _ := token.SignedString([]byte("devsecret"))
  ```
  ```bash
  curl -sS -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $s" \
    -d '{}' http://127.0.0.1:8080/user.v1.UserService/GetMe
  ```

3) OIDC（JWKS）での検証（Keycloak/Cognitoなど）
- `AUTH_JWKS_URL` を設定し（必要に応じて `AUTH_ISSUER`/`AUTH_AUDIENCE` も）、IdP発行のアクセストークンを `Authorization: Bearer` で送付

---

## 3. 認証の仕組み（内部動作）

- Unary Interceptor（`internal/adapter/grpc/auth_middleware.go`）が**最初に**リクエストを受け、以下の順に判定します。
  1) AllowListに該当するメソッド（公開API）なら認証スキップ
  2) `DEV_AUTH_BYPASS=1` なら開発用Principalを注入
  3) `AUTH_JWKS_URL` があればJWKS（RS署名）で検証（標準クレームiss/aud/exp/nbfも検証）
  4) なければ `AUTH_HS256_SECRET`（HS256）で検証
  5) いずれもなければ Unauthenticated
- 検証OKなら `internal/auth/principal.go` の Principal を context に注入し、ハンドラに渡します。

---

## 4. 公開/保護エンドポイントの出し分け（AllowList）

- 既定ではAllowListは空（＝認証必須）です。
- 公開にしたいメソッドがある場合、**各サービスのroutes**でAllowListを与えて Interceptor を装着します。

### 適用例（ArticleService に認証をかける）

テンプレートでは適用例のコードは含めませんが、実際のアプリでは以下のようにします。

```go
// internal/adapter/grpc/article_routes.go
import (
  "net/http"
  "connectrpc.com/connect"
  articlev1connect "<module>/gen/article/v1/articlev1connect"
  mysqlrepo "<module>/internal/adapter/repository/mysql"
  "<module>/internal/usecase"
)

func init() { Add(registerArticle) }

func registerArticle(mux *http.ServeMux, deps Deps) {
  repo := mysqlrepo.NewArticleRepository(deps.Gorm)
  uc := usecase.NewArticleUsecase(repo)
  h := NewArticleHandler(uc)

  // 認証を装着。allowlistに公開メソッドのみ記述（空なら全保護）
  allow := map[string]struct{}{
    // "/article.v1.ArticleService/ListArticles": {},
  }
  opts := connect.WithInterceptors(AuthUnaryInterceptor(allow))
  path, handler := articlev1connect.NewArticleServiceHandler(h, opts)
  mux.Handle(path, handler)
}
```

- 全メソッド認証必須にする場合は `allow := map[string]struct{}{}`（空）
- 一部公開する場合は、公開したいメソッドをフル名で追加
  - 例: `"/article.v1.ArticleService/ListArticles": {}`

---

## 5. User自己参照APIの目的と使いどころ

- UserService（GetMe/UpdateMyProfile）は「認証後の利用者が自分のプロフィールを扱う」ための最小APIです。
- 認証バイパス or JWT検証が通っていれば、GetMeではcontext内のPrincipalに基づきDBからユーザーを取得します。
- UpdateMyProfileは表示名・アイコンURLなど、ユーザーが自身で更新可能なフィールドに限定します。
- 管理者向けAPI（ListUsers/SetUserRoles/UpdateUserStatusなど）は必要に応じて追加し、認可（roles）チェックを併用してください。

### 役割（RBAC）の最小セット
- 例として roles に `admin`/`user` を投入し、user_roles で付与します。
  ```sql
  INSERT IGNORE INTO roles(name, description, created_at, updated_at)
  VALUES ('admin','administrator',NOW(6),NOW(6)),('user','normal user',NOW(6),NOW(6));

  -- 開発ユーザー id=1 に admin/user を付与
  INSERT IGNORE INTO user_roles(user_id, role_id, created_at)
  SELECT 1, id, NOW(6) FROM roles WHERE name IN ('admin','user');
  ```

---

## 6. よくある構成例（ローカル→本番）

- ローカル最短
  - `DEV_AUTH_BYPASS=1` でまずはAPIを通す
  - User自己参照APIで動作・配線を確認
- 次の段階（安全性を上げる）
  - `AUTH_HS256_SECRET` でJWT検証を導入（subを"1"にしてテスト）
- 本番運用
  - `AUTH_JWKS_URL` を設定（Cognito/Keycloakなど）
  - 必要に応じて `AUTH_ISSUER`/`AUTH_AUDIENCE` を設定し、iss/audチェックを有効化
  - DEV_AUTH_BYPASS は**必ず無効**に

---

## 7. ハマりどころとTips

- AllowList
  - フルプロシージャ名で指定します。誤った文字列だと公開されません。
- subの扱い
  - HS256の例では sub を"1"（文字列）としてUserIDにパースしています。IdPによりsubが非数値のことが多いので、実際には user_identities(issuer, sub) → users(id) の解決を導入してください。
- 時計ズレ
  - `AUTH_CLOCK_SKEW` でexp/nbfの前後ぶれを吸収できます（既定60s）
- 本番と開発
  - 開発中は `DEV_AUTH_BYPASS` で最短の手触り、仕上げで JWT/JWKS に切替、公開/保護の出し分けはAllowListで整える方針がおすすめです。

---

## 8. 参考（環境変数のCompose例）

```yaml
services:
  api:
    environment:
      # dev DB/test DB はテンプレの既定を使用
      # 認証（いずれかを有効に）
      # 開発用（バイパス）
      # DEV_AUTH_BYPASS: "1"
      # DEV_USER_ID: "1"

      # HS256（簡易検証）
      # AUTH_HS256_SECRET: devsecret

      # OIDC（本番）
      # AUTH_JWKS_URL: "https://<idp>/.well-known/jwks.json"
      # AUTH_ISSUER: "https://<idp>/realms/dev"
      # AUTH_AUDIENCE: "myclient"
      # AUTH_JWKS_TTL: "5m"
      # AUTH_CLOCK_SKEW: "60s"
```

---

## 9. 既存サービス（Articleなど）への適用例

テンプレート自体には適用しませんが、実アプリ側では以下のようにroutesでInterceptorを付与します（再掲）。

```go
import (
  "net/http"
  "connectrpc.com/connect"
  articlev1connect "<module>/gen/article/v1/articlev1connect"
  mysqlrepo "<module>/internal/adapter/repository/mysql"
  "<module>/internal/usecase"
)

func init() { Add(registerArticle) }

func registerArticle(mux *http.ServeMux, deps Deps) {
  repo := mysqlrepo.NewArticleRepository(deps.Gorm)
  uc := usecase.NewArticleUsecase(repo)
  h := NewArticleHandler(uc)

  // 例: Listだけ公開し、Get/Update/Deleteは認証必須
  allow := map[string]struct{}{
    "/article.v1.ArticleService/ListArticles": {},
  }
  opts := connect.WithInterceptors(AuthUnaryInterceptor(allow))
  path, handler := articlev1connect.NewArticleServiceHandler(h, opts)
  mux.Handle(path, handler)
}
```

---

以上をベースに、まずは開発用バイパスで最短経路→HS256→OIDC(JWKS)の順でステップアップし、公開/保護の出し分けはAllowListで段階的に整えてください。

