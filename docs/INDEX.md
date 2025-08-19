# ドキュメント・インデックス（MCP Server Plugin）

- 推奨順序: 要件 → 効率化 → 計画/次アクション → スキーマ。

## 中核ドキュメント
- 要件: [requirements/MCP_SERVER_PLUGIN_REQUIREMENTS.md](requirements/MCP_SERVER_PLUGIN_REQUIREMENTS.md) — 目的/スコープ/機能・非機能/検知/受入基準。
- 効率化計画: [DEVELOPMENT_EFFICIENCY_PLAN.md](DEVELOPMENT_EFFICIENCY_PLAN.md) — スキーマ駆動/コード生成/E2E/CI可視化の方針。
- 計画: [project/PROJECT_PLAN.md](project/PROJECT_PLAN.md) — フェーズ/マイルストーン/タイムライン/リスク。
- フェーズ1タスク: [project/PHASE1_TASKS.md](project/PHASE1_TASKS.md) — フェーズ1の詳細タスク/受入/検証/DoD。
- フェーズ2タスク: [project/PHASE2_TASKS.md](project/PHASE2_TASKS.md) — テストサーバ/シナリオ/CI可視化/Producer拡張の詳細。
- 次アクション: [NEXT_ACTIONS.md](NEXT_ACTIONS.md) — 直近タスクと受入基準（実行優先度つき）。
- スキーマ: [schema/mcp_audit_v1.json](schema/mcp_audit_v1.json) — 監査イベントのJSONスキーマ（v1）。
- アーキテクチャ: [architecture/OVERVIEW.md](architecture/OVERVIEW.md) — 高レベル構成とフロー（MECE）。
- アーキテクチャ詳細: [architecture/DETAILED.md](architecture/DETAILED.md) — モジュール/API/メモリ/設定/テスト詳細。

## 実装リソース（雛形/テンプレ）
- ルールテンプレ: [../rules/templates/mcp_baseline_rules.yaml](../rules/templates/mcp_baseline_rules.yaml) — ベースライン検知のテンプレ集。
- ルール本線: [../rules/mcp_baseline.yaml](../rules/mcp_baseline.yaml) — 実運用を想定した本線ルール。
- 生成スクリプト: [../scripts/gen-fields.sh](../scripts/gen-fields.sh) — フィールド定義/変換テーブルのコード生成（雛形）。
- JSONアサーション: [../scripts/lib/assert_json.sh](../scripts/lib/assert_json.sh) — `.rule`/`output_fields` 等の検証ヘルパー。
- サンプルデータ: [../test/data/mcp_audit/](../test/data/mcp_audit/) — 正常/異常サンプルJSON。
- E2Eスクリプト: [../test/e2e/basic/test_mcp_audit_basics.sh](../test/e2e/basic/test_mcp_audit_basics.sh), [../test/e2e/attack/test_mcp_anomalies.sh](../test/e2e/attack/test_mcp_anomalies.sh) — オフライン/スキップ対応の基本検証。
- 設定ガイド: [config/README.md](config/README.md), [config/EXAMPLE_VALUES.yaml](config/EXAMPLE_VALUES.yaml) — 許可リスト/しきい値/レッドアクト例。
- 監査Producer: [audit/PRODUCERS.md](audit/PRODUCERS.md) — stdio/WS/client-emitter の設計・起動・シーケンス。
- CI設計: [ci/CI_DESIGN.md](ci/CI_DESIGN.md) — トリガー/ジョブ/ゲート/可視化の設計。
- テスト用MCP: [testing/MCP_TEST_SERVER_DESIGN.md](testing/MCP_TEST_SERVER_DESIGN.md), [testing/MCP_TEST_SCENARIOS.md](testing/MCP_TEST_SCENARIOS.md) — テストサーバ設計と具体シナリオ。

## 注意
- `internal_docs/` はローカル参照用（リモート非公開）。公開が必要な要点は本 `docs/` 側に再編集して配置してください。
