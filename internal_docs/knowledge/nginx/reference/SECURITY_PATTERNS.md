# 🛡️ セキュリティパターンリファレンス

> 最終更新: 2025-08-03
> カテゴリ: リファレンス

## 📋 目次

1. [攻撃パターン一覧](#攻撃パターン一覧)
2. [検出ルール](#検出ルール)
3. [Falcoルール例](#falcoルール例)
4. [ログサンプル](#ログサンプル)

## 攻撃パターン一覧

### SQLインジェクション

| パターン | 説明 | 危険度 |
|---------|------|--------|
| `' OR '1'='1` | 基本的な認証バイパス | HIGH |
| `UNION SELECT` | データ抽出 | HIGH |
| `'; DROP TABLE` | データ破壊 | CRITICAL |
| `/**/` | コメントを使った回避 | MEDIUM |
| `CONCAT(CHAR(`,`CHAR(` | 文字列結合での回避 | MEDIUM |

### XSS（クロスサイトスクリプティング）

| パターン | 説明 | 危険度 |
|---------|------|--------|
| `<script>alert(` | 基本的なXSS | HIGH |
| `javascript:` | URLベースXSS | HIGH |
| `onerror=` | イベントハンドラXSS | HIGH |
| `<svg onload=` | SVGを使ったXSS | HIGH |
| `&#x3C;script&#x3E;` | エンコードされたXSS | MEDIUM |

### ディレクトリトラバーサル

| パターン | 説明 | 危険度 |
|---------|------|--------|
| `../../../etc/passwd` | 基本的なトラバーサル | HIGH |
| `..%2F..%2F` | URLエンコード | HIGH |
| `..%252f..%252f` | 二重エンコード | MEDIUM |
| `..%c0%af` | Unicodeエンコード | MEDIUM |
| `..\..\..\` | Windowsスタイル | HIGH |

### コマンドインジェクション

| パターン | 説明 | 危険度 |
|---------|------|--------|
| `; cat /etc/passwd` | 基本的なコマンド実行 | CRITICAL |
| `\| id` | パイプを使った実行 | CRITICAL |
| `$(whoami)` | コマンド置換 | CRITICAL |
| `` `id` `` | バッククォート | CRITICAL |
| `&& ls -la` | 論理演算子 | HIGH |

### 悪意のあるUser-Agent

| パターン | 説明 | 危険度 |
|---------|------|--------|
| `sqlmap/` | SQLインジェクションツール | HIGH |
| `nikto/` | 脆弱性スキャナー | MEDIUM |
| `nmap` | ポートスキャナー | MEDIUM |
| `masscan` | 高速スキャナー | MEDIUM |
| `vulnerability scanner` | 一般的なスキャナー | LOW |

## 検出ルール

### Go実装例

```go
package detection

import (
    "regexp"
    "strings"
)

type ThreatDetector struct {
    patterns map[string][]*regexp.Regexp
}

func NewThreatDetector() *ThreatDetector {
    td := &ThreatDetector{
        patterns: make(map[string][]*regexp.Regexp),
    }

    // SQLインジェクションパターン
    td.patterns["sql_injection"] = []*regexp.Regexp{
        regexp.MustCompile(`(?i)union.*select`),
        regexp.MustCompile(`(?i)select.*from.*where`),
        regexp.MustCompile(`(?i)'.*or.*'='`),
        regexp.MustCompile(`(?i)drop\s+table`),
        regexp.MustCompile(`(?i)insert\s+into`),
        regexp.MustCompile(`(?i)update.*set`),
        regexp.MustCompile(`(?i)delete\s+from`),
    }

    // XSSパターン
    td.patterns["xss"] = []*regexp.Regexp{
        regexp.MustCompile(`(?i)<script[^>]*>`),
        regexp.MustCompile(`(?i)javascript:`),
        regexp.MustCompile(`(?i)on\w+\s*=`),
        regexp.MustCompile(`(?i)<iframe`),
        regexp.MustCompile(`(?i)document\.cookie`),
    }

    // ディレクトリトラバーサル
    td.patterns["traversal"] = []*regexp.Regexp{
        regexp.MustCompile(`\.\./`),
        regexp.MustCompile(`\.\.\\`),
        regexp.MustCompile(`%2e%2e/`),
        regexp.MustCompile(`%252e%252e`),
    }

    return td
}

func (td *ThreatDetector) Detect(input string) []string {
    var threats []string

    for threatType, patterns := range td.patterns {
        for _, pattern := range patterns {
            if pattern.MatchString(input) {
                threats = append(threats, threatType)
                break
            }
        }
    }

    return threats
}
```

### 最適化された検出関数

```go
// パフォーマンスを考慮した実装
func OptimizedDetect(input string) ThreatType {
    // 早期リターンで高速化
    if len(input) > 10000 {
        return ThreatType{Type: "oversized_input", Severity: "MEDIUM"}
    }

    // 小文字変換は一度だけ
    lowerInput := strings.ToLower(input)

    // 最も頻繁な攻撃を先にチェック
    if strings.Contains(lowerInput, "union") && strings.Contains(lowerInput, "select") {
        return ThreatType{Type: "sql_injection", Severity: "HIGH"}
    }

    if strings.Contains(lowerInput, "<script") {
        return ThreatType{Type: "xss", Severity: "HIGH"}
    }

    if strings.Contains(input, "../") || strings.Contains(input, "..\\") {
        return ThreatType{Type: "directory_traversal", Severity: "HIGH"}
    }

    return ThreatType{Type: "none", Severity: "NONE"}
}
```

## Falcoルール例

### 基本的な検出ルール

```yaml
- rule: Suspicious nginx access
  desc: Detect potential attacks in nginx logs
  condition: >
    nginx.evt_type = "access" and (
      nginx.request contains "union select" or
      nginx.request contains "<script" or
      nginx.request contains "../.." or
      nginx.request contains "etc/passwd"
    )
  output: >
    Suspicious request detected
    (client=%nginx.remote_addr method=%nginx.method
     uri=%nginx.request status=%nginx.status)
  priority: WARNING
  tags: [web, nginx, attack]

- rule: SQL injection attempt
  desc: Detect SQL injection patterns
  condition: >
    nginx.evt_type = "access" and
    nginx.request regex "(?i)(union.*select|select.*from|'.*or.*'=')"
  output: >
    SQL injection attempt
    (client=%nginx.remote_addr uri=%nginx.request
     user_agent=%nginx.user_agent)
  priority: ERROR
  tags: [web, nginx, sql_injection]

- rule: Excessive 404 errors
  desc: Detect scanning behavior
  condition: >
    nginx.evt_type = "access" and
    nginx.status = 404 and
    count_over_time(%nginx.remote_addr, 60s) > 20
  output: >
    Possible scanning activity
    (client=%nginx.remote_addr count=%count)
  priority: WARNING
  tags: [web, nginx, scanning]
```

### 高度な検出ルール

```yaml
- rule: Advanced threat detection
  desc: Multi-factor threat analysis
  condition: >
    nginx.evt_type = "access" and (
      # 複数の攻撃指標
      (nginx.request contains "select" and
       nginx.request contains "from" and
       nginx.status >= 500) or

      # 疑わしいUser-Agent
      (nginx.user_agent in (sqlmap, nikto, nmap) and
       nginx.status = 200) or

      # 大量データ流出の可能性
      (nginx.body_bytes_sent > 10000000 and
       nginx.request contains "dump")
    )
  output: >
    Advanced threat detected
    (type=%threat.type client=%nginx.remote_addr
     uri=%nginx.request bytes=%nginx.body_bytes_sent)
  priority: CRITICAL
  tags: [web, nginx, advanced_threat]

- rule: Geolocation anomaly
  desc: Access from unexpected location
  condition: >
    nginx.evt_type = "access" and
    nginx.remote_addr not in trusted_networks and
    geoip(%nginx.remote_addr).country not in allowed_countries
  output: >
    Access from unexpected location
    (client=%nginx.remote_addr country=%geoip.country
     uri=%nginx.request)
  priority: WARNING
  tags: [web, nginx, geolocation]
```

## ログサンプル

### 正常なアクセス

```
192.168.1.100 - - [03/Aug/2025:10:15:30 +0000] "GET /api/users HTTP/1.1" 200 1234 "-" "Mozilla/5.0"
192.168.1.101 - - [03/Aug/2025:10:15:31 +0000] "POST /api/login HTTP/1.1" 200 456 "-" "Mozilla/5.0"
192.168.1.102 - - [03/Aug/2025:10:15:32 +0000] "GET /static/css/main.css HTTP/1.1" 200 8901 "-" "Mozilla/5.0"
```

### 攻撃パターンを含むログ

```
# SQLインジェクション
10.0.0.1 - - [03/Aug/2025:10:20:15 +0000] "GET /api/users?id=1' OR '1'='1 HTTP/1.1" 500 0 "-" "sqlmap/1.2.3"
10.0.0.1 - - [03/Aug/2025:10:20:16 +0000] "GET /api/users?id=1 UNION SELECT password FROM users-- HTTP/1.1" 500 0 "-" "sqlmap/1.2.3"

# XSS
10.0.0.2 - - [03/Aug/2025:10:21:30 +0000] "GET /search?q=<script>alert('XSS')</script> HTTP/1.1" 200 5432 "-" "Mozilla/5.0"
10.0.0.2 - - [03/Aug/2025:10:21:31 +0000] "POST /comment HTTP/1.1" 200 234 "-" "Mozilla/5.0"
# POSTデータ: text=<img src=x onerror=alert(1)>

# ディレクトリトラバーサル
10.0.0.3 - - [03/Aug/2025:10:22:45 +0000] "GET /download?file=../../../../etc/passwd HTTP/1.1" 403 0 "-" "curl/7.64.0"
10.0.0.3 - - [03/Aug/2025:10:22:46 +0000] "GET /files?path=..%2F..%2F..%2Fetc%2Fpasswd HTTP/1.1" 403 0 "-" "curl/7.64.0"

# スキャニング
10.0.0.4 - - [03/Aug/2025:10:23:00 +0000] "GET /admin HTTP/1.1" 404 0 "-" "nikto/2.1.5"
10.0.0.4 - - [03/Aug/2025:10:23:01 +0000] "GET /backup HTTP/1.1" 404 0 "-" "nikto/2.1.5"
10.0.0.4 - - [03/Aug/2025:10:23:02 +0000] "GET /.git HTTP/1.1" 404 0 "-" "nikto/2.1.5"
```

### 検出結果の例

```json
{
  "timestamp": "2025-08-03T10:20:15Z",
  "source_ip": "10.0.0.1",
  "threat_type": "sql_injection",
  "severity": "HIGH",
  "request": "GET /api/users?id=1' OR '1'='1",
  "user_agent": "sqlmap/1.2.3",
  "response_status": 500,
  "action_taken": "blocked",
  "details": {
    "pattern_matched": "' OR '1'='1",
    "rule": "SQL injection attempt"
  }
}
```

---

## 使用方法

1. **パターンの選択**: 環境に応じて必要なパターンを選択
2. **しきい値の調整**: 誤検知を減らすため適切に調整
3. **定期的な更新**: 新しい攻撃パターンを追加
4. **テスト**: 本番環境適用前に十分なテスト

## 関連ドキュメント
- [セキュリティガイド](../development/SECURITY_UNIFIED.md)
- [Falco統合ガイド](./falco-integration.md)
- [ログ分析ガイド](./log-analysis.md)