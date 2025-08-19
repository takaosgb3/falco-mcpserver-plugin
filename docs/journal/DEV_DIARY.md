# 開発日記（Development Diary）

- 趣旨: 進捗・意思決定・変更点・課題・次アクションを日次で簡潔に記録し、トレーサビリティと合意形成を高める。
- 運用: 1日1節。太字で日付、箇条書きは最大6項目を目安。関係ID（PH/MS/Px/NX/AC）を併記。

## 2025-08-19

- 立上げ（要件/計画/知見取り込み）
  - 要件定義を追加（docs/requirements/...）。監査はメタデータ中心、eBPFはPhase4（将来）に位置付け（PH-1/PH-4）。
  - nginxプラグインの内部ドキュメントを参照用に整理し、公開向けにdocsへ再編集（INDEXに導線）。
  - プロジェクト計画（PROJECT_PLAN.md）と次アクション（NEXT_ACTIONS.md）を初稿（PH-1, NX-*）。

- アーキテクチャ/設計
  - OVERVIEWをMECEで全面リライト（構成/データ/フロー/非機能/導入/受入/リスク/次アクション）。
  - DETAILEDに設定例・ルール抜粋・出力例・アサーション断片を追加。
  - PRODUCERSに方式別（stdio/ws/client-emitter）の起動/シーケンス/擬似コードを整理。

- CI/テスト方針
  - CIはGitHub Actionsランナー（ubuntu-latest）。Pushは無効化し、PR/手動で実行（CI_DESIGN.md）。
  - E2Eは当面オフライン/スキップ設計。将来、実行系は別ジョブ化予定（PH-2/PH-3）。

- 監査Producer（MVP）実装（PH-1 完了）
  - mcp-audit-wrap（stdio中継・双方向バイト集計）、mcp-audit-proxy（tcp/ws透過・TLS非終端）を追加（P1-2）。
  - 監査イベント出力ヘルパ（pkg/audit/event.go）。
  - ルール最小4本（テンプレ/本線）配置（P1-3）。

- テスト用MCPサーバ/統合テスト（PH-2 着手）
  - mcp-test-server（stdio）を追加（P2-1.1）。tools.list/prompts.get/tools.exec、遅延/サイズ/エラー制御。
  - Actionsランナー統合テスト（stdio_wrap_basic.sh）を実行に組込み、成果物をアップロード（P2-2.1, P2-4）。

- 優先タスクの更新
  - WSモード実装（P2-1.2 / NX-1）、WS統合テスト（P2-2.2 / NX-2）、`.method`抽出/per-call（P2-3 / NX-3, NX-4）、CI可視化強化（NX-5）。

## 2025-08-18（サマリ補完）

- リポジトリ初期化と公開ドキュメント方針
  - GitHubリポジトリ作成、docs/ 配下に公開用ドキュメントを集約。internal_docs/ はローカル参照のみ（.gitignore）。
  - INDEX整備（クリック可能リンク＋1行サマリ）。
- CI初期デザインの合意
  - ランナーはGitHub Actions（ubuntu-latest）。PR/手動トリガのみ、Pushは無効化。
  - オフラインE2E（SKIP）で配線チェックを優先し、段階的に可視化/ゲート強化へ。

## 2025-08-18 以前（まとめ）

- リポジトリ作成と初期ドキュメントの骨子作成。internal_docsは非公開・docsへ再編方針を決定。
