# 目的

MCP Server 向け Falco プラグイン開発の将来の開発効率を体系的に高め、CIを安定化し、変更コストを最小化する。

# 前提

- SDK-first / 小さく可逆 / CI優先。
- 既存の nginx プラグイン知見をテンプレ/ツール化して再利用。

# 結論（要旨）

- スキーマ駆動（監査JSON）× フィールド自動生成 × ルールテンプレ × JSONアサーションE2E を標準化。
- ローカル再現性（サンプルイベント/リプレイ）と CI の可視化（Step Summary/アーティファクト）を強化。

# 施策（コンパクト版）

- スキーマ: `docs/schema/mcp_audit_v1.json` を定義。`mcp.*` フィールドの型/説明を単一ソース化。
- コード生成: スキーマ→Go のフィールド定義/バリデーション/変換テーブルを `scripts/gen-fields.sh` で生成。
- ルールテンプレ: 代表4本（未承認/非TLS/大量送信/過剰呼び出し）を `rules/templates/` に整備し、許可リスト/しきい値は外部化。
- サンプルデータ: `test/data/mcp_audit/` に最小セット（正常/異常）を格納。E2E はここから再生。
- JSONアサーション: `jq` ベースの共通関数 `scripts/lib/assert_json.sh` を用意（`.rule`/`.priority`/`output_fields["mcp.*"]`）。
- SKIPポリシー: Falco/権限がない環境は `exit 0` でスキップし、理由を出力。CI失敗を回避。
- 可視化: CI Step Summary に検出率/件数/ルール別内訳を表示。`test-results/` アーティファクトを収集。
- 開発テンプレ: 新規フィールド/ルール追加時の `PRテンプレ` と `変更チェックリスト` を `internal_docs/templates/` に配置。

# 受入基準

- `make fmt && make lint` でクリーン。
- `make test` が変更パッケージで安定通過。E2E は SKIP 方針で安定。
- 代表E2Eで JSON アサーションが再現可能（ローカル/CI 同値）。

# リスク/代替

- スキーマ逸脱: フィールド生成前に `schema check` を CI に追加。逸脱時は失敗。
- テストフレーク: データセット固定化/乱数排除/リトライ上限を導入。
- CI環境差: 依存の明示インストール（`jq bc file`）と GitHub Actions ランナー（ubuntu-latest）固定。

# 次アクション

- `schema/` と `scripts/gen-fields.sh` の骨子作成。
- 代表E2Eスクリプトの雛形と共通アサーション関数を実装。
- PRテンプレ/チェックリストの叩き台を追加。
