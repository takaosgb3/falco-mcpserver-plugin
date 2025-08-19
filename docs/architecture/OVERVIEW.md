# 目的

MCP Server 向け Falco プラグインのアーキテクチャを、読みやすく・わかりやすく・理解しやすく・MECE（漏れなく/ダブりなく）に示す。設計の意図、構成、データ、フロー、責任、非機能、導入方法をひと目で把握できるようにする。

# 前提

- 入力: 監査イベント（JSON行、`docs/schema/mcp_audit_v1.json` 準拠）
- 方針: SDK-first／小さく可逆／後方互換／PII最小化／SKIP方針／自己ホストRunner
- 除外: eBPF/カーネルモジュールは Phase 1 では不使用（将来の拡張候補）
- フィールド: `mcp.*` 名前空間。Falco ルールは JSON アサーションで検証

# 結論（要旨）

- 直列パイプライン: 「監査JSON → プラグイン（検証/変換） → Falco（`mcp.*`） → ルール検知 → JSON出力/可視化」
- 監査Producerはアプリ/トランスポート層に挟み込む（StdIOラッパ／WSローカルプロキシ／クライアント内エミッタ）
- 運用は外部化設定（許可リスト/しきい値/レッドアクト）＋ SKIP で安定化。CI は JSON アサーション中心

# 構成（役割と責任）

- 監査Producer（外部）: Claude Code/Codex CLI の MCP I/O を観測し、メタデータのみ抽出して JSON 行出力
- プラグイン・コア（SDK）: JSON を読み取り、スキーマ検証→`mcp.*` 抽出→ Falco イベント化（StringPool 遵守）
- ルールセット: ベースライン検知（未承認エンドポイント/非TLS/大量送信/過剰呼び出し）。環境値は YAML で外部化
- 可観測性/CI: JSON 出力、`jq` アサーション、メトリクス（処理/スキップ/警告数）、アーティファクト収集

# データモデル（`mcp.*` 主要フィールド）

- セッション: `mcp.session_id`, `mcp.client_process`, `mcp.timestamp`
- 接続: `mcp.server_host`, `mcp.server_port`, `mcp.tls`, `mcp.auth_scheme`
- 呼出: `mcp.method`, `mcp.tool_invoke_count`, `mcp.file_access_count`
- サイズ: `mcp.request_bytes`, `mcp.response_bytes`
- 拡張: `mcp.dataset_id`, `mcp.scope`, `mcp.error_code`, `mcp.cert_fingerprint`

# データフロー（段階）

1) 監査Producerがメタデータを JSON 行で出力
2) プラグインが行を読取り、スキーマ検証（破損/欠損はスキップ）
3) 正規化して `mcp.*` にマッピング（StringPool で文字列を一括管理）
4) Falco へイベント発行（`evt.source=mcp_audit`）
5) ルール評価（許可リスト・しきい値は YAML 外部化の値を参照）
6) JSON アラート出力／`test-results/` へ集約（E2E/CI 用）

# 監査Producer（eBPF不使用）

- StdIO ラッパ（ローカル）: `mcp-audit-wrap -- <real-mcp>` が stdin/stdout を中継し、行フレームやサイズを計測
- WS ローカルプロキシ（リモート）: `mcp-audit-proxy --listen :8989 --target wss://...` がフレーム長や方向を記録（TLS透過）
- クライアント内エミッタ（Codex CLI）: 呼出後に JSON 行を直接出力（最小・確実）

# ルール設計（概要）

- 未承認エンドポイント: `mcp.server_host` が許可外
- 非TLS: `mcp.tls=false`
- 大量送信: `mcp.request_bytes` or `mcp.response_bytes` > しきい値
- 過剰呼出: `mcp.tool_invoke_count` > しきい値

# 設定/運用（外部化）

- 許可リスト/しきい値/レッドアクト: `docs/config/EXAMPLE_VALUES.yaml` を基に環境で上書き
- 変更は小さく可逆に（ルール変更時はテスト更新）。CIは自己ホストRunner＋SKIP方針

# セキュリティ/プライバシー原則

- デフォルトで本文を収集しない（メタデータ中心）
- 認証/識別子はハッシュ/マスク可能（将来の `redaction` 設定で制御）
- 監査ファイル権限は 0600、ローテーション/保持期間を明示

# 信頼境界と責任分担

- Producer: メタデータ抽出の正確性・過収集の抑制
- プラグイン: 入力堅牢性・型安全・メモリ管理（StringPool）
- ルール/運用: 閾値/許可リストの調整、誤検知/過検知のバランス
- CI/E2E: JSON アサーション、アーティファクト可視化

# 可観測性/メトリクス（例）

- `events_processed_total`, `events_skipped_total`, `alerts_emitted_total`
- `avg_request_bytes`, `avg_response_bytes`, `host_violation_rate`

# エラー処理/フォールトトレランス

- 壊れた JSON/欠損はスキップし、件数をメトリクス化。プラグインは継続
- 設定不備はデフォルト緩和＋明示ログ（fail-open か fail-safe は環境で選択）

# パフォーマンス/容量（目安）

- 目標: 1万イベント/分、単純マッピングの低オーバーヘッド
- 監査JSON: 1行あたり 0.3–1KB 程度（メタデータのみ）。日次ローテーション推奨

# 非機能要件

- 互換性: フィールドの追加は自由、削除/リネームは原則禁止（後方互換を維持）
- 拡張性: 将来の eBPF/プロキシ強化/自動レポートに拡張可能

# 導入パターン（例）

- Claude Code（ローカルMCP）: VS Code 設定の `command` を `mcp-audit-wrap` に変更
- Claude Code（リモートMCP）: `url` を `ws://127.0.0.1:8989` にして `mcp-audit-proxy` を起動
- Codex CLI: クライアント内エミッタ（`MCP_AUDIT_SINK` で出力先指定）

# 図（ASCII：全体像）

```
┌──────────────────────┐    JSON (line-by-line)    ┌──────────────────────┐
│  Audit Producer(s)   │  ───────────────────────▶  │  Plugin (SDK-first)  │
│  - stdio wrapper     │                            │  - parse/validate    │
│  - ws local proxy    │                            │  - map → mcp.*       │
│  - client emitter    │                            │  - StringPool batch  │
└─────────┬────────────┘                            └─────────┬────────────┘
          │ Falco event (evt.source=mcp_audit)                 │
          ▼                                                    ▼
      ┌───────────────────────────────────────────────────────────────┐
      │                          Falco Engine                         │
      │  rules: unapproved host / insecure TLS / data / calls         │
      └─────────┬─────────────────────────────────────────────────────┘
                │ JSON alerts (with output_fields["mcp.*"]) 
                ▼
          ┌──────────────┐        store/analyze        ┌──────────────────┐
          │ Sinks        │  ───────────────────────▶   │ test-results/     │
          │ stdout/file  │                             │ CI summary/jq     │
          └──────────────┘                             └──────────────────┘
```

# 受入基準

- 代表シナリオで `output_fields["mcp.*"]` が期待通り、ベースルールが発火
- 破損 JSON/欠損で安全にスキップ（クラッシュしない）、メトリクス記録
- CI（自己ホストRunner）で JSON アサーションが安定

# リスク/代替

- 監査フィード未整備 → クライアント内エミッタ or 軽量プロキシで代替
- 過検知 → 許可リスト/しきい値調整、検出率しきい値/学習期間の導入
- プライバシー懸念 → レッドアクト/ハッシュ化/保持期間短縮

# 次アクション

- プラグイン雛形と変換テーブル実装（コード生成の適用）
- Producer（wrap/proxy）の機能拡張（メソッド名抽出、セッション相関）
- ルール本線化と E2E の JSON アサーション強化

