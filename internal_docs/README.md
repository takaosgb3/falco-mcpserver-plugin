# internal_docs 概要（MCP Server Plugin）

- 本フォルダは、MCP Server向けFalcoプラグインの要件/設計/運用と、nginxプラグインから転用した知見の整理を目的とします。
- 索引: `INDEX.md` を起点にドキュメントを辿ってください。
- 変更方針: 小さく、可逆に。CI優先・ドキュメント更新を同時に行います。

構成:
- 要件: `requirements/`
- 知見: `knowledge/nginx/`（nginxプラグインの再利用ドキュメント）
- 横断: `DEVELOPMENT_EFFICIENCY_PLAN.md`（将来の効率化）

注意事項:
- `internal_docs/falco-nginx-plugin-docs/` は参照専用の原本集であり、リモートにはプッシュしません（`.gitignore` 済）。
- 公開/共有が必要な内容は、要点を抽出し `knowledge/nginx/` に再整理して配置してください。
