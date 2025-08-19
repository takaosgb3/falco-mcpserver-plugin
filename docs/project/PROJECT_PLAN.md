# 目的

MCP Server 向け Falco プラグイン（以降「MCP プラグイン」）の開発を段階的に推進し、短期で価値を提供しつつ、将来拡張に耐える土台（スキーマ/ルール/テスト/CI）を整備する。

# 前提

- ランタイム: GitHub Actions ランナー（ubuntu-latest）/ローカル環境（Falcoはドライバレス可）。
- ポリシー: SDK-first、小さく可逆、CI優先、SKIPポリシー、PII最小化。
- 入力: まずは MCP 監査 JSON（sidecar/proxy/クライアントで生成）。将来 eBPF/プロキシ連携に拡張。

# 結論（要旨）

- フェーズ1で「監査JSON→mcp.*フィールド→ルール→JSONアサーションE2E」を成立させ、基本検知4本を提供。
- フェーズ2で「カバレッジ拡張（メトリクス/しきい値学習/許可リスト運用）と CI 可視化」を強化。
- フェーズ3で「観測面の拡張（eBPF/プロキシ）と運用自動化（集計/レポート）」に進む。

# フェーズ計画（更新）

- フェーズ1（MVP, 観測最小・検知最小）
  - 監査スキーマ v1 固定（`docs/schema/mcp_audit_v1.json`）。
  - 監査Producer 最小実装（完了）: `mcp-audit-wrap`（stdio）, `mcp-audit-proxy`（tcp/ws透過）。
  - CI 初期構成（完了）: GitHub Actions（PR/手動）で Build + オフラインE2E（SKIP設計）。
  - ルール最小4本（完了）: 未承認エンドポイント/非TLS/大量送信/過剰呼び出し。
  - 方針: eBPF/カーネルモジュールは不使用。監査は StdIO/WS/クライアント内で生成。
- フェーズ2（品質/運用強化・実動テスト導入）
  - テスト用 MCP Server を実装（`docs/testing/MCP_TEST_SERVER_DESIGN.md` 準拠）。
  - 代表シナリオ（S1〜S4）を自動化（`docs/testing/MCP_TEST_SCENARIOS.md`）。
  - 監査Producer拡張: JSON-RPC `.method` 抽出（wrap/proxy）と per-call イベント出力（必要に応じて有効化）。
  - CI 可視化: Step Summary とアーティファクト収集（`test-results/`）。
- フェーズ3（運用外部化/ゲート強化）
  - 許可リスト/しきい値の外部化（YAML）と環境別上書き。
  - 検出率/ゲート条件の導入（E2Eの一部を必須化）。
  - サンプル/エッジケース拡充、壊れたJSON/欠損フィールド耐性強化。
- フェーズ4（観測拡張/自動化）
  - eBPF/システムコール観測 or 軽量プロキシでネットワーク/ファイルI/O補足。
  - 日次レポート/傾向検知（しきい値自動提案）、許可リスト更新ワークフロー。

# マイルストーン / 成果物（更新）

- M1: スキーマ・ルール・Producer初期（完了）
  - `docs/schema/mcp_audit_v1.json`, `rules/mcp_baseline.yaml`, `cmd/mcp-audit-wrap`, `cmd/mcp-audit-proxy`
- M2: CI基盤（初期）（完了）
  - `.github/workflows/ci.yml`（PR/手動）, `docs/ci/CI_DESIGN.md`
- M3: テスト用 MCP Server + シナリオ（新規）
  - `mcp-test-server`（stdio/ws）, `docs/testing/*` 反映, 自動化スクリプト
- M4: CI 可視化/安定化（新規）
  - Step Summary, `actions/upload-artifact` による `test-results/` 収集
- M5: 運用外部化/ゲート強化（新規）
  - 設定YAML, 検出率/ゲート, ユニット/`go vet`/`fmt -s -d` の導入
- M6: 観測拡張 PoC（将来）

# 受入基準

- ユニット: JSON→フィールド変換のテーブル駆動テストが安定通過。
- E2E: 代表シナリオで `.rule/.priority/output_fields["mcp.*"]` が一致。
- CI: GitHub Actions ランナーで安定、`jq bc file` 導入、SKIP時は明確な理由を出力。
- 互換: `mcp.*` フィールドは後方互換、追加はOK/リネームは原則禁止（同時にルール/テスト更新）。

# リスク/代替

- 監査フィード未整備 → 代替: クライアントラッパ/ローカルプロキシで JSON を生成。
- 過検知/誤検知 → 代替: 許可リスト/学習期間/検出率しきい値/緩和ルールを導入。
- プライバシー懸念 → 代替: レッドアクト/ハッシュ化/オプトアウト設定と透明性の文書化。

# タイムライン（例示・更新）

- 週1–2: M1/M2 完了（達成済）
- 週3–4: M3（テストサーバー/シナリオ実装）、M4（CI可視化）
- 週5+: M5（運用外部化/ゲート強化）、M6 PoC と次期計画レビュー

# 体制/役割（R&R）

- オーナー: 要件/優先度の決定、受入判断。
- 実装: プラグイン/ルール/テスト/スクリプトの実装と文書更新。
- レビュー: 互換性・セキュリティ・性能の観点レビュー、CI 監視。

# 運用/CI

- ランナー: `runs-on: ubuntu-latest` を前提。
- 依存: `jq bc file` を明示インストール。アーティファクト収集を標準化。
- ガード: `grep -r "ubuntu-latest" .github/workflows` などで誤設定を検知。

# 次アクション（更新）

- テスト用 MCP Server の最小実装（stdio/ws）と自動化スクリプトの追加。
- 監査Producer: JSON-RPC `.method` 抽出と per-call 出力のオプション化（wrap/proxy）。
- CI 可視化: Step Summary とアーティファクト収集を追加、E2E一部をゲート化する方針を試験。
- 運用外部化: 設定YAMLの導入（許可リスト/しきい値）とドキュメント整備。
