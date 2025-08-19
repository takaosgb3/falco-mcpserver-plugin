# 目的

フェーズ2（品質/運用強化・実動テスト導入）のタスクを詳細化し、受入基準・検証手順・完了条件（DoD）を明確にする。フェーズ1の最小構成を土台に、実サーバ検証とCI可視化で信頼性を高める。

# スコープ

- 実動テスト: テスト用 MCP Server（stdio/ws）と代表シナリオ（S1〜S4）。
- Producer強化: `.method` 抽出（wrap/proxy）と per-call 監査イベント（任意有効）。
- CI可視化: Step Summary と `test-results/` アーティファクト収集。
- 運用外部化: 設定YAML（許可リスト/しきい値）と軽量プリプロセス。
- 併せて開始: ユニット/静的解析（最小セット）。
- 除外: eBPF/カーネル（Phase4）。

# タスク詳細（Workstreams／タスク番号付与）

- [P2-1] テスト用 MCP Server 実装
  - 目的: 監査導線の実動作を再現（stdio/ws/wss）。
  - 作業:
    - [ ] `cmd/mcp-test-server` 追加（Go）: `--mode stdio|ws`, `--listen`, `--tls-cert/key`
    - [ ] メソッド: `tools.list`, `prompts.get`, `tools.exec`
    - [ ] 制御: `--resp-size-bytes`, `--delay-ms`, `--error-rate`, `--burst-calls`, `--seed`
    - [ ] ドキュメント: `docs/testing/MCP_TEST_SERVER_DESIGN.md` 仕様との差分反映
  - 受入: stdio/ws起動でき、制御がレスポンス/遅延/エラーに反映。

- [P2-2] シナリオ自動化（S1〜S4）
  - 目的: 代表的な検知が再現・検証できる手順を自動化。
  - 作業:
    - [ ] `test/integration/` にスクリプト追加（SKIP設計）
    - [ ] 期待値: 監査JSONの必須フィールド/JQ断片・将来Falcoイベント断片
    - [ ] 最小負荷プロファイル（CI用）と拡張プロファイル（ローカル用）のテンプレ化
  - 受入: ローカル/CI（手動トリガ）で決定論的に再現可能。

- [P2-3] 監査Producer強化（wrap/proxy）
  - 目的: 粒度の細かい観測（メソッド/呼出単位）。
  - 作業:
    - [ ] `mcp-audit-wrap`: JSON-RPC `.method` の抽出、per-call で `audit.Emit`（フラグで切替）
    - [ ] `mcp-audit-proxy`: 可能な範囲で `.method` 抽出（TLSは透過維持）
    - [ ] セッション/相関IDの付与（任意）
  - 受入: サンプル呼出で `.method` と per-call イベントが出力（有効時）。

- [P2-4] CI 可視化/成果物
  - 目的: CI上で結果が一目で把握でき、再現性が高まる。
  - 作業:
    - [ ] Step Summary 出力（PASS/FAIL/SKIP件数、主要ファイルの有無）
    - [ ] `actions/upload-artifact` で `test-results/` を収集
    - [ ] （任意）同一PRの重複実行キャンセル `concurrency`
  - 受入: PRページからサマリ/成果物が確認可。

- [P2-5] 運用外部化/ルール生成
  - 目的: 設定変更を安全・簡便に。
  - 作業:
    - [ ] `docs/config/EXAMPLE_VALUES.yaml` をベースに許可リスト/しきい値のYAML導入
    - [ ] 軽量プリプロセスで `rules/mcp_baseline.yaml` を生成（置換レベル）
    - [ ] ドキュメント更新（変更手順/注意点）
  - 受入: 設定変更がルールに反映、E2Eで整合。

- [P2-6] 最小ユニット/静的解析
  - 目的: 基盤品質の底上げ。
  - 作業:
    - [ ] `go fmt -s -d`, `go vet` をCIに追加
    - [ ] 監査イベント→`mcp.*` 変換のテーブル駆動ユニットを開始（最小）
  - 受入: CIでlint/testが実行され、失敗時に明確な原因が出る。

# 受入基準（Phase 2）

- テストサーバ: stdio/ws起動・メソッド/制御が動作。
- シナリオ: S1〜S4の監査JSONが期待通り（JQ断片一致）。
- Producer: `.method` 抽出と per-call 出力が有効化時に機能。
- CI: Step Summary/成果物が表示・取得できる。
- 設定外部化: YAML変更がルールに反映・検証で一致。
- Lint/Test: CIで `fmt`/`vet`/最小ユニットが走る。

# 検証手順（抜粋）

- ローカル:
  - `mcp-test-server --mode stdio ...` を `mcp-audit-wrap -- ...` で挟む
  - `mcp-test-server --mode ws --listen :8080 ...` と `mcp-audit-proxy --listen :8989 --target ws://127.0.0.1:8080 ...`
  - `jq` で監査JSONの `.method`, `bytes`, `host`, `tls` を確認
- CI（手動）:
  - `workflow_dispatch` で実行、Step Summary/Artifacts を確認

# リスク/注意

- `.method` 抽出の負荷: パースは軽量に、失敗時はスキップ・メトリクス化。
- wss検証: 自己署名証明書の扱い・検証無効化が必要な場合は明示化。
- ランナー資源: レスポンスサイズ/遅延は小さく、時間内に収束させる。

# 完了の定義（DoD）

- [ ] テスト用MCPサーバが実装され、基本メソッド/制御が動作
- [ ] S1〜S4自動化が動作し、監査JSONの期待値が一致
- [ ] Producerの `.method`/per-call（有効時）が動作
- [ ] CIにStep Summary/Artifactsが追加され可視化
- [ ] 設定YAML→ルール生成→検証の一連が成立
- [ ] 最小lint/testがCIに組み込まれ安定

