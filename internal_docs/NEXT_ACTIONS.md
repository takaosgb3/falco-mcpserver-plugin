# 目的

直近スプリント（1–2 週）の実行項目を明確化し、MCP プラグイン MVP の着地を加速する。

# 次アクション（優先順）

- [ ] コード生成の実装強化：`scripts/gen-fields.sh` → フィールド定義/型/変換テーブル出力
- [ ] 抽出器雛形：SDKベースで `mcp.*` をフィールド登録（StringPool順守）
- [ ] ルール本線化：`rules/templates/mcp_baseline_rules.yaml` → `rules/mcp_baseline.yaml`
- [ ] E2E 接続：Falco 実行 or リプレイ導線から JSON アサーションを実行
- [ ] CI 可視化：Step Summary に検出統計/サマリ出力、`test-results/` 収集
- [ ] 運用外部化：許可リスト/しきい値の YAML 化（環境別上書き対応）

# 受入基準

- 生成コードを利用したユニットテストが通過
- 代表 E2E シナリオが JSON アサーションで一致
- ルール4本が本線に配置され CI で検証可能

