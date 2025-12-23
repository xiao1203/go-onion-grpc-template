# go-onion-grpc-template

Go 言語で **オニオンアーキテクチャ + gRPC（connect-go）** を採用した  
**フル Docker 開発環境付きテンプレート**です。

<img height="700" alt="image" src="https://github.com/user-attachments/assets/f93129b5-568f-404b-8ef7-418eb46bb465" />


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
- proto を生成
- internal/domain/entity（エンティティ）を生成
- internal/domain/repository（ドメイン側のRepositoryインターフェース）を生成
- internal/usecase を生成（domain/repository に依存）
- internal/adapter/repository/mysql（実装）を生成
- internal/adapter/grpc/{handler,routes} を生成（レジストリ登録）
- db/schema.sql を追記
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
- 🧰 ORM: GORM（MySQL）
- ❗ エラー: [ergo](https://github.com/newmo-oss/ergo) を採用（コード付与 + スタック保持）
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
make protogen
```

補足
- 生成直後の配線は MySQL repository です（DBに永続化）。
- ルーティング登録はレジストリ方式です。scaffold は `internal/adapter/grpc/<entity>_routes.go` を生成し、`init()` で登録します（`main.go` は編集しません）。
- メモリ実装はオプションです。必要な場合のみ以下のいずれかで生成してください。
  - `make scaffold name=User fields="..." mem=1`
  - もしくは `go run ./cmd/scaffold -name User -fields "..." -with-memory`

### Fields（対応型）
- 指定例: `make scaffold name=Device fields="name:string level:int8 code:uint8 serial:uint32 big:uint64 ok:bool note:text"`
- サポート型（左: 指定値 → 右: Proto/Go/SQL）
  - `string` → string / string / VARCHAR(255)
  - `text` → string / string / TEXT
  - `bool` → bool / bool / TINYINT(1)
  - `int`, `int32` → int32 / int32 / INT
  - `int8` → int32 / int32 / TINYINT
  - `int64` → int64 / int64 / BIGINT
  - `uint8` → uint32 / uint32 / TINYINT UNSIGNED
  - `uint32` → uint32 / uint32 / INT UNSIGNED
  - `uint64` → uint64 / uint64 / BIGINT UNSIGNED

注意
- Protobufにはint8/uint8の直接型がないため、`int8` は `int32`、`uint8` は `uint32` として表現します（Go/SQLは上記の通り）。
- 予約語（`text`/`order`/`group`/`value`）はフィールド名に使用できません。別名（例: `value_col`）に変更してください。

---

## ディレクトリ構成
```
.
├── cmd/
│ └── server/
│ └── main.go # エントリポイント
├── internal/
│ ├── domain/ # ドメイン層
│ │   ├── entity/       # エンティティ / 値オブジェクト
│ │   └── repository/   # 永続化境界（Repositoryインターフェース）
│ ├── usecase/ # ユースケース（アプリケーションサービス）
│ └── adapter/
│   ├── grpc/ # gRPC / connect ハンドラ + ルート登録（registry）
│   │   ├── registry.go # レジストリ本体（Add / RegisterAll）
│   │   └── <entity>_{handler|routes}.go # scaffold 生成
│   └── repository/ # 外部依存
│       └── memory/ # 仮実装（後で DB に差し替え）
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
   [Repository IF (domain)]
        ↓
[Repository Impl (memory / mysql / ent)]
```


- usecase は DB / gRPC / フレームワークを知らない（domain のみ依存）
- Repository インターフェースは domain 配下（internal/domain/repository）に配置
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
mysqldef は安全運用のため、DROP を伴う破壊的変更をデフォルトでは実行しません。  
そのため、DROP を伴う変更を schema.sql に加えた場合は、以下のように `DROP_FLAGS="--enable-drop"` を指定して実行してください。

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
### gRPC / ルーティングについて

connect-go を使用  
proto 定義は `proto/` 配下  
buf 設定（`buf.yaml` / `buf.gen.yaml`）を同梱  
`make protogen` で protoc/プラグインのローカル導入なしにコード生成可能  

ルーティング登録はレジストリ方式です。`cmd/server/main.go` は以下のみ行います。

- MySQL接続の初期化（1回、GORM使用: `internal/infra/mysql.OpenGormFromEnv`）
- `grpcadapter.RegisterAll(mux, grpcadapter.Deps{Gorm: db})` の呼び出し

各エンティティは `internal/adapter/grpc/<entity>_routes.go` に registrar が生成され、`init()` でレジストリへ登録されます。
このため、`main.go` を手で編集する必要はありません（scaffold/clear による編集も不要）。

### 手動でAPIを作る（scaffoldを使わない場合）

最小手順は以下です。

1. Protoを追加: `proto/<entity>/v1/<entity>.proto`
   - `option go_package = "<your-module>/gen/<entity>/v1;<entity>v1"` を忘れずに設定
2. 生成: `make protogen`（`make proto` でも可）
3. Usecase実装: `internal/usecase/<entity>_usecase.go`
4. Repository実装（GORM）: `internal/adapter/repository/mysql/<entity>_repository.go`
5. Handler実装: `internal/adapter/grpc/<entity>_handler.go`
6. ルート登録: `internal/adapter/grpc/<entity>_routes.go`（initでレジストリにAdd）
7. DDL追加: `db/schema.sql` にCREATE TABLEを追記
8. マイグレーション: `make migrate`
9. 再起動: `make restart`
10. 疎通: curl で `/<pkg>.<ver>.<Service>/<Method>` をPOST

scaffoldはこの手順を自動化しています。手作業で進めたい場合は上記を参考にしてください。

### よくあるコマンドまとめ
```
make up
make down
make logs
make restart
make sh
make test
make scaffold name=Article fields="name:string content:string"
make scaffold-all name=Article fields="name:string content:string"
make scaffold-clean name=Article
make clear Article [drop=1]  # 生成物とschemaの該当ブロックを削除。drop=1でDBにDROP適用
make protogen
make dry-run [DROP_FLAGS="--enable-drop"]
make migrate [DROP_FLAGS="--enable-drop"]
```

### エラー方針（ergo）
- 本テンプレートはエラーライブラリとして [newmo-oss/ergo](https://github.com/newmo-oss/ergo) を利用します。
  - アプリ共通のエラーコードは `internal/apperr` に集約し、`ergo.WithCode` でエラーにコードを付与します。
  - gRPC（connect-go）への変換は `apperr.ToConnect(err)` を使用します（`ergo.CodeOf(err)` に応じて `connect.Code*` にマッピング）。
- よくある使い方
  - 新規作成: `ergo.New("something bad happened")`
  - ラップ: `ergo.Wrap(err, "while saving")`
  - コード付与: `ergo.WithCode(err, apperr.Internal)`
  - ハンドラ返却: `return nil, apperr.ToConnect(err)`

任意: 静的解析（ergocheck）
- 必要に応じて、ergo同梱の静的解析器「ergocheck」を導入できます（errors.New や fmt.Errorf の使用、フォーマット文字列の誤用などを検出）。
- ergocheckはビルド時の実行挙動には影響せず、lint/CI のフェーズで規約違反を検出して失敗させる用途です。
- 導入は任意です（テンプレートでは同梱していません）。プロジェクト方針に合わせて golangci-lint などへの組み込みをご検討ください。

### clear の動作（レジストリ方式）
- 削除対象
  - `proto/<entity>` / `gen/<entity>`
  - `internal/domain/entity/<entity>.go`
  - `internal/domain/repository/<entity>_repository.go`
  - `internal/usecase/<entity>_usecase.go`
  - `internal/adapter/grpc/<entity>_{handler,routes}.go`
  - `internal/adapter/repository/{memory,mysql}/<entity>_repository.go`
  - `db/schema.sql` の対象テーブルの CREATE TABLE ブロックと見出しコメント
- 備考
  - `main.go` は編集しません（レジストリ方式のため不要）
  - DBにDROPを適用する場合は `make clear <Name> drop=1`（内部で `mysqldef --enable-drop` を実行）

### 将来の拡張ポイント
GitHub Actions（CI）  

---

## 変更履歴

タグ/バージョンごとの詳細は CHANGELOG.md を参照してください。
