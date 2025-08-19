# 目的

直近スプリント（1–2 週）の実行項目を明確化し、MCP プラグイン MVP の着地を加速する。

# 次アクション（優先順・更新／タスク番号付与）

- [ ] [NX-1] テスト用 MCP Server 最小実装（stdio/ws）と起動スクリプト追加
- [ ] [NX-2] 代表シナリオ（S1〜S4）の自動化（検証スクリプト/期待値）
- [ ] [NX-3] `mcp-audit-wrap` の JSON-RPC `.method` 抽出と per-call イベント出力（オプション）
- [ ] [NX-4] `mcp-audit-proxy` の JSON-RPC `.method` 抽出（可能な範囲、TLS透過維持）
- [ ] [NX-5] CI 可視化の実装：Step Summary 出力＋`actions/upload-artifact` で `test-results/` 収集
- [ ] [NX-6] 運用外部化：許可リスト/しきい値の YAML 化とルール生成フロー（軽量プリプロセス）
- [ ] [NX-7] コード生成の実装強化：`scripts/gen-fields.sh` → フィールド定義/型/変換テーブル出力

# 受入基準

- [AC-1] テストサーバー（stdio/ws）から生成される監査JSONに期待フィールドが出現
- [AC-2] 代表シナリオ（S1〜S4）が JSON アサーションで一致
- [AC-3] CI で Step Summary/アーティファクトが確認でき、最低限のゲートが機能
