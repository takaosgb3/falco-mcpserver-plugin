---
title: トラブルシューティングガイド
description: Falco nginx プラグインプロジェクトで発生する可能性のある問題と解決方法
category: operations
tags: [troubleshooting, debugging, errors, solutions]
status: active
priority: high
---

# トラブルシューティングガイド

## 概要

このドキュメントは、Falco nginx プラグインプロジェクトで発生する可能性のある問題と、その解決方法について説明します。

## 一般的な問題と解決方法

### Falco 関連

#### 問題: Falco が起動しない
**症状:**
- `systemctl status falco` で失敗状態が表示される
- ログに "Failed to load rules" エラーが出る

**解決方法:**
1. 設定ファイルの構文を確認
   ```bash
   falco --validate /etc/falco/falco.yaml
   ```
2. ルールファイルの権限を確認
   ```bash
   sudo chown -R root:root /etc/falco
   sudo chmod 644 /etc/falco/*.yaml
   ```

### nginx 関連

#### 問題: nginx が起動しない
**症状:**
- `nginx -t` でエラーが表示される
- ポート 80 がすでに使用されている

**解決方法:**
1. 設定ファイルの構文を確認
   ```bash
   nginx -t
   ```
2. ポートの使用状況を確認
   ```bash
   sudo lsof -i :80
   ```

### JSON エンコーディングエラー

#### 問題: update-progress.sh 実行時に JSON パースエラーが発生
**症状:**
- "parse error: Invalid numeric literal at line 1, column 9" エラー
- "no low surrogate in string" などの UTF-16 エラー
- API Error 400 が発生

**原因:**
- Unicode 文字（絵文字など）が適切にエンコードされていない
- `jq` コマンドがインストールされていない

**解決方法:**
1. `jq` をインストール
   ```bash
   # Ubuntu/Debian
   sudo apt-get install -y jq

   # macOS
   brew install jq

   # CentOS/RHEL
   sudo yum install -y jq
   ```

2. スクリプトが正しく動作するか確認
   ```bash
   ./scripts/update-progress.sh
   ```

3. 生成された JSON ファイルが有効か確認
   ```bash
   jq . .progress-data.json
   ```

**予防策:**
- シェルスクリプトで JSON を生成する際は、必ず `jq` を使用する
- Unicode 文字を含む変数は適切にエスケープする

### GitHub Actions 関連

#### 問題: ワークフローが失敗する
**症状:**
- Actions タブで赤い × マークが表示される
- "Permission denied" エラーが出る

**解決方法:**
1. リポジトリの Actions 権限を確認
2. シークレットが正しく設定されているか確認
3. ワークフローファイルの構文を確認

## デバッグ方法

### ログの確認

#### Falco ログ
```bash
# systemd ログ
sudo journalctl -u falco -f

# Falco 直接実行（デバッグモード）
sudo falco -o log_level=debug
```

#### nginx ログ
```bash
# アクセスログ
tail -f /var/log/nginx/access.log

# エラーログ
tail -f /var/log/nginx/error.log
```

### プロセスの確認
```bash
# Falco プロセス
ps aux | grep falco

# nginx プロセス
ps aux | grep nginx
```

## よくある質問 (FAQ)

### Q: Falco のルールを追加したが反映されない
A: Falco サービスを再起動してください:
```bash
sudo systemctl restart falco
```

### Q: nginx の設定変更が反映されない
A: nginx をリロードしてください:
```bash
sudo nginx -s reload
```

### Q: JSON ファイルの生成で絵文字が文字化けする
A: `jq` がインストールされているか確認し、`update-progress.sh` スクリプトが最新版であることを確認してください。

### Q: Progress Widget の更新が失敗する
A: 以下の手順で問題を解決してください:

1. **権限エラーの場合**:
   ```bash
   # ファイル権限を確認
   ls -la docs/project/PROGRESS_WIDGET.md

   # 権限を修正
   chmod 644 docs/project/PROGRESS_WIDGET.md
   ```

2. **GitHub Actions ワークフローが失敗する場合**:
   - ワークフローの権限設定を確認
   - レート制限に達していないか確認
   - スクリプトのバリデーション:
     ```bash
     shellcheck scripts/update-progress-widget.sh
     ./scripts/test-progress-widget.sh
     ```

3. **進捗計算が正しくない場合**:
   ```bash
   # 手動で再計算
   make progress-calc

   # ダッシュボードのフォーマットを確認
   grep -E "Phase [0-9]:.*[0-9]+%" docs/project/PROGRESS_DASHBOARD.md
   ```

4. **手動更新が必要な場合**:
   ```bash
   # 現在の進捗を計算
   ./scripts/calculate-progress.sh

   # ドライラン（変更なし）
   ./scripts/update-progress-widget.sh

   # 実際の更新
   ./scripts/update-progress-widget.sh --update
   ```

5. **スクリプトのデバッグ**:
   ```bash
   # デバッグモードで実行
   bash -x scripts/update-progress-widget.sh

   # 構文エラーをチェック
   bash -n scripts/update-progress-widget.sh
   ```

## サポート

問題が解決しない場合は、以下の情報を含めて Issue を作成してください:

1. エラーメッセージの全文
2. 実行したコマンド
3. 環境情報（OS、バージョンなど）
4. 関連するログファイル

[Issue を作成](https://github.com/takaosgb3/falco-nginx-plugin-claude/issues/new)