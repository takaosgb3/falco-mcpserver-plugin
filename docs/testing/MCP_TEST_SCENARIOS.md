# 目的

テスト用 MCP Server を用いた実行シナリオを、ルール/フィールド/期待結果と合わせて具体化する。開発者が同じ手順で再現できることを重視する。

# 前提

- 監査Producer: `mcp-audit-wrap`（stdio）/ `mcp-audit-proxy`（ws）/ client-emitter（任意）。
- ルール: `rules/mcp_baseline.yaml`。
- サンプル: `test/data/mcp_audit/*.json`（補助）。

# シナリオ一覧（MVP）

- S1: 未承認エンドポイント検出
  - 目的: `mcp.server_host` が許可外の場合にルール発火。
  - 手順:
    - wsモードで mcp-test-server を `--listen :8443 --tls-cert/-key` で起動（ホスト名: `unknown.example.com` をDNS/hostsで割当 or 代替としてプロキシ側で `--target wss://unknown.example.com:8443`）。
    - `mcp-audit-proxy --listen :8989 --target wss://unknown.example.com:8443 --sink audit.jsonl --assume-tls`。
    - エディタ/CLI を `ws://127.0.0.1:8989` に接続、`tools.list` を1回実行。
  - 期待: 監査JSONの `mcp.server_host=unknown.example.com`、ルール「MCP Unapproved Endpoint Access」が発火。

- S2: 非TLS通信の検出
  - 目的: `mcp.tls=false` の場合にルール発火。
  - 手順:
    - `mcp-test-server --mode ws --listen :8080`（平文）。
    - `mcp-audit-proxy --listen :8989 --target ws://127.0.0.1:8080 --sink audit.jsonl`。
    - `ws://127.0.0.1:8989` に接続し `tools.list` 実行。
  - 期待: 監査JSONの `mcp.tls=false`、ルール「MCP Insecure TLS」が発火。

- S3: 大量データ送信の検出
  - 目的: `mcp.request_bytes` または `mcp.response_bytes` がしきい値超過でルール発火。
  - 手順:
    - `mcp-test-server --mode ws --listen :8080 --delay-ms 1 --resp-size-bytes 15728640`（約15MiB）。
    - `mcp-audit-proxy --listen :8989 --target ws://127.0.0.1:8080 --sink audit.jsonl`。
    - `tools.exec` を1回実行。
  - 期待: 監査JSONの `mcp.response_bytes >= 10485760`、ルール「MCP Excessive Data Transfer」が発火。

- S4: 過剰呼び出しの検出
  - 目的: セッション内の `mcp.tool_invoke_count` がしきい値超過でルール発火。
  - 手順:
    - `mcp-test-server --mode stdio --burst-calls 120`。
    - `mcp-audit-wrap --sink audit.jsonl --session-id sess-1 --client-process codex-cli -- /path/to/mcp-test-server --mode stdio ...`。
    - クライアントから `tools.exec` を1回発行（burstにより内部で120回相当）。
  - 期待: 監査JSONの `mcp.tool_invoke_count > 100`、ルール「MCP Excessive Tool Invocations」が発火。

# アサーション（JSON）

- 例（S1の発火確認断片）:
```bash
jq -e '.server_host=="unknown.example.com"' audit.jsonl | head -n1 >/dev/null
# Falco出力（将来E2E）
# assert_json_field falco_event.json '.rule' 'MCP Unapproved Endpoint Access'
```

# 注意/落とし穴

- DNS/ホスト名の取り扱い: ローカル検証では `unknown.example.com` の名前解決が難しい場合がある。
  - 代替: プロキシの `--target` でホスト名をそのまま使用し、実際は `127.0.0.1` に向ける。
- レスポンス肥大化: ランナーでは帯域/メモリ制約あり。サイズ/遅延は小さく段階的に。
- 証明書: 自己署名を使う場合、クライアント側の検証を無効化できる設定が必要。

# 受入基準

- 各シナリオの監査JSONに期待フィールドが出現する。
- Falco連携（将来E2E）で対応ルールが発火する。
- 手順が決定論的に再現できる（設定/サイズ/遅延/回数が明記）。

