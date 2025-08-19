# 統合テストフレームワーク設計書

> 作成日: 2025-07-31
> 作成者: Claude Code
> Phase 2 Day 4実装

## 概要

このドキュメントは、Falco nginxプラグインの統合テストフレームワークの設計を記述します。
フレームワークは、プラグインがFalcoと正しく連携し、nginxログからセキュリティイベントを検出できることを検証します。

## テストアーキテクチャ

### 階層構造

```
test/
├── integration/              # 統合テストルート
│   ├── framework/           # テストフレームワークコア
│   │   ├── runner.go        # テストランナー
│   │   ├── falco.go         # Falco制御
│   │   ├── nginx.go         # nginx制御
│   │   └── validator.go     # 結果検証
│   ├── scenarios/           # テストシナリオ
│   │   ├── basic/          # 基本テスト
│   │   ├── security/       # セキュリティ検出テスト
│   │   ├── performance/    # パフォーマンステスト
│   │   └── stress/         # 負荷テスト
│   ├── fixtures/            # テストデータ
│   │   ├── logs/           # サンプルログ
│   │   ├── configs/        # 設定ファイル
│   │   └── rules/          # Falcoルール
│   └── helpers/            # ヘルパー関数
│       ├── log_generator.go # ログ生成
│       ├── event_matcher.go # イベントマッチング
│       └── metrics.go       # メトリクス収集
```

## コンポーネント設計

### 1. テストランナー (framework/runner.go)

```go
type IntegrationTestRunner struct {
    falcoController  *FalcoController
    nginxController  *NginxController
    validator        *EventValidator
    config          *TestConfig
}

type TestConfig struct {
    FalcoBinary     string
    PluginPath      string
    NginxConfigPath string
    LogPath         string
    Timeout         time.Duration
}

func (r *IntegrationTestRunner) Run(scenario TestScenario) (*TestResult, error) {
    // 1. 環境セットアップ
    // 2. シナリオ実行
    // 3. 結果収集
    // 4. クリーンアップ
}
```

### 2. Falco制御 (framework/falco.go)

```go
type FalcoController struct {
    binaryPath  string
    configPath  string
    process     *os.Process
    outputChan  chan *FalcoEvent
}

func (f *FalcoController) Start() error
func (f *FalcoController) Stop() error
func (f *FalcoController) WaitForEvent(timeout time.Duration) (*FalcoEvent, error)
func (f *FalcoController) LoadRules(rulesPath string) error
```

### 3. nginx制御 (framework/nginx.go)

```go
type NginxController struct {
    configPath  string
    logPath     string
    container   string  // Docker container ID
}

func (n *NginxController) Start() error
func (n *NginxController) Stop() error
func (n *NginxController) SendRequest(req *http.Request) (*http.Response, error)
func (n *NginxController) GetLogs() ([]string, error)
```

### 4. 結果検証 (framework/validator.go)

```go
type EventValidator struct {
    expectedEvents []ExpectedEvent
    actualEvents   []FalcoEvent
}

type ExpectedEvent struct {
    Rule        string
    Priority    string
    Output      map[string]interface{}
    TimeWindow  time.Duration
}

func (v *EventValidator) Validate() *ValidationResult
func (v *EventValidator) AddExpected(event ExpectedEvent)
func (v *EventValidator) AddActual(event FalcoEvent)
```

## テストシナリオ

### 1. 基本動作テスト

```go
// scenarios/basic/plugin_loading_test.go
func TestPluginLoading(t *testing.T) {
    // プラグインが正しくロードされることを確認
}

func TestBasicLogParsing(t *testing.T) {
    // 正常なログが正しくパースされることを確認
}

func TestFieldExtraction(t *testing.T) {
    // 17個のフィールドが正しく抽出されることを確認
}
```

### 2. セキュリティ検出テスト

```go
// scenarios/security/sql_injection_test.go
func TestSQLInjectionDetection(t *testing.T) {
    scenarios := []struct {
        name     string
        payload  string
        expected string
    }{
        {
            name:     "Basic SQL Injection",
            payload:  "/api/users?id=1' OR '1'='1",
            expected: "SQL injection attempt detected",
        },
        {
            name:     "Encoded SQL Injection",
            payload:  "/api/users?id=%31%27%20%4F%52%20%27%31%27%3D%27%31",
            expected: "SQL injection attempt detected",
        },
    }
}

// scenarios/security/xss_test.go
func TestXSSDetection(t *testing.T) {
    // XSS攻撃の検出テスト
}

// scenarios/security/path_traversal_test.go
func TestPathTraversalDetection(t *testing.T) {
    // パストラバーサル攻撃の検出テスト
}
```

### 3. パフォーマンステスト

```go
// scenarios/performance/throughput_test.go
func TestEventProcessingThroughput(t *testing.T) {
    // イベント処理スループットの測定
    // 目標: 2000 events/sec
}

func TestFieldExtractionLatency(t *testing.T) {
    // フィールド抽出レイテンシの測定
    // 目標: <0.5ms/event
}
```

### 4. 負荷テスト

