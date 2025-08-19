# ドキュメント・インデックス（MCP Server Plugin）

- 推奨順序: 要件 → 効率化 → 計画/次アクション → スキーマ。

## 中核ドキュメント
- 要件: `requirements/MCP_SERVER_PLUGIN_REQUIREMENTS.md`
- 効率化計画: `DEVELOPMENT_EFFICIENCY_PLAN.md`
- 計画: `project/PROJECT_PLAN.md`
- 次アクション: `NEXT_ACTIONS.md`
- スキーマ: `schema/mcp_audit_v1.json`
 - アーキテクチャ: `architecture/OVERVIEW.md`, `architecture/DETAILED.md`

## 実装リソース（雛形/テンプレ）
- ルールテンプレ: `../rules/templates/mcp_baseline_rules.yaml`
- ルール本線: `../rules/mcp_baseline.yaml`
- 生成スクリプト: `../scripts/gen-fields.sh`
- JSONアサーション: `../scripts/lib/assert_json.sh`
- サンプルデータ: `../test/data/mcp_audit/*.json`
- E2Eスクリプト: `../test/e2e/basic/test_mcp_audit_basics.sh`, `../test/e2e/attack/test_mcp_anomalies.sh`
- 設定ガイド: `config/README.md`, `config/EXAMPLE_VALUES.yaml`

## 注意
- `internal_docs/` はローカル参照用（リモート非公開）。公開が必要な要点は本 `docs/` 側に再編集して配置してください。
