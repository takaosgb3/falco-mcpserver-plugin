# 目的

実運用前に「監査Producer（stdio/WS/client-emitter）」「Falcoプラグイン（mcp.*）」の実動作を検証するためのテスト用 MCP Server の設計と仕様を定義する。ランナー制約下（GitHub Actions）でも再現しやすく、ローカルでも容易に起動できることを目指す。

# 前提

- プロトコル: JSON-RPC 2.0 準拠（疑似）。
- 輸送層: StdIO（ローカル）/ WebSocket（`ws://`/`wss://`）をサポート。
- 目的: メタデータ（メソッド/バイト/頻度/TLS可否/接続先）を安定に生成し、監査Producer→Falcoまでの導線確認。
- セキュリティ: テスト専用。デフォルトは `127.0.0.1` で待受。`wss`は自己署名証明書を許容。

# 機能要件（MVP）

- 複数モード:
  - StdIO サーバー: 標準入出力で JSON-RPC を処理（`--mode stdio`）。
  - WebSocket サーバー: `--listen :8080` 等で待受（`--mode ws`）。
  - TLS（任意）: `--tls-cert`/`--tls-key` 指定で `wss://` 提供。
- 実装メソッド（最小）:
  - `tools.list`: ダミーツール一覧を返す（小応答）。
  - `prompts.get`: 指定IDのプロンプトメタを返す（小～中応答）。
  - `tools.exec`: 入力サイズ/出力サイズ/遅延を調整可能（大応答テスト用）。
- 制御パラメータ（フラグ/環境変数/YAML）:
  - 応答サイズ（bytes）: `--resp-size-bytes=N`（`tools.exec`デフォルト値）。
  - 遅延（ms）: `--delay-ms=N`（全メソッド共通の人工レイテンシ）。
  - エラー率: `--error-rate=0.0..1.0`（一定確率でJSON-RPCエラーを返す）。
  - 実行回数増幅: `--burst-calls=N`（1回の呼び出しでN回分の内部カウントを発火＝頻度異常再現）。
  - ファイルアクセス数: `--file-access-inc=N`（監査上のfile_access_count相当の擬似カウント）。
- 観測/ログ:
  - サーバー自身のアクセスログは最小限（INFOレベル）。
  - 監査はProducer側（wrap/proxy/client-emitter）で出力（本サーバーは原則出力しない）。

# 非機能要件

- 単体バイナリで軽量起動（Go製、依存最小）。
- 負荷の上限を安全に制御（レスポンス上限、遅延上限）。
- 決定論的モードを提供（`--seed`で疑似乱数固定）。

# インタフェース（CLI）

- 起動例（StdIO）:
  - `mcp-test-server --mode stdio --delay-ms 10 --resp-size-bytes 0`
- 起動例（WS）:
  - `mcp-test-server --mode ws --listen :8080 --delay-ms 5`
- 起動例（WSS）:
  - `mcp-test-server --mode ws --listen :8443 --tls-cert cert.pem --tls-key key.pem --delay-ms 1`

# JSON-RPC（例）

- リクエスト（tools.exec）:
```json
{"jsonrpc":"2.0","id":1,"method":"tools.exec","params":{"cmd":"echo","payload":"..."}}
```
- レスポンス（正常）:
```json
{"jsonrpc":"2.0","id":1,"result":{"ok":true,"size":10485760}}
```
- レスポンス（エラー）:
```json
{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"simulated error"}}
```

# フロー（監査導線）

- StdIO 経路:
```
Editor/CLI ──(stdio JSON-RPC)──▶ mcp-audit-wrap ──▶ mcp-test-server (stdio)
Editor/CLI ◀─(stdio JSON-RPC)──  mcp-audit-wrap ◀── mcp-test-server (stdio)
                              └─ Emit JSONL (bytes/method/timestamp) → sink
```
- WS 経路（wss対応）:
```
Editor/CLI ──ws://127.0.0.1:8989──▶ mcp-audit-proxy ──wss://127.0.0.1:8443──▶ mcp-test-server (ws)
Editor/CLI ◀──────────────────────────────────────────────────────────────────────── mcp-test-server (ws)
                                   └─ Emit JSONL (frames/host/tls) → sink
```

# 監査マッピング（期待）

- `server_host`/`server_port`/`tls`: プロキシ/接続情報から自明。
- `method`: JSON-RPCの`method`（wrap/プロキシのパーサ追加で抽出）。
- `request_bytes`/`response_bytes`: 送受信した合計バイト。
- `tool_invoke_count`/`file_access_count`: サーバー側の制御パラメータまたはクライアントエミッタで加算。
- `client_process`: 実行中のクライアント識別（`--client-process` など）

# テストシナリオ（MVP）

- 未承認エンドポイント: `server_host` を許可外にしてアクセス（ルール発火を期待）。
- 非TLS通信: `ws://` で接続（`mcp.tls=false`）。
- 大量送信: `--resp-size-bytes 10485760` 以上で応答を肥大化。
- 過剰呼び出し: `--burst-calls 200` などで連続呼出を模擬。

# 導入/統合

- ローカル: `mcp-audit-wrap`/`mcp-audit-proxy` と組合せることで、監査JSONLを生成→Falcoプラグインへ。
- CI: GitHub Actionsでは負荷を抑えた短時間プロファイル（小応答/短遅延/エラー率0）で実行。現状は設計ドキュメントのみ、将来は別ワークフロー/ジョブに分離し、SKIP可能に。

# 将来拡張

- メソッドの拡充（`files.read`/`files.write` 等）。
- バイナリペイロードの疑似分割送信（フレーム化テスト）。
- 応答パターンのスクリプト化（YAMLシナリオ駆動）。
- 低負荷ベンチマークモード（bps指定・遅延分布）。

# 受入基準

- 2モード（stdio/ws）で起動可能。
- 指定サイズ/遅延/エラー率が応答に反映され、Producer の監査JSONに期待フィールドが現れる。
- テストシナリオに応じて、ルール（未承認/非TLS/大量/過剰）が発火する。（Falco側のE2E整備後）

---

## 付録: WS モード設計（最小仕様・Phase 2 追記）

- 依存ポリシー: 外部ライブラリを追加せず、標準ライブラリで RFC6455 の最小要件を満たす実装を目指す（検討中）。
- フレーム形式: テキストフレームのみ対応。クライアント→サーバは必ずマスク、サーバ→クライアントは非マスク。
- 分割/継続フレーム: Phase 2 では未対応（単一テキストメッセージのみ）。
- サイズ上限: 受信テキストは 1–5MiB 程度に上限を設け、超過時はクローズ（DoS抑止）。
- 同時接続: デフォルト最大 8 接続程度に制限（資源保護）。
- JSON-RPC 取り扱い: 1リクエスト=1レスポンス、`id` は string/number/null 許容、バッチは未対応。
- 制御フラグ: `--delay-ms`, `--resp-size-bytes`, `--error-rate`, `--burst-calls`, `--seed` を stdio と同等に反映。
- WSS 対応: Phase 2 では任意（将来対応）。CI 複雑化回避のため既定は `ws://` のみ。
- ログ/運用: 起動時にバインド先/上限を INFO で明示。異常入力はエラー応答または即時クローズ（フェイルセーフ）。
