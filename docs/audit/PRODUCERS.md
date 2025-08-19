# 目的

Claude Code / Codex CLI など MCP クライアントのアクセスを、eBPF/カーネルモジュールなしでリアルタイムに可観測化する監査イベント（JSONL）生成手段を、仕組み・起動方法・データフローまで具体的に解説する。

# ポリシー（Phase 1）

- eBPF/カーネルモジュールは使用しない。
- 収集はメタデータ中心（メソッド名/バイト数/回数/接続先）。コンテンツ本文は収集しない。

# 監査Producerの3方式（MECE）

1) StdIO ラッパ（ローカルMCPサーバー向け）
- 何をするか: MCPサーバーを子プロセスとして起動し、親（エディタ/CLI）との stdin/stdout を中継。その際に「行単位ストリーム（JSON-RPC/NDJSON）」を観測してサイズ・方向を計測。必要に応じて`.method`を抽出。
- どう起動するか:
  - 実サーバーを `--` の後に渡す。
  - 例: `mcp-audit-wrap --sink audit.jsonl --session-id sess-123 --client-process claude-code -- /path/to/real-mcp --transport stdio`
- データの流れ:
  1. エディタ/CLI → ラッパ stdin（リクエスト）
  2. ラッパ → サーバー stdin（転送）
  3. サーバー stdout → ラッパ（レスポンス）
  4. ラッパ → エディタ/CLI stdout（転送）
  5. ラッパが転送量/方向/時刻を集計し、監査JSONLを sink へ書き出し

2) WebSocket ローカルプロキシ（リモートMCPサーバー向け）
- 何をするか: ローカルで `ws://` を待ち受け、ターゲット `wss://` へ透過接続。フレーム長・方向・接続メタを計測（TLSは終端しない）。可能ならJSON-RPCレイヤで`.method`抽出。
- どう起動するか:
  - 例: `mcp-audit-proxy --listen :8989 --target wss://mcp.example.com --sink audit.jsonl --client-process claude-code --assume-tls`
  - エディタ/CLI の接続URLを `ws://127.0.0.1:8989` に切替。
- データの流れ:
  1. エディタ/CLI → ローカル `ws://127.0.0.1:8989`
  2. プロキシ → `wss://mcp.example.com`（透過転送）
  3. 双方向のフレーム長/方向/時刻を計測し、監査JSONLを書き出し

3) クライアント内エミッタ（Codex CLI向け推奨）
- 何をするか: MCPクライアントの呼び出し直後（または前後）に監査JSONを直接書き出す。外部挟み込み不要・最小オーバーヘッド。
- どう起動/組み込むか:
  - 実装側で、各RPC呼び出し箇所にフック（AfterCall）を挿入。
  - 出力先は環境変数で制御（例: `MCP_AUDIT_SINK=/path/to/audit.jsonl`）。未指定時は stdout に出力。
- データの流れ:
  1. CLI が MCP リクエストを発行
  2. レスポンス受領後、メタ情報（method/bytes/host/tls/時刻など）を Event に詰める
  3. JSON1行を sink に追記（即時フラッシュ）

## 図（方式別シーケンス）

StdIO ラッパ:
```
Editor/CLI ──request──▶ mcp-audit-wrap ──request──▶ Real MCP (stdio)
Editor/CLI ◀─response── mcp-audit-wrap ◀─response── Real MCP (stdio)
                       └─ Emit JSONL (bytes/method/timestamp) → sink
```

WS ローカルプロキシ:
```
Editor/CLI ──ws://127.0.0.1:8989──▶ mcp-audit-proxy ──wss://mcp.example.com──▶ MCP Server
Editor/CLI ◀─────────────────────────────────────────────────────────────────────── MCP Server
                              └─ Emit JSONL (frames/timing/host/tls) → sink
```

クライアント内エミッタ:
```
CLI (MCP client) ──call──▶ MCP Server
CLI (hook AfterCall) ── Emit JSONL (method/bytes/host/tls/timestamp) → sink
```

# CLI 仕様（詳細）

`mcp-audit-wrap`
- Usage: `mcp-audit-wrap [--sink <path>|stdout] [--session-id <id>] [--client-process name] -- <cmd> [args...]`
- 動作: `<cmd>` を起動し、親↔子の stdin/stdout を完全中継。中継した合計バイト数（req/resp）等を集計し、イベントを発火。
- 出力フィールド（例）: `mcp.session_id`, `mcp.client_process`, `mcp.server_host=stdio`, `mcp.tls=false`, `mcp.request_bytes`, `mcp.response_bytes`, `mcp.timestamp` 他。

`mcp-audit-proxy`
- Usage: `mcp-audit-proxy --listen <addr> --target <wss://...|host:port> [--sink <path>|stdout] [--client-process name] [--assume-tls]`
- 動作: TCPレベルで透過中継（ws/wssどちらも可）。フレーム長/方向/時刻・接続先メタを記録。TLSは終端しない（プライバシー配慮）。
- 出力フィールド（例）: `mcp.server_host`, `mcp.server_port`, `mcp.tls`, `mcp.request_bytes(=up)`, `mcp.response_bytes(=down)`。

クライアント内エミッタ（実装ガイド）
- Env: `MCP_AUDIT_SINK`, `MCP_AUDIT_LEVEL=meta`, `MCP_SESSION_ID`
- 擬似コード（Go）:
```go
w, _ := audit.NewWriter(os.Getenv("MCP_AUDIT_SINK"))
defer w.Close()
// before call: ts0 := time.Now()
resp, err := mcp.Call(ctx, method, params)
// after call: measure sizes (reqLen, respLen) from buffers/encoders
_ = w.Emit(audit.Event{
  SessionID: os.Getenv("MCP_SESSION_ID"),
  ClientProcess: "codex-cli",
  ServerHost: host, ServerPort: port, TLS: tlsOn,
  Method: method,
  RequestBytes: uint64(reqLen), ResponseBytes: uint64(respLen),
})
```

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

# 導入手順（例）

- Claude Code（ローカルMCP）: VS Code の MCP サーバ設定で `command` を `mcp-audit-wrap` にし、`args` で実サーバーを `--` の後に指定。
- Claude Code（リモートMCP）: MCP URL を `ws://127.0.0.1:8989` に変更し、`mcp-audit-proxy` を起動して `--target wss://...` に中継。
- Codex CLI: クライアント内エミッタを組み込み、`MCP_AUDIT_SINK` を設定して起動。

# セキュリティ/プライバシー

- デフォルトで本文を収集しない。認証情報や識別子はハッシュ/マスク可能（将来の `redaction`）。
- sink ファイルの権限は 0600 を推奨。ローテーション/保持期間を明示。

# 検証のコツ

- まずは `--sink stdout` で動作を確認し、JSON行が見えることを確認。
- `jq` でフィールド検査（`.method`, `.request_bytes` など）。
- Falco 連携は、生成された JSONL をプラグイン入力に与え、`rules/mcp_baseline.yaml` の発火を JSON アサーションで確認。

