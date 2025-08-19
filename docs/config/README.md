# Config Guide (Public)

- 目的: 運用時に利用する許可リスト/しきい値/レッドアクト方針の例を示す（参考実装）。
- 注意: 現時点のルールは固定値ベース。将来的にプラグイン/生成で動的化を検討。

## Example Values

- サンプル: `docs/config/EXAMPLE_VALUES.yaml`
- 意味:
  - `allowlist.hosts`: 許可する MCP サーバホスト。
  - `thresholds.*`: 警告のしきい値（バイト/回数）。
  - `redaction.*`: 認証情報やIDのマスキング/ハッシュ。

## 運用のヒント

- 環境別に上書きファイル（例: `values.local.yaml`）を持ち、CI/本番で切替。
- プラグイン側で評価済みブール（例: `mcp.host_allowed`）を出力できるようにすると、ルールが簡潔に。
- 変更は小さく可逆に。ルール変更時はテスト（JSONアサーション）も併せて更新。

