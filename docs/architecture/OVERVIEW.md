# 目的

MCP Server 向け Falco プラグインの全体像を共有し、設計判断・実装・テスト・運用の一貫性を担保する。高レベルの構成要素とデータフローを明確化する。

# 前提

- 入力: MCP 監査イベント（JSON 行）を標準スキーマで受け取る。
- 方針: SDK-first、小さく可逆、後方互換、PII最小化、SKIP方針、自己ホストRunner。
- フィールド: `mcp.*` 名前空間。Falco ルールは JSON アサーションで検証。

# 結論（要旨）

- 「監査イベント →（プラグインで変換）→ Falco イベント（mcp.*）→ ルール検知 → JSON 出力/可視化」という直列パイプライン。
- コアはソース型プラグインで、スキーマ準拠の JSON を検証・マッピングし、StringPool で文字列を安全に扱う。
- 検知は未承認エンドポイント/非TLS/大量送信/過剰呼び出しのベースから開始し、外部化設定で運用を容易化。

# 構成要素

- 監査イベント供給源（外部）: sidecar/proxy/クライアントラッパが JSON 生成（`docs/schema/mcp_audit_v1.json` 準拠）。
- プラグイン・コア（SDK）: 入力を読み出し、検証→フィールド抽出→Falco イベント化。
- パーサ/バリデータ: スキーマ準拠チェック・型安全な取り出し・異常値スキップ。
- フィールド抽出/メモリ管理: `mcp.*` へマッピング、StringPool による一括管理（バッチ毎クリア）。
- 設定ローダ: 許可リスト/しきい値/レッドアクト方針を YAML で外部化。
- ルールセット: ベースライン4本 + テンプレ（`rules/templates/`）。
- 可観測性: JSON 出力、メトリクス（処理/スキップ/警告数）、SKIP ログ。

# データフロー（論理）

1) 監査 JSON 生成 → 2) プラグイン読取（行） → 3) スキーマ検証/正規化 → 4) `mcp.*` へ変換（StringPool） → 5) Falco イベント発行（`evt.source=mcp_audit`） → 6) ルール評価 → 7) JSON 出力/集約（`test-results/`）。

# 図（ASCII）

```
┌──────────────────────┐    JSON (line-by-line)    ┌──────────────────────┐
│  Audit Producer(s)   │  ───────────────────────▶  │  Plugin (SDK-first)  │
│  - sidecar/proxy     │                            │  - parse/validate    │
│  - client wrapper    │                            │  - map → mcp.*       │
└─────────┬────────────┘                            │  - StringPool batch  │
          │                                         └─────────┬────────────┘
          │ Falco event (evt.source=mcp_audit)                │
          ▼                                                   ▼
      ┌───────────────────────────────────────────────────────────────┐
      │                          Falco Engine                         │
      │  rules:                                                       │
      │   - unapproved endpoint (mcp.server_host)                     │
      │   - insecure TLS (mcp.tls=false)                              │
      │   - excessive data (mcp.request_bytes/response_bytes)         │
      │   - excessive calls (mcp.tool_invoke_count)                   │
      └─────────┬─────────────────────────────────────────────────────┘
                │ JSON alerts (with output_fields["mcp.*"]) 
                ▼
          ┌──────────────┐        store/analyze        ┌──────────────────┐
          │ Sink(s)      │  ───────────────────────▶   │ test-results/     │
          │ - stdout     │                             │ CI summary/jq     │
          │ - file/syslog│                             │ assertions        │
          └──────────────┘                             └──────────────────┘
```

# バウンダリ/責任

- プラグイン: 入力検証・変換・堅牢性・メモリ管理。内容（PII/ペイロード）は最小限。
- ルール/運用: 閾値・許可リスト・検出率方針の調整と可視化。
- CI/E2E: JSON アサーション、SKIPポリシー、アーティファクト収集。

# 受入基準

- 代表シナリオで `output_fields["mcp.*"]` が期待通りに出力され、ベースルールが発火。
- 破損 JSON/欠損フィールドで安全にスキップし、プラグインが落ちない。

# リスク/代替

- 監査フィード未整備 → 軽量プロキシ/クライアントラッパで代替。
- 過検知 → 許可リスト/検出率しきい値/学習期間の導入。

# 次アクション

- SDK 雛形と抽出器を実装、スキーマ→フィールドのコード生成を活用。
- ベースルールを本線化し、E2E を JSON アサーションで接続。
