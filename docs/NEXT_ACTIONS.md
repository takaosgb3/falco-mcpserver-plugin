# 目的

直近スプリント（1–2 週）の実行項目を明確化し、MCP プラグイン MVP の着地を加速する。

# 次アクション（優先順・更新／タスク番号付与・P番号併記）

- [ ] [NX-1] WSモード実装（`mcp-test-server --mode ws`）と起動スクリプト追加（stdioは完了）（=P2-1.2, P2-1.6）
- [ ] [NX-2] WS経路の統合テスト追加（`mcp-audit-proxy` 経由）（=P2-2.2）
- [ ] [NX-3] `mcp-audit-wrap` の JSON-RPC `.method` 抽出と per-call イベント出力（オプション）（=P2-3）
- [ ] [NX-4] `mcp-audit-proxy` の JSON-RPC `.method` 抽出（可能な範囲、TLS透過維持）（=P2-3）
- [ ] [NX-5] CI 可視化の強化：Step Summary 集計（PASS/FAIL/SKIP件数）と `test-results/` 収集の拡充（=P2-4）
- [ ] [NX-6] 運用外部化：許可リスト/しきい値の YAML 化とルール生成フロー（軽量プリプロセス）（=P2-5）
- [ ] [NX-7] コード生成の実装強化：`scripts/gen-fields.sh` → フィールド定義/型/変換テーブル出力（=未割当／将来：P3候補）

# 受入基準

- [AC-1] テストサーバー（stdio/ws）から生成される監査JSONに期待フィールドが出現
- [AC-2] 代表シナリオ（S1〜S4）が JSON アサーションで一致
- [AC-3] CI で Step Summary/アーティファクトが確認でき、最低限のゲートが機能
