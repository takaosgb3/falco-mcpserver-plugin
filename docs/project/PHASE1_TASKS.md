# 目的

フェーズ1（MVP: 観測最小・検知最小）のタスクを作業単位で明確化し、受入基準・検証手順・完了条件を統一する。

# スコープ

- 入力: 監査JSON（`docs/schema/mcp_audit_v1.json`）を単一ソースとする。
- 出力: `mcp.*` フィールドでのFalcoイベント、ベースルール4本の発火確認。
- 除外: eBPF/カーネルモジュール、重い本格E2E（Phase 2以降）。

# タスク詳細（Workstreams）

- スキーマ定義（schema v1）
  - 目的: 監査イベントの最小スキーマを確定。
  - 作業:
    - [x] `docs/schema/mcp_audit_v1.json` 作成（必須フィールド: session_id/client/server/tls/method/bytes/timestamp）。
    - [x] `docs/requirements/MCP_SERVER_PLUGIN_REQUIREMENTS.md` と整合。
  - 受入: JSONサンプルが `jq -e` で基本検証を通過。

- 監査Producer（MVP）
  - 目的: 実運用に近いI/Oでメタデータを収集（no eBPF）。
  - 作業:
    - [x] `cmd/mcp-audit-wrap`（stdio中継・双方向バイト計測）
    - [x] `cmd/mcp-audit-proxy`（tcp/ws透過・双方向バイト計測・TLS透過）
    - [x] `pkg/audit/event.go`（JSONL出力ヘルパ）
  - 検証: ローカルで `--sink stdout` 動作確認、JSON行が出力される。
  - メモ: `.method`抽出/per-call発火はPhase2で拡張。

- ルール（ベースライン4本）
  - 目的: 最低限の検知（未承認/非TLS/大量/過剰）。
  - 作業:
    - [x] テンプレ: `rules/templates/mcp_baseline_rules.yaml`
    - [x] 本線: `rules/mcp_baseline.yaml`
  - 受入: E2Eのサンプルで想定フィールドが確認できること（JSONアサーション）。

- CI（初期）
  - 目的: GitHub Actionsでビルドと軽量E2Eを安定実行。
  - 作業:
    - [x] `.github/workflows/ci.yml`（PR/手動トリガ、build + offline E2E）
    - [x] スクリプト: `test/e2e/basic/...`, `test/e2e/attack/...`, `scripts/lib/assert_json.sh`
  - メモ: Step Summary/アーティファクトはPhase2で導入。

- ドキュメント
  - 目的: 認知負荷を下げ、導線を明確化。
  - 作業:
    - [x] `docs/architecture/OVERVIEW.md`（MECEで再構成）
    - [x] `docs/architecture/DETAILED.md`（設定例・出力例）
    - [x] `docs/audit/PRODUCERS.md`（方式/起動/シーケンス）
    - [x] `docs/INDEX.md`（クリック可能リンク＋1行サマリ）

# 受入基準（Phase 1）

- ビルド: `go build ./cmd/mcp-audit-wrap`, `go build ./cmd/mcp-audit-proxy` が成功。
- サンプル検証: `test/data/mcp_audit/*.json` を `jq` で検証し、E2Eスクリプトが `[PASS]`/`[SKIP]` で安定。
- ルール: 代表サンプルで `mcp.*` フィールドが想定通り、ベースルールに対応する値が生成される。
- CI: PR/手動トリガでワークフローが完走（Build成功、E2EはSKIP/またはPASS）。

# 検証手順（抜粋）

- ローカル簡易チェック:
  - `go build ./cmd/mcp-audit-wrap && ./mcp-audit-wrap --sink stdout -- /bin/cat < /dev/null`
  - `go build ./cmd/mcp-audit-proxy && ./mcp-audit-proxy --listen :8989 --target 127.0.0.1:80 --sink stdout`
- サンプル/スクリプト:
  - `bash test/e2e/basic/test_mcp_audit_basics.sh`
  - `bash test/e2e/attack/test_mcp_anomalies.sh`

# リスク/注意

- `.method`未抽出: Phase1はバイト計測中心。詳細はPhase2で拡張。
- CIノイズ: Pushトリガは無効化済。PR/手動での検証を基本に。
- プライバシー: デフォルトでメタデータのみ。ログの権限/保持期間に注意。

# 完了の定義（DoD）

- チェックリスト（全てYes）
  - [ ] スキーマv1で最低フィールドが揃い、サンプルが検証済み
  - [ ] wrap/proxyがビルド成功し、`--sink stdout` でJSON行が出力
  - [ ] ルール本線が配置され、サンプルで期待フィールドが確認できる
  - [ ] CI（PR/手動）でワークフローが安定完走
  - [ ] 主要ドキュメント（INDEX/OVERVIEW/PRODUCERS）が最新
