# 🚀 CI/CDガイド（統合版）

> 最終更新: 2025-08-03
> 統合元: 複数のCI/CD関連ドキュメント

## 📋 目次

1. [概要](#概要)
2. [クイックスタート](#クイックスタート)
3. [ワークフロー構成](#ワークフロー構成)
4. [セルフホストランナー設定](#セルフホストランナー設定)
5. [よくある問題と解決策](#よくある問題と解決策)
6. [コスト最適化](#コスト最適化)
7. [トラブルシューティング](#トラブルシューティング)
8. [ベストプラクティス](#ベストプラクティス)

## 概要

このドキュメントは、Falco nginx pluginプロジェクトのCI/CDに関する包括的なガイドです。

### 統合されたドキュメント
- CI_CD_GUIDE.md
- CI_CD_QUICKSTART_TEMPLATE.md
- CI_CD_TROUBLESHOOTING_GUIDE.md
- CI_CD_PITFALLS_AND_SOLUTIONS.md
- CI_CD_ERROR_PREVENTION_GUIDE.md
- GITHUB_ACTIONS_OPTIMIZATION.md
- GITHUB_ACTIONS_COST_REDUCTION_PLAN.md
- CI_INFRASTRUCTURE_GUIDE.md

## クイックスタート

### 1. 基本的なワークフロー作成

```yaml
name: Test Workflow
on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    # 🔴 重要: 必ずセルフホストランナーを使用
    runs-on: [self-hosted, linux, x64, local]

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run tests
        run: make test
```

### 2. ローカルでのテスト

```bash
# actを使用してローカルでワークフローをテスト
act -j test

# 特定のイベントをシミュレート
act pull_request

# 秘密情報を含むテスト
act -s GITHUB_TOKEN=$GITHUB_TOKEN
```

## ワークフロー構成

### アクティブなワークフロー

| ワークフロー | 目的 | トリガー | 実行時間 |
|------------|------|---------|----------|
| test.yml | 単体テスト実行 | push, PR | ~5分 |
| build.yml | バイナリビルド | push, PR | ~3分 |
| release.yml | リリース作成 | タグプッシュ | ~10分 |
| integration-test.yml | 統合テスト | PR, 日次 | ~15分 |
| security-scan.yml | セキュリティスキャン | PR, 週次 | ~8分 |

### ジョブ間の依存関係

```yaml
jobs:
  test:
    runs-on: [self-hosted, linux, x64, local]
    # テストを最初に実行

  build:
    needs: test  # testが成功後に実行
    runs-on: [self-hosted, linux, x64, local]

  integration:
    needs: [test, build]  # 両方成功後に実行
    runs-on: [self-hosted, linux, x64, local]
```

## セルフホストランナー設定

### 💰 コスト削減の鉄則

**絶対にubuntu-latestを使用しない**

```yaml
# ❌ 絶対に使わない（料金発生）
runs-on: ubuntu-latest

# ✅ 必ず使用（料金なし）
runs-on: [self-hosted, linux, x64, local]
```

### セルフホストランナーのセットアップ

```bash
# 1. ランナーのダウンロード
curl -o actions-runner-linux-x64-2.316.1.tar.gz -L \
  https://github.com/actions/runner/releases/download/v2.316.1/actions-runner-linux-x64-2.316.1.tar.gz

# 2. 展開
tar xzf ./actions-runner-linux-x64-2.316.1.tar.gz

# 3. 設定
./config.sh --url https://github.com/takaosgb3/falco-nginx-plugin-claude \
  --token YOUR_TOKEN

# 4. サービスとして起動
sudo ./svc.sh install
sudo ./svc.sh start
```

### ランナーの管理

```bash
# ステータス確認
sudo ./svc.sh status

# ログ確認
journalctl -u actions.runner.takaosgb3-falco-nginx-plugin-claude.runner-1.service -f

# 再起動
sudo ./svc.sh stop
sudo ./svc.sh start
```

## よくある問題と解決策

### 1. Docker Buildx権限エラー

**問題**: `Set up Docker Buildx` が exit code 128 で失敗

**解決策**:
```bash
# Dockerグループに追加
sudo usermod -aG docker $USER

# 権限確認
docker ps

# ワークフローでの回避策
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3
  continue-on-error: true  # エラーを無視
```

### 2. Goビルドキャッシュ競合

**問題**: `File exists` エラーでキャッシュ復元失敗

**解決策**:
```yaml
- name: Clean cache locks
  run: |
    find ~/go/pkg/mod -name "*.lock" -delete || true

- name: Restore cache
  uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
  continue-on-error: true
```

### 3. CGOビルドエラー

**問題**: `use of cgo in test not supported`

**解決策**:
```yaml
- name: Build with CGO
  env:
    CGO_ENABLED: 1
  run: |
    # プラグインビルド（CGO必須）
    make build-plugin

    # テストは内部ロジックのみ
    go test ./pkg/parser ./pkg/watcher
```

### 4. 並行実行の競合

**問題**: 複数のPRで同時実行時の競合

**解決策**:
```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.ref }}
  cancel-in-progress: true
```

## コスト最適化

### 使用状況の監視

```bash
# GitHub CLIで使用状況確認
gh api /repos/takaosgb3/falco-nginx-plugin-claude/actions/billing/usage

# 月次レポート生成
./scripts/generate-github-usage-report.sh
```

### 最適化戦略

1. **キャッシュの活用**
```yaml
- uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

2. **ジョブの並列化**
```yaml
strategy:
  matrix:
    go-version: ['1.21', '1.22']
    os: [self-hosted]  # セルフホストのみ
```

3. **条件付き実行**
```yaml
- name: Run expensive tests
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  run: make integration-test
```

## トラブルシューティング

### デバッグ手法

1. **詳細ログの有効化**
```yaml
- name: Enable debug logging
  run: |
    echo "ACTIONS_STEP_DEBUG=true" >> $GITHUB_ENV
    echo "ACTIONS_RUNNER_DEBUG=true" >> $GITHUB_ENV
```

2. **SSH デバッグセッション**
```yaml
- name: Setup tmate session
  if: ${{ failure() }}
  uses: mxschmitt/action-tmate@v3
  timeout-minutes: 15
```

3. **アーティファクトの保存**
```yaml
- name: Upload logs
  if: always()
  uses: actions/upload-artifact@v4
  with:
    name: debug-logs
    path: |
      **/*.log
      **/test-results.xml
```

### よくあるエラーメッセージ

| エラー | 原因 | 解決策 |
|--------|------|--------|
| `permission denied` | Docker権限不足 | userをdockerグループに追加 |
| `no space left on device` | ディスク容量不足 | `docker system prune -a` |
| `rate limit exceeded` | API制限 | GITHUB_TOKENを設定 |
| `job was cancelled` | タイムアウト | timeout-minutesを増加 |

## ベストプラクティス

### 1. ワークフローの構造化

```yaml
name: CI Pipeline
on:
  workflow_dispatch:  # 手動実行可能
  pull_request:
    types: [opened, synchronize, reopened]
  push:
    branches: [main]

env:
  GO_VERSION: '1.22'

jobs:
  # 軽量なチェックを先に
  lint:
    runs-on: [self-hosted, linux, x64, local]
    steps:
      - uses: actions/checkout@v4
      - name: Run linters
        run: make lint

  # 重いテストは後に
  test:
    needs: lint
    runs-on: [self-hosted, linux, x64, local]
    # ...
```

### 2. エラーハンドリング

```yaml
- name: Critical step
  id: critical
  run: make build

- name: Handle failure
  if: failure() && steps.critical.outcome == 'failure'
  run: |
    echo "Build failed, collecting diagnostics..."
    make diagnose
```

### 3. 秘密情報の管理

```yaml
- name: Use secrets safely
  env:
    API_KEY: ${{ secrets.API_KEY }}
  run: |
    # 秘密情報をログに出力しない
    set +x
    ./deploy.sh
```

### 4. 再利用可能なワークフロー

```yaml
# .github/workflows/reusable-test.yml
on:
  workflow_call:
    inputs:
      go-version:
        required: false
        type: string
        default: '1.22'

# 使用側
jobs:
  test:
    uses: ./.github/workflows/reusable-test.yml
    with:
      go-version: '1.22'
```

## 定期メンテナンス

### 週次タスク
- [ ] ワークフロー実行時間の確認
- [ ] 失敗率の分析
- [ ] キャッシュ効率の確認

### 月次タスク
- [ ] GitHub Actions使用料の確認
- [ ] ランナーのアップデート
- [ ] 不要なワークフローの削除

### 四半期タスク
- [ ] ワークフロー全体の見直し
- [ ] セキュリティアップデート
- [ ] パフォーマンス最適化

---

## 関連リソース

- [GitHub Actions公式ドキュメント](https://docs.github.com/actions)
- [セルフホストランナーガイド](https://docs.github.com/actions/hosting-your-own-runners)
- [act（ローカル実行ツール）](https://github.com/nektos/act)

## 更新履歴
- 2025-08-03: 複数のCI/CDドキュメントを統合して作成