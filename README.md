# go-onion-grpc-template

Go 言語で **オニオンアーキテクチャ + gRPC（connect-go）** を採用した  
**フル Docker 開発環境付きテンプレート**です。

このテンプレートは、1コマンドで CRUD の gRPC API を雛形生成し、
マイグレーションと疎通確認（HTTPリクエスト）まで実行できます。

---

## クイックスタート（雛形生成〜疎通まで）

1) 起動（初回はコンテナを構築）

```
make up
```

2) 例：Article エンティティを生成（name:string, content:string）

```
make scaffold-all name=Article fields="name:string content:string"
```

実行内容（自動）
- proto/handler/usecase/repository(memory)/schema.sql 生成
- buf generate によるコード生成（gen 配下）
- mysqldef で dev/test DB へ適用
- API 再起動 → curl（内蔵の curler サービス）で Create/Get/List を叩いて疎通確認

3) 手動で叩く例（ホストから）

```
curl -sS -X POST -H 'Content-Type: application/json' \
  --data '{"name":"Hello","content":"World"}' \
  http://127.0.0.1:8080/article.v1.ArticleService/CreateArticle
```

---

## 特徴

- 🧅 オニオンアーキテクチャ
  - domain / usecase / adapter を明確に分離
- 🔌 gRPC（connect-go）
  - HTTP/2 + Unary RPC
- 🐳 フル Docker 環境
  - Go API サーバー
  - MySQL 開発 DB
  - MySQL テスト DB（tmpfs）
- 🧪 dev / test DB 完全分離
- 🚀 `scaffold` によるCRUD雛形生成（buf + mysqldef 連携）

---

## Scaffold の使い方（詳細）

- 雛形だけ生成（コード生成・整形まで）

```
make scaffold name=User fields="name:string email:string age:int"
```

- 生成物を作り直したい（同名エンティティのクリーン）

```
make scaffold-clean name=User
```

- proto→Go/Connect 生成のみ手動で実行

```
make generate
```

補足
- 生成直後の配線は memory repository です（DBなしで起動可）。
- MySQL 実装に切り替える場合は `internal/adapter/repository/mysql` を実装して DI に差し替えます。

---

## ディレクトリ構成
```
.
├── cmd/
│ └── server/
│ └── main.go # エントリポイント
├── internal/
│ ├── domain/ # ドメインモデル
│ ├── usecase/ # ユースケース
│ └── adapter/
│ ├── grpc/ # gRPC / connect ハンドラ
│ └── repository/ # 外部依存
│ └── memory/ # 仮実装（後で DB に差し替え）
├── proto/ # gRPC 定義
├── gen/ # buf generate の生成コード
├── docker/ # DB 初期化用（任意）
├── scripts/ # 補助スクリプト
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

---

## アーキテクチャ概要

依存関係は **必ず内向き** になります。
```
[gRPC Handler]
        ↓
    [Usecase]
        ↓
   [Repository IF]
        ↓
[Repository Impl (memory / mysql / ent)]
```


- usecase は DB / gRPC / フレームワークを知らない
- DB や ORM（ent）は adapter に閉じ込める
- 将来の技術変更に強い構成

---

## 必要要件

- Docker
- Docker Compose
- Go（`gonew` 実行用）
- gonew

```bash
go install golang.org/x/tools/cmd/gonew@latest
```

## テンプレートの使い方（gonew）

1. 新規プロジェクト作成

```
gonew github.com/xiao1203/go-onion-grpc-template github.com/yourname/myservice
cd myservice
```

go.mod の module path  
import path  
は自動で置き換えられます。

## Docker 開発環境
### 起動

```
make up

```

起動するサービス：

| サービス       | 説明             |
| ---------- | -------------- |
| api        | Go API サーバー    |
| mysql_dev  | 開発用 DB         |
| mysql_test | テスト用 DB（tmpfs） |

### 停止

```
make down
```

### ログ確認

```
make logs
```

### API コンテナに入る

```
make sh
```

### テスト実行（test DB 使用）

```
make test
```
mysql_test を使用（毎回クリーン） / CI 実行を想定

---

## マイグレーション（mysqldef）

- 適用（通常）

```
make migrate
```

- 破壊的変更（DROP など）も許可して適用

```
make dry-run DROP_FLAGS="--enable-drop"   # まず差分確認
make migrate DROP_FLAGS="--enable-drop"    # 問題なければ適用
```

- 差分だけ確認（適用しない）

```
make dry-run
```

- テストDBを完全リセット

```
make reset-test-db
```

### Docker Compose 構成
### API コンテナ
Go 1.24  
ソースコードを volume マウント  
`go run ./cmd/server` で起動

### MySQL（開発）
永続化 volume 使用  
ホストポート: `13306`

### MySQL（テスト）
tmpfs 使用（永続化しない） / ホストポート: `23306`

### curler（疎通確認用）
`curlimages/curl` ベースの使い捨てコンテナ。`scaffold-all` 実行時に API へ HTTP POST を自動送信します。

--------------
### gRPC について

connect-go を使用  
proto 定義は `proto/` 配下  
buf 設定（`buf.yaml` / `buf.gen.yaml`）を同梱  
`make generate` で protoc/プラグインのローカル導入なしにコード生成可能  

### よくあるコマンドまとめ
```
make up
make down
make logs
make sh
make test
make scaffold name=Article fields="name:string content:string"
make scaffold-all name=Article fields="name:string content:string"
make scaffold-clean name=Article
make generate
make dry-run [DROP_FLAGS="--enable-drop"]
make migrate [DROP_FLAGS="--enable-drop"]
```

### 将来の拡張ポイント
ent（ORM）  
sqldef（DDL 管理）  
buf による proto 自動生成（導入済み）  
wire による DI  
GitHub Actions（CI）  
