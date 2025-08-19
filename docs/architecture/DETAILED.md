# 目的

MCP プラグインの詳細設計（モジュール/インタフェース/メモリ/設定/テスト/運用）を定義し、実装の一貫性と将来拡張性を担保する。

# 前提

- SDK-first、`evt.source=mcp_audit`、`mcp.*` フィールド、JSON アサーション E2E、StringPool 準拠。
- 公開ドキュメントは `docs/`、ローカル参照は `internal_docs/`（非公開）。

# モジュール構成（提案）

- `cmd/plugin-sdk/`（将来）: SDK エントリポイント・初期化/ライフサイクル（open/next/close）。
- `pkg/plugin/`（将来）: 抽出器/フィールド定義・イベント生成・メモリ管理委譲。
- `pkg/parser/`（将来）: 監査 JSON のパース/検証（`docs/schema/mcp_audit_v1.json`）。
- `pkg/config/`（将来）: YAML 設定（許可リスト/しきい値/レッドアクト）とホットリロード方針（任意）。
- `pkg/generated/`: `scripts/gen-fields.sh` によるフィールド定義/変換テーブルの自動生成物。
- `rules/`: ベースライン/テンプレート（`templates/mcp_baseline_rules.yaml` → 本線化）。
- `scripts/`: コード生成・JSON アサーション共通関数・E2E 補助。
- `test/`: サンプル JSON・E2E スクリプト（SKIP 対応）。

# プラグイン I/F（概念）

- 情報系 API: `plugin_get_name`, `plugin_get_version`, `plugin_get_description`, `plugin_get_fields`（SDK がラップ）。
- ソース系 API: `plugin_init`, `plugin_open`, `plugin_next`, `plugin_close`。
- イベント: 1 行 JSON → 正規化構造体 → `mcp.*` へマッピング → Falco バッファへ（StringPool 文字列）。

# フィールド/型（例）

- string: `mcp.session_id`, `mcp.client_process`, `mcp.server_host`, `mcp.method`, `mcp.auth_scheme`
- int: `mcp.server_port` (u16), `mcp.request_bytes`/`mcp.response_bytes` (u64), `mcp.tool_invoke_count`/`mcp.file_access_count` (u32)
- bool: `mcp.tls`
- time: `mcp.timestamp` (ns)

# メモリ/CGO

- 文字列: Falco 所有/プラグイン所有の原則に従い、戻り値は `C.CString`、抽出結果は StringPool 管理でバッチ毎にクリア。
- ポインタ: C ポインタを跨って保持しない。Go GC と `free` を混在させない。

# エラー処理/堅牢性

- 壊れた JSON/必須欠損はイベントをスキップし、件数をメトリクス化。プラグインは継続。
- 設定不備はデフォルト緩和（許可ホスト空=検知弱め）と明示ログ。

# 設定/運用

- YAML に許可ホスト/しきい値/レッドアクト方針。環境別オーバレイ（例: `values.local.yaml`）。
- ルールはテンプレ＋環境値で生成可能な構成に（将来）。

# 設定例（公開用）

```yaml
# docs/config/EXAMPLE_VALUES.yaml
allowlist:
  hosts:
    - localhost
    - 127.0.0.1
    - mcp.internal.local
thresholds:
  request_bytes_warn: 10485760    # 10 MiB
  response_bytes_warn: 10485760   # 10 MiB
  tool_invoke_warn: 100
redaction:
  mask_auth: true
  hash_identifiers: true
```

# ルール構成（本線）

- `rules/mcp_baseline.yaml` にベースルール4本を配置。
- 依存フィールドは `mcp.*` に限定（未実装の評価ロジックは持ち込まない）。

```yaml
# rules/mcp_baseline.yaml の骨子（抜粋）
list:
  allowed_mcp_hosts:
    - localhost
    - 127.0.0.1
    - mcp.internal.local
macros:
  - name: mcp_unapproved_host
    condition: not (mcp.server_host in (allowed_mcp_hosts))
rules:
  - rule: MCP Unapproved Endpoint Access
    condition: evt.source = mcp_audit and mcp_unapproved_host
    priority: WARNING
    source: mcp_audit
  - rule: MCP Insecure TLS
    condition: evt.source = mcp_audit and mcp.tls=false
    priority: NOTICE
    source: mcp_audit
```

# 出力例（Falco JSON）

```json
{
  "rule": "MCP Unapproved Endpoint Access",
  "priority": "Warning",
  "output_fields": {
    "mcp.server_host": "unknown.example.com",
    "mcp.server_port": 443,
    "mcp.client_process": "claude-code",
    "mcp.session_id": "sess-999"
  }
}
```

# E2E アサーション断片

```bash
. scripts/lib/assert_json.sh
assert_json_field artifact.json '.rule' 'MCP Unapproved Endpoint Access'
assert_json_contains artifact.json '.output_fields."mcp.server_host"' 'unknown.example.com'
```

# ルール連携

- `source: mcp_audit`、`output_fields["mcp.*"]` を出力。
- ベース: 未承認エンドポイント/非TLS/大量送信/過剰呼び出し。検出率しきい値を E2E で評価。

# 可観測性/CI

- JSON 出力 + `jq` アサーション（`.rule/.priority/output_fields["mcp.*"]`）。
- `test-results/` 収集、Step Summary に統計を表示（将来）。
- SKIP 方針で環境未整備時も CI を赤くしない。

# テスト戦略

- ユニット: 変換テーブルのテーブル駆動、境界値/欠損/壊れた JSON。
- E2E: サンプル JSON の再生→ルール発火の JSON アサーション。
- ベンチ（任意）: 抽出/変換のスループット・アロケーション測定。

# 拡張（将来）

- eBPF/システムコール観測: ネットワーク/ファイル I/O の補完、PID/コマンドラインの相関。
- プロキシ連携: TLS 指紋/ミラーリング/レイテンシ計測。
- 運用自動化: 日次レポート/傾向検知/許可リスト自動提案。

# 受入基準/次アクション

- 受入: ユニット/E2E/CI が安定、ルール最小セットが機能、後方互換を維持。
- 次アクション: SDK 雛形作成、コード生成強化、ルール本線化、CI 可視化の実装。
