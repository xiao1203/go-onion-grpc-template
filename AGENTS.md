# Agent Guidelines for go-onion-grpc-template

このリポジトリでエージェント（自動化ツール/AIアシスタント）が作業する際の補足ガイドです。本ファイルのスコープはリポジトリ全体です。

## コミュニケーション
- 応答は日本語で行ってください。

## Scaffold/Clear の振る舞い
- `make scaffold`/`make scaffold-all` はエンティティの雛形を生成します。
- 生成物のクリーンアップは `make clear <Name>` または `make scaffold-clean name=<Name>` を使用してください。
- 修正点: `make clear <Name>`（内部で `scaffold-clean` を呼び出し）により、以下も削除されます。
  - `internal/domain/entity/<snake>.go`
  - `internal/domain/repository/<snake>_repository.go`（ドメイン側リポジトリIF）
  - `internal/usecase/<snake>_usecase.go`
  - `internal/adapter/grpc/<snake>_{handler,routes}.go`
  - `internal/adapter/repository/{memory,mysql}/<snake>_repository.go`
  - `proto/<snake>/` と `gen/<snake>/`
- これにより、`internal/domain/entity/<snake>.go` が残存しないようになっています。

## 注意事項
- `make clear <Name>` に `drop=1` を付与すると、mysqldef の `--enable-drop` を使った DROP の適用も行います（DB スキーマに影響）。
- 重大な変更（`proto` や `db/schema.sql`）を行った場合は、必要に応じて `make protogen` や `make migrate` を実行してください。

## レイヤリング方針（ドメインにRepositoryインターフェースを配置）
- `internal/domain/entity` … エンティティ/値オブジェクト
- `internal/domain/repository` … 永続化境界のインターフェース
- `internal/usecase` … アプリケーションサービス（インタラクタ）。依存先は domain のみ
- `internal/adapter/repository/{mysql,memory}` … domain/repository の実装
- `internal/adapter/grpc` … ハンドラ/ルーティング

Scaffold は `internal/domain/repository/<snake>_repository.go` にIFを生成し、実装は adapter 側に出力します。
