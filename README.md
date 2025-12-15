# go-onion-grpc-template

Go 言語で **オニオンアーキテクチャ + gRPC（connect-go）** を採用した  
**フル Docker 開発環境付きテンプレート**です。

`gonew` を使うことで、Rails の `rails new` に近い体験で  
API サーバーの雛形を作成できます。

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
- 🚀 `gonew` による雛形生成

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
├── gen/ # 生成済みコード（テンプレ同梱）
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
mysql_test を使用  
毎回クリーンな DB  
CI 実行を想定した構成  

### Docker Compose 構成
### API コンテナ
Go 1.24  
ソースコードを volume マウント  
`go run ./cmd/server` で起動

### MySQL（開発）
永続化 volume 使用  
ホストポート: `13306`

### MySQL（テスト）
tmpfs 使用（永続化しない）  
ホストポート: `23306`

--------------
### gRPC について

connect-go を使用  
proto 定義は `proto/` 配下  
テンプレでは 生成済みコードを同梱  
protoc / buf 不要  
すぐにビルド・起動可能  

### よくあるコマンドまとめ
```
make up
make down
make logs
make sh
make test
```

### 将来の拡張ポイント
ent（ORM）  
sqldef（DDL 管理）  
buf による proto 自動生成  
wire による DI  
GitHub Actions（CI）  
