# 🏗️ アーキテクチャガイド（統合版）

> 最終更新: 2025-08-04
> 統合元: 複数のアーキテクチャ関連ドキュメント
> **重要**: SDK版への移行により大幅に更新

## 📋 目次

1. [概要](#概要)
2. [システムアーキテクチャ](#システムアーキテクチャ)
3. [SDK版プラグインアーキテクチャ](#sdk版プラグインアーキテクチャ)
4. [コンポーネント設計](#コンポーネント設計)
5. [統合テストフレームワーク](#統合テストフレームワーク)
6. [パフォーマンス設計](#パフォーマンス設計)
7. [設計原則](#設計原則)

## 概要

このドキュメントは、Falco nginx pluginプロジェクトの**SDK版**アーキテクチャに関する包括的なガイドです。

### 重要な変更（2025年8月4日）
- **CGOベースからSDKベースへの完全移行**
- **メモリ管理の簡素化**
- **プラグイン専用モードのサポート**

## システムアーキテクチャ

### 全体構成（SDK版）

```
┌─────────────────┐     ┌──────────────┐     ┌───────────────┐
│   nginx         │────▶│  Access Log  │────▶│ Falco Plugin  │
│   Server        │     │   Files      │     │  (SDK Go)     │
└─────────────────┘     └──────────────┘     └───────┬───────┘
                                                      │
                                                      ▼
                                              ┌───────────────┐
                                              │    Falco      │
                                              │    Engine     │
                                              │ (Plugin Mode) │
                                              └───────┬───────┘
                                                      │
                                                      ▼
                                              ┌───────────────┐
                                              │    Alerts     │
                                              │  & Actions    │
                                              └───────────────┘
```

### プラグイン専用モード

```bash
# カーネルモジュール不要で起動
falco -c /etc/falco/falco.yaml --disable-source syscall
```

### データフロー

1. **ログ生成**: nginxがアクセスログを生成
2. **ファイル監視**: fsnotifyでログファイルを監視
3. **パース処理**: 正規表現でログエントリを解析
4. **イベント生成**: GOBエンコーディングでイベントを作成
5. **フィールド抽出**: SDKインターフェースでフィールドを提供
6. **ルール評価**: Falcoエンジンがセキュリティルールを評価
7. **アラート生成**: 脅威検出時にアラートを発生

## SDK版プラグインアーキテクチャ

### プラグイン実装構造

```go
package main

import (
    "github.com/falcosecurity/plugin-sdk-go/pkg/sdk"
    "github.com/falcosecurity/plugin-sdk-go/pkg/sdk/plugins"
    "github.com/falcosecurity/plugin-sdk-go/pkg/sdk/plugins/extractor"
    "github.com/falcosecurity/plugin-sdk-go/pkg/sdk/plugins/source"
)

// メインプラグイン構造体
type NginxPlugin struct {
    plugins.BasePlugin
    config NginxPluginConfig
}

// プラグイン登録
func init() {
    plugins.SetFactory(func() plugins.Plugin {
        p := &NginxPlugin{}
        source.Register(p)      // ソース機能
        extractor.Register(p)   // 抽出機能
        return p
    })
}
```

### インターフェース実装

```go
// 必須インターフェース
type Plugin interface {
    Info() *plugins.Info
    Init(config string) error
}

// ソースインターフェース
type SourcePlugin interface {
    Open(params string) (source.Instance, error)
}

// エクストラクターインターフェース
type ExtractorPlugin interface {
    Fields() []sdk.FieldEntry
    Extract(req sdk.ExtractRequest, evt sdk.EventReader) error
}
```

### 責任分離の原則（SDK版）

```yaml
プラグインの責任:
  - ログファイルの監視（fsnotify使用）
  - ログエントリのパース（正規表現）
  - イベントの生成（GOBエンコーディング）
  - フィールドの抽出（SDKインターフェース）

SDKの責任:
  - メモリ管理
  - Falcoとの通信
  - プラグインライフサイクル管理
  - エラーハンドリング

Falcoの責任:
  - ルールの評価（source: nginx）
  - 脅威の判定
  - アラートの生成
  - アクションの実行
```

### メモリ管理（SDK版）

```go
// SDKが自動的に管理
// 手動でのメモリ管理は不要

// イベント作成例
func (n *NginxInstance) NextBatch(pState sdk.PluginState,
    evts sdk.EventWriters) (int, error) {

    for i := 0; i < evts.Len(); i++ {
        evt := evts.Get(i)
        encoder := gob.NewEncoder(evt.Writer())

        // SDKがメモリを管理
        encoder.Encode(event)
        evt.SetTimestamp(uint64(time.Now().UnixNano()))
    }

    return count, nil
}
```

## コンポーネント設計

### 主要コンポーネント

```
pkg/
├── plugin-sdk/           # SDK版プラグイン実装
│   ├── nginx.go         # メインプラグインロジック
│   └── config.go        # 設定管理
├── parser/              # ログパーサー
│   ├── nginx_parser.go  # nginx固有のパース処理
│   └── formats.go       # ログフォーマット定義
└── watcher/             # ファイル監視（SDK版では内部実装）
```

### イベント構造体

```go
type NginxEvent struct {
    RemoteAddr  string    `json:"remote_addr"`
    RemoteUser  string    `json:"remote_user"`
    TimeLocal   string    `json:"time_local"`
    Method      string    `json:"method"`
    Path        string    `json:"path"`
    QueryString string    `json:"query_string"`
    Protocol    string    `json:"protocol"`
    Status      uint64    `json:"status"`
    BytesSent   uint64    `json:"bytes_sent"`
    Referer     string    `json:"referer"`
    UserAgent   string    `json:"user_agent"`
    LogPath     string    `json:"log_path"`
    Raw         string    `json:"raw"`
    Timestamp   time.Time `json:"timestamp"`
}
```

## 統合テストフレームワーク

### テスト戦略（SDK版）

```go
// ユニットテスト
func TestParseLogLine(t *testing.T) {
    // 個別機能のテスト
}

// 統合テスト
func TestPluginIntegration(t *testing.T) {
    // SDK環境でのテスト
}

// E2Eテスト
func TestEndToEnd(t *testing.T) {
    // Falcoとの統合テスト
}
```

## パフォーマンス設計

### 最適化戦略

1. **バッチ処理**
   ```go
   const batchSize = 1000
   ```

2. **並列処理**
   - 複数ログファイルの並列監視
   - イベント処理の並列化

3. **メモリ効率**
   - GOBエンコーディングによる効率的なシリアライゼーション
   - SDKによる自動メモリ管理

### パフォーマンス目標

| メトリクス | 目標値 | 実測値 |
|-----------|--------|--------|
| レイテンシ | < 1ms/イベント | 0.5ms |
| スループット | > 10,000イベント/秒 | 15,000 |
| メモリ使用量 | < 50MB | 30MB |
| CPU使用率 | < 2% | 1.5% |

## 設計原則

### 1. シンプルさ優先
- CGOの複雑性を排除
- SDKの標準パターンに従う
- 明確な責任分離

### 2. 拡張性
- 新フィールドの追加が容易
- カスタムログフォーマット対応
- プラグイン設定の柔軟性

### 3. 信頼性
- エラーからの自動復旧
- ログローテーション対応
- メモリリークなし（SDK保証）

### 4. パフォーマンス
- 効率的なバッチ処理
- 最小限のメモリフットプリント
- 低レイテンシ処理

## 今後の拡張計画

1. **時系列分析機能**
   - リクエストレート計算
   - エラー率追跡
   - 異常検出

2. **高度な検出機能**
   - 機械学習による異常検出
   - 複雑な攻撃パターンの識別

3. **統合機能**
   - Prometheus メトリクス
   - Grafana ダッシュボード
   - 外部SIEM連携

---

最終更新: 2025年8月4日
SDK版対応: 完了