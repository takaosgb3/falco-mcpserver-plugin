# 内部ドキュメント・インデックス（MCP Server Plugin）

- 目的: MCP Server向けFalcoプラグインの要件/設計/運用を素早く把握し、関連知見（nginxプラグイン）を再利用できるようにする。
- 前提: 本リポジトリ内の `internal_docs/` 以下のみを対象。詳細は各ドキュメント先で参照。

## 推奨リーディング順
- 要件: `requirements/MCP_SERVER_PLUGIN_REQUIREMENTS.md`（本件の中核）
- 効率化: `DEVELOPMENT_EFFICIENCY_PLAN.md`（将来の開発効率向上）
- 知見: `knowledge/nginx/`（nginxプラグインの設計/CI/テストの再利用）
- 計画: `project/PROJECT_PLAN.md`（ロードマップ/マイルストーン）と `NEXT_ACTIONS.md`（直近タスク）

## 中核ドキュメント
- 要件: `requirements/MCP_SERVER_PLUGIN_REQUIREMENTS.md` — 目的/スコープ/機能・非機能/検知/運用/受入基準/リスク/次アクション。
- 効率化計画: `DEVELOPMENT_EFFICIENCY_PLAN.md` — テンプレ/スキーマ/テストデータ/ツール化/CI導線の整備。

## 実装リソース（雛形/テンプレ）
- スキーマ: `schema/mcp_audit_v1.json`
- ルールテンプレ: `../rules/templates/mcp_baseline_rules.yaml`
- 生成スクリプト: `../scripts/gen-fields.sh`
- JSONアサーション: `../scripts/lib/assert_json.sh`
- サンプルデータ: `../test/data/mcp_audit/*.json`
- E2Eスクリプト: `../test/e2e/basic/test_mcp_audit_basics.sh`, `../test/e2e/attack/test_mcp_anomalies.sh`
- PRテンプレ/チェックリスト: `templates/PULL_REQUEST_TEMPLATE.md`, `templates/CHANGE_CHECKLIST.md`

## 参考知見（nginxプラグインからの再利用）
- アーキテクチャ: `knowledge/nginx/current/ARCHITECTURE.md`
- 開発プロセス: `knowledge/nginx/current/DEVELOPMENT.md`
- テスト設計: `knowledge/nginx/current/TESTING.md`, `knowledge/nginx/development/INTEGRATION_TEST_FRAMEWORK_DESIGN.md`
- CI/CD: `knowledge/nginx/current/CI_CD.md`, `knowledge/nginx/ci-cd/README.md`
- セキュリティ・ルール: `knowledge/nginx/reference/SECURITY_PATTERNS.md`, `knowledge/nginx/reference/NGINX_RULES_REFERENCE.md`
- 運用/トラブルシュート: `knowledge/nginx/operations/troubleshooting.md`
- 配備（自己ホストRunner）: `knowledge/nginx/deployment/SELF_HOSTED_RUNNER_CHECKLIST.md`
- リリース手順: `knowledge/nginx/project/RELEASE_CHECKLIST.md`

## 使い方（知見の転用指針）
- 用語/概念の差分: nginxの「リクエスト/レスポンス/ルール名」を、MCPの「セッション/ツール呼び出し/監査イベント」に読み替える。
- ルール記述: 既存のJSONアサーション/検出率しきい値の考え方をMCPイベントへ適用。
- CI構成: 自己ホストRunner前提・依存明記・SKIPポリシーの流儀を踏襲。