```go
// scenarios/stress/high_volume_test.go
func TestHighVolumeLogProcessing(t *testing.T) {
    // 大量ログ処理時の安定性確認
    // 10万件/分のログを処理
}

func TestMemoryLeaks(t *testing.T) {
    // 長時間実行時のメモリリーク確認
}
```

## ヘルパー関数

### ログ生成器 (helpers/log_generator.go)

```go
type LogGenerator struct {
    format LogFormat
}

func (g *LogGenerator) GenerateNormalTraffic(count int) []string
func (g *LogGenerator) GenerateAttackTraffic(attackType string, count int) []string
func (g *LogGenerator) GenerateMixedTraffic(normalRatio float64, count int) []string
```

### イベントマッチャー (helpers/event_matcher.go)

```go
type EventMatcher struct {
    rules []MatchRule
}

func (m *EventMatcher) Match(event *FalcoEvent, expected ExpectedEvent) bool
func (m *EventMatcher) FindMissing(expected []ExpectedEvent, actual []FalcoEvent) []ExpectedEvent
func (m *EventMatcher) FindUnexpected(expected []ExpectedEvent, actual []FalcoEvent) []FalcoEvent
```

## CI/CD統合

### GitHub Actions ワークフロー

```yaml
name: Integration Tests

on:
  pull_request:
    paths:
      - 'pkg/**'
      - 'test/integration/**'
      - 'config/falco/**'

jobs:
  integration-test:
    runs-on: [self-hosted, linux, x64, local]
    steps:
      - uses: actions/checkout@v4

      - name: Setup Test Environment
        run: |
          ./scripts/setup-integration-test-env.sh

      - name: Run Integration Tests
        run: |
          go test -v ./test/integration/... -tags=integration

      - name: Generate Test Report
        if: always()
        run: |
          go test -json ./test/integration/... -tags=integration > test-report.json
          ./scripts/generate-test-report.sh test-report.json

      - name: Upload Test Results
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: integration-test-results
          path: |
            test-report.html
            test-report.json
```

## テスト実行

### ローカル実行

```bash
# 環境セットアップ
make setup-integration-test

# 全統合テストの実行
make test-integration

# 特定のシナリオの実行
go test -v ./test/integration/scenarios/security/... -tags=integration

# パフォーマンステストの実行
go test -v ./test/integration/scenarios/performance/... -tags=integration -bench=.
```

### Dockerでの実行

```bash
# Dockerコンテナでテスト環境を構築
docker-compose -f test/integration/docker-compose.yml up -d

# テスト実行
docker-compose -f test/integration/docker-compose.yml run test-runner

# クリーンアップ
docker-compose -f test/integration/docker-compose.yml down -v
```

## メトリクスとレポート

### 収集するメトリクス

1. **機能メトリクス**
   - テスト成功率
   - 検出精度（True Positive Rate）
   - 誤検知率（False Positive Rate）

2. **パフォーマンスメトリクス**
   - イベント処理スループット（events/sec）
   - フィールド抽出レイテンシ（ms/event）
   - メモリ使用量（MB）
   - CPU使用率（%）

3. **安定性メトリクス**
   - 連続実行時間
   - エラー発生率
   - リソースリーク

### レポート形式

```json
{
  "timestamp": "2025-07-31T10:00:00Z",
  "summary": {
    "total_tests": 50,
    "passed": 48,
    "failed": 2,
    "skipped": 0,
    "duration": "5m32s"
  },
  "scenarios": [
    {
      "name": "security/sql_injection",
      "status": "passed",
      "duration": "12.5s",
      "metrics": {
        "detection_rate": 0.98,
        "false_positive_rate": 0.01,
        "avg_latency_ms": 0.3
      }
    }
  ]
}
```

## エラーハンドリング

### タイムアウト処理

```go
func WithTimeout(timeout time.Duration, fn func() error) error {
    done := make(chan error, 1)
    go func() {
        done <- fn()
    }()

    select {
    case err := <-done:
        return err
    case <-time.After(timeout):
        return fmt.Errorf("operation timed out after %v", timeout)
    }
}
```

### リトライロジック

```go
func RetryWithBackoff(attempts int, delay time.Duration, fn func() error) error {
    for i := 0; i < attempts; i++ {
        if err := fn(); err == nil {
            return nil
        }

        if i < attempts-1 {
            time.Sleep(delay)
            delay *= 2
        }
    }
    return fmt.Errorf("failed after %d attempts", attempts)
}
```

## 今後の拡張

1. **追加シナリオ**
   - マルチテナント環境でのテスト
   - 高可用性構成でのテスト
   - Kubernetes環境でのテスト

2. **自動化の強化**
   - テストケースの自動生成
   - 回帰テストの自動実行
   - パフォーマンスベースラインの自動更新

3. **可視化**
   - リアルタイムテスト実行ダッシュボード
   - トレンド分析レポート
   - パフォーマンス比較グラフ

## まとめ

この統合テストフレームワークにより、Falco nginxプラグインの品質を包括的に検証できます。
継続的な改善とフィードバックループを通じて、より堅牢なセキュリティ監視ソリューションを提供します。