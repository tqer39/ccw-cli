# Formula/ — homebrew-core 提出用 draft

このディレクトリの `ccw.rb` は **homebrew-core 提出を想定した draft** です。

- これは tap ではありません。`brew install ./Formula/ccw.rb` のような直接指定はサポートしていません
- 通常のインストールは `brew install tqer39/tap/ccw`（自動生成された pre-built バイナリ formula）を使ってください
- 提出 PR は notability 要件（stars / forks / watchers ≥ 30 など）達成後に別タスクで実施します
- 提出時には version / sha256 を最新 release に合わせて bump してから PR を出します

## 雛形を更新する手順（参考）

1. 最新 release タグの source tarball sha256 を取得:

   ```bash
   curl -sL https://github.com/tqer39/ccw-cli/archive/refs/tags/v<X.Y.Z>.tar.gz | shasum -a 256
   ```

2. `ccw.rb` の `url` / `sha256` を書き換える
3. `brew audit --new --formula ./Formula/ccw.rb`（ローカル Homebrew 環境がある場合）でドライ検証
