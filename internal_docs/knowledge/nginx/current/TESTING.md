# 🧪 テストガイド（統合版）

> 最終更新: 2025-08-03
> 統合元: 複数のテスト関連ドキュメント

## 📋 目次

1. [概要](#概要)
2. [テスト戦略](#テスト戦略)
3. [テスト実装ガイド](#テスト実装ガイド)
4. [ベストプラクティス](#ベストプラクティス)
5. [よくある落とし穴と解決策](#よくある落とし穴と解決策)
6. [カバレッジ分析](#カバレッジ分析)
7. [統合テストフレームワーク](#統合テストフレームワーク)
8. [E2Eテスト](#e2eテスト)

## 概要

このドキュメントは、Falco nginx pluginプロジェクトのテストに関する包括的なガイドです。

### 統合されたドキュメント
- TEST_IMPLEMENTATION_BEST_PRACTICES.md
- TEST_IMPLEMENTATION_LESSONS_LEARNED.md
- TEST_PITFALLS_AND_SOLUTIONS.md
- TEST_COVERAGE_ANALYSIS.md
- TESTING.md
- E2E_TEST_WALKTHROUGH.md

## テスト戦略

### テストピラミッド
```
         /\
        /E2E\      (5%)  - エンドツーエンドテスト
       /------\
      /統合テスト\   (20%) - Falco統合、API連携
     /----------\
    /単体テスト    \  (75%) - 関数レベル、高速実行
   /--------------\
```

### テストカテゴリ

| カテゴリ | 目的 | ツール | 実行時間 |
|---------|------|--------|----------|
| 単体テスト | 関数の正確性 | go test | <1秒 |
| 統合テスト | コンポーネント連携 | go test + Docker | <30秒 |
| E2Eテスト | 実環境動作確認 | Docker Compose | <5分 |
| 性能テスト | パフォーマンス | go test -bench | <2分 |
| セキュリティテスト | 脆弱性検出 | 専用ツール | <10分 |

## テスト実装ガイド

### 1. 単体テストの書き方

```go
func TestParseNginxLog(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected *LogEntry
        wantErr  bool
    }{
        {
            name:  "valid combined log",
            input: `192.168.1.1 - - [01/Aug/2024:12:34:56 +0000] "GET /api HTTP/1.1" 200 1234`,
            expected: &LogEntry{
                RemoteAddr: "192.168.1.1",
                Method:     "GET",
                Path:       "/api",
                Status:     200,
                BodyBytes:  1234,
            },
            wantErr: false,
        },
        // エッジケース、エラーケースを追加
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseNginxLog(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseNginxLog() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.expected) {
                t.Errorf("ParseNginxLog() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

### 2. CGO関連テストの対処

CGOを使用する関数は直接テストできないため、内部ロジックを分離：

```go
// ❌ 悪い例
//export plugin_init
func plugin_init(config *C.char, rc *int32) unsafe.Pointer {
    // テスト不可能
}

// ✅ 良い例
//export plugin_init
func plugin_init(config *C.char, rc *int32) unsafe.Pointer {
    cfg := C.GoString(config)
    state, err := initializePlugin(cfg) // テスト可能な関数
    if err != nil {
        *rc = 1
        return nil
    }
    *rc = 0
    return unsafe.Pointer(state)
}

func initializePlugin(config string) (*pluginState, error) {
    // テスト可能なロジック
}
```

## ベストプラクティス

### 1. テーブルドリブンテスト
```go
tests := []struct {
    name     string
    input    interface{}
    expected interface{}
    wantErr  bool
}{
    // テストケース
}
```

### 2. 並行実行対応
```go
t.Run(tt.name, func(t *testing.T) {
    t.Parallel() // 並行実行を許可
    // テストロジック
})
```

### 3. テストヘルパー
```go
func setupTest(t *testing.T) (*TestEnv, func()) {
    t.Helper()
    env := &TestEnv{
        // セットアップ
    }
    cleanup := func() {
        // クリーンアップ
    }
    return env, cleanup
}
```

### 4. モックとスタブ
```go
type MockLogger struct {
    mock.Mock
}

func (m *MockLogger) Log(level, message string) {
    m.Called(level, message)
}
```

## よくある落とし穴と解決策

### 1. タイムゾーン依存
**問題**: CI環境でタイムゾーンが異なる
```go
// ❌ 環境依存
time.Parse("02/Jan/2006:15:04:05 -0700", "01/Aug/2024:12:34:56 JST")

// ✅ 固定値使用
time.Parse("02/Jan/2006:15:04:05 -0700", "01/Aug/2024:12:34:56 +0000")
```

### 2. ファイルパス
**問題**: OS間でパス区切り文字が異なる
```go
// ❌ ハードコード
path := "testdata/logs/access.log"

// ✅ filepath使用
path := filepath.Join("testdata", "logs", "access.log")
```

### 3. 並行実行の競合
**問題**: グローバル変数の競合状態
```go
// ❌ グローバル変数
var counter int

// ✅ ローカルスコープ
func TestConcurrent(t *testing.T) {
    var counter int32
    // atomic操作を使用
}
```

### 4. リソースリーク
**問題**: ファイルやコネクションのクローズ忘れ
```go
// ✅ defer使用
f, err := os.Open(filename)
if err != nil {
    t.Fatal(err)
}
defer f.Close()
```

## カバレッジ分析

### 現在のカバレッジ状況
```
Package               Coverage
parser                88.5%    ✅
plugin                4.3%     ⚠️  (CGO制約)
watcher               68.6%    ✅
Overall               ~35%
```

### カバレッジ向上戦略
1. **parser**: エッジケースの追加でこaバレッジ向上
2. **plugin**: 内部ロジックの分離でテスト可能に
3. **watcher**: エラーケースのテスト追加

### カバレッジ計測
```bash
# HTMLレポート生成
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# 関数別カバレッジ
go tool cover -func=coverage.out

# 未テストコードの特定
go tool cover -func=coverage.out | grep -E "0.0%|0%"
```

## 統合テストフレームワーク

### アーキテクチャ
```
┌─────────────┐     ┌──────────────┐     ┌───────────┐
│   テスト    │────▶│ テストランナー │────▶│   Falco   │
│  シナリオ   │     │   (Docker)    │     │ コンテナ  │
└─────────────┘     └──────────────┘     └───────────┘
                            │
                            ▼
                    ┌──────────────┐
                    │    nginx     │
                    │  コンテナ    │
                    └──────────────┘
```

### 実装例
```go
func TestFalcoIntegration(t *testing.T) {
    ctx := context.Background()

    // Docker環境セットアップ
    env := setupDockerEnv(t)
    defer env.Cleanup()

    // nginxログ生成
    generateTestLogs(t, env.NginxContainer)

    // Falcoイベント確認
    events := waitForFalcoEvents(t, env.FalcoContainer, 30*time.Second)

    // アサーション
    assert.Contains(t, events, "Suspicious nginx activity detected")
}
```

## E2Eテスト

### シナリオ例
1. **正常系フロー**
   - nginx起動
   - 通常アクセス
   - ログ生成確認
   - Falcoアラートなし

2. **攻撃検出フロー**
   - SQLインジェクション試行
   - XSS試行
   - Falcoアラート確認
   - アラート内容検証

### 実行方法
```bash
# E2Eテスト環境構築
docker-compose -f test/e2e/docker-compose.yml up -d

# テスト実行
go test ./test/e2e/... -tags=e2e

# クリーンアップ
docker-compose -f test/e2e/docker-compose.yml down
```

## CI/CDでのテスト実行

### GitHub Actions設定
```yaml
test:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4

    - name: Run Unit Tests
      run: go test ./... -race -cover

    - name: Run Integration Tests
      run: |
        docker-compose up -d
        go test ./... -tags=integration

    - name: Upload Coverage
      uses: codecov/codecov-action@v3
```

## トラブルシューティング

### テストが失敗する場合
1. **環境変数の確認**
   ```bash
   go env | grep -E "GOPROXY|GOSUMDB"
   ```

2. **依存関係のクリーンアップ**
   ```bash
   go clean -modcache
   go mod download
   ```

3. **詳細ログの出力**
   ```bash
   go test -v -race ./...
   ```

### CI特有の問題
- タイムアウト：テストタイムアウトを調整
- メモリ不足：並行実行数を制限
- 権限エラー：Dockerソケットの権限確認

---

## 関連ドキュメント
- [開発ガイド](./README.md)
- [CI/CDガイド](./CI_CD.md)
- [運用ガイド](./OPERATIONS.md)

## 更新履歴
- 2025-08-03: 複数ドキュメントを統合して作成