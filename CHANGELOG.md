# Changelog

このプロジェクトのタグ/バージョンごとの変更点と移行手順を管理します。最新が上に来るように追記してください。

記入のコツ:
- Highlights: 主要な追加/変更の要約（1〜2行）
- Breaking Changes: 互換性に影響する変更（なければ "-"）
- Migration: 必要な移行手順（なければ "-"）
- Notes: 既知の注意点や補足

| Tag | Date | Highlights | Breaking Changes | Migration | Notes |
| --- | --- | --- | --- | --- | --- |
| v0.2.0 | 2024-12-20 | GORM採用のMySQL Repositoryをデフォルト化。レジストリベースDIへ移行。`make protogen`/`make restart` を導入。 | main.goへの直編集を廃止（レジストリ自動登録に統一）。`make generate` → `make protogen` に改名。 | `make protogen` → `make migrate` → `make restart`。旧Greeterサンプルは削除済み。 | `proto/` 直下のgo_packageは gonew 後の module を指すよう注意。 |
| v0.1.0 | 2024-12-10 | 初版リリース。CRUD scaffold、Docker一式、mysqldef、buf、connect-goを同梱。 | - | `make up` で起動、`make scaffold-all` で雛形生成。 | - |
