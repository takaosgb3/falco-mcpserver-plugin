# 目的

このリポジトリのCI（GitHub Actions）設計を明文化し、誰が見ても「何を・いつ・どの基準で」実行しているかを理解できるようにする。初期は軽量・堅牢（SKIP設計）で進め、段階的に品質ゲートを強化する。

# 前提

- ランナー: GitHub Actions（`ubuntu-latest`）。
- 方針: 小さく可逆／決定論／SKIP方針／最小権限／ノイズ低減（PR/手動中心）。
- スコープ: 監査Producer（wrap/proxy）のビルド、オフラインE2E（Falco非依存）、将来のlint/test/可視化。

# 結論（要旨）

- トリガーは `pull_request` と `workflow_dispatch` を基本とし、Pushの自動実行は当面無効化。
- ジョブは Build（Go 1.21）＋ オフラインE2E（SKIP対応）で最小構成。将来 `lint/test` を段階導入。
- 可視化は Step Summary/アーティファクト収集で強化予定（現状はログ中心）。

# ワークフロートポロジー

- ワークフロー: `.github/workflows/ci.yml`
- トリガー:
  - `pull_request`: mainへのPRで実行（規定）。
  - `workflow_dispatch`: 手動実行（任意検証）。
- ジョブ: `build-and-check`
  - Setup: `actions/checkout@v4`, `actions/setup-go@v5 (1.21.x)`
  - Build: `go build ./cmd/mcp-audit-wrap`, `go build ./cmd/mcp-audit-proxy`
  - Deps: `apt-get install -y jq bc file`
  - Offline E2E: `bash test/e2e/basic/test_mcp_audit_basics.sh || true`, `bash test/e2e/attack/test_mcp_anomalies.sh || true`

# ステップ詳細（現行）

- Build: 監査Producerのコンパイル確認（文法/型/依存の破綻を早期検知）。
- Offline E2E:
  - 目的: サンプルJSON/スクリプト/ルールの整合性と導線を検証（Falco非依存）。
  - SKIP設計: 依存不足や環境未整備時は `exit 0` によりスキップ、CIは落とさない（`|| true`）。

# 成果物/可視化（設計）

- 近々: Step Summary に `[PASS]/[SKIP]` 件数、主要ファイル確認状況を出力。
- 近々: `test-results/`（JSON/ログ）をアーティファクト化（ダウンロード可）。
- 現状: Actionsログで確認（`[INFO]`, `[PASS]`, `[SKIP]` が目印）。

# 失敗基準と段階的ゲート

- 現段階（MVP準備）: Build 失敗はジョブ失敗。Offline E2E は情報提供（非ゲート、`|| true`）。
- フェーズ移行後: Offline E2E をゲート化（例: ルール/JSONアサーション不一致で失敗）。
- 将来: 代表ユニットテスト/`go vet`/`fmt -s -d` を必須ゲートに昇格。

# セキュリティ/権限

- 最小権限: デフォルトトークン（`GITHUB_TOKEN`）の権限は最小限（contents: read）を推奨。
- 秘密情報: 現状は未使用。将来利用時は `secrets` に限定し、不要な出力・ログを避ける。

# パフォーマンス/キャッシュ

- Go モジュールキャッシュ: `actions/setup-go` のビルトインキャッシュを活用。
- 将来: `actions/cache` で `~/.cache/go-build` 等を明示キャッシュ検討。

# 並列/同時実行

- 同一PRの不要な重複実行を避けるため、必要に応じ `concurrency` でキャンセル（将来導入）。

# ログ/命名/可読性

- ステップ名は動詞＋対象（例: `Build audit tools`, `Run E2E (offline)`）。
- スクリプトは `[INFO]/[PASS]/[SKIP]` の記号を含む整形出力で判読性を確保。

# 将来拡張（ロードマップ）

- Lint/Test: `go fmt -s -d`, `go vet`, `go test ./...` を追加。
- 可視化: Step Summary生成、アーティファクト収集、自動コメント（PR）で要約提示。
- マトリクス: GoバージョンやOSの追加（必要時）。
- 本格E2E: Falco実行可能な環境向けの別ジョブ/別ワークフロー（SKIP設計は維持）。
- 保護: 必須チェックとして `build-and-check` をブランチ保護に設定。

# 受入基準

- PR/手動実行で安定完走（Build成功、E2EはSKIP/またはPASS）。
- 主要ステップのログが読みやすく、失敗時の原因が特定しやすい。
- 将来のlint/test/可視化の追加に耐える構成（分離・順序性）。

# リスク/代替

- リスク: Offline E2E が緩すぎる → 代替: 一部アサーションをゲート化／Step Summaryで注意喚起。
- リスク: ランナー差異 → 代替: 依存明示・固定化、キャッシュ活用、必要時に容器化。
- リスク: ログ肥大 → 代替: 重要要約は Step Summary、詳細はアーティファクトへ。

# 次アクション

- Step Summary とアーティファクト収集の実装。
- Lint/ユニットテストの追加と、段階的ゲート化ポリシーの定義。
- `concurrency` 導入による重複実行の抑制（同一PR）。

