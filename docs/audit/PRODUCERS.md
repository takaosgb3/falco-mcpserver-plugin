# 目的

Claude Code / Codex CLI など MCP クライアントのアクセスを、eBPF/カーネルモジュールなしでリアルタイムに可観測化する監査イベント（JSON行）生成手段を定義する。

# ポリシー（Phase 1）

- eBPF/カーネルモジュールは使用しない。
- 収集はメタデータ中心（メソッド名/バイト数/回数/接続先）。コンテンツ本文は収集しない。

# 選択肢と設計

- StdIO ラッパ（ローカルMCPサーバー向け）
  - 方式: `mcp-audit-wrap -- <real-mcp-server> <args...>` でサーバープロセスを子として起動し、stdin/stdout を中継。
  - 取得: 行フレーム（JSON-RPC/NDJSON）を監視し、`.method`、フレーム長、方向、時刻を計測。
  - 出力: 監査 JSON（`docs/schema/mcp_audit_v1.json` 準拠）を `stdout` またはファイルへ追記。
  - 設定例（Claude Code）: VS Code MCP サーバー設定の `command` を `mcp-audit-wrap` に、`args` で実サーバーを後段へ。

- WebSocket ローカルプロキシ（リモートMCPサーバー向け）
  - 方式: `mcp-audit-proxy --listen :8989 --target wss://mcp.example.com` をローカルで待受し、クライアントのURLを `ws://127.0.0.1:8989` に置換。
  - 取得: ws フレーム長/方向/時刻/接続先メタ。TLSは終端せず透過（メソッドは可能ならJSON-RPC層で抽出）。
  - 出力: 監査 JSON を `stdout`/ファイルへ追記。

- クライアント内エミッタ（Codex CLI向け推奨）
  - 方式: MCP 呼び出し後にメタデータを JSON 行で吐くフックを組み込み。出力先は環境変数で制御。
  - 出力先: `MCP_AUDIT_SINK=/path/to/audit.jsonl`（未指定時は `stdout`）。

# CLI 仕様（提案）

- mcp-audit-wrap
  - Usage: `mcp-audit-wrap [--sink <path>|stdout] [--session-id <id>] -- <cmd> [args...]`
  - 動作: `<cmd>` を起動し、双方向に中継。各リクエスト/レスポンスのメタを集計して随時出力。
  - 出力フィールド例: `mcp.session_id`, `mcp.client_process`, `mcp.server_host`, `mcp.server_port`, `mcp.tls`, `mcp.method`, `mcp.request_bytes`, `mcp.response_bytes`, `mcp.tool_invoke_count`, `mcp.file_access_count`, `mcp.timestamp`。

- mcp-audit-proxy
  - Usage: `mcp-audit-proxy --listen <addr> --target <wss://...> [--sink <path>|stdout] [--no-method-parse]`
  - 動作: listen→target へ完全透過転送。フレーム長/方向/時刻/接続メタを記録。`--no-method-parse` でアプリ層を覗かない運用も可能。

- Codex CLI エミッタ
  - Env: `MCP_AUDIT_SINK`, `MCP_AUDIT_LEVEL=meta`（メタのみ）、`MCP_SESSION_ID`（任意指定）。
  - 動作: 呼び出しごとに 1 行 JSON を追記。

# 監査JSON（例）

```json
{
  "schema_version": "1",
  "timestamp": 1724050000000000000,
  "session_id": "sess-123",
  "client_process": "claude-code",
  "server_host": "localhost",
  "server_port": 8080,
  "tls": true,
  "method": "tools.list",
  "request_bytes": 512,
  "response_bytes": 2048,
  "tool_invoke_count": 2,
  "file_access_count": 0
}
```

# セキュリティ/プライバシー

- デフォルトで本文を収集しない。認証情報や識別子はハッシュ/マスク可能（将来の `redaction` 設定）。
- sink ファイルの権限は 0600 を推奨。ローテーション/保持期間を明示。

# 次アクション

- `mcp-audit-wrap` と `mcp-audit-proxy` の雛形を作成（Go）。
- Codex CLI 側にエミッタのフックを実装（もしくはフックポイントの調査）。
- Falco プラグイン入力に監査JSONを供給し、`rules/mcp_baseline.yaml` で検証。

