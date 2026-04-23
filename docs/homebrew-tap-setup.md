# Homebrew Tap セットアップ手順

Phase 4 の goreleaser 設定が機能するために、Phase 5（v0.1.0 タグ）を打つ前に
以下を済ませておくこと。すべて **ccw-cli リポ外** の作業。

## 1. tap リポジトリを作成

GitHub 上で `tqer39/homebrew-tap` を新規作成する（public / `main` branch / README 自動生成ありでよい）。

```bash
gh repo create tqer39/homebrew-tap --public --description "Homebrew tap for tqer39 tools" --add-readme --clone
cd homebrew-tap
mkdir -p Formula
touch Formula/.gitkeep
git add Formula/.gitkeep
git commit -m "init: Formula directory"
git push origin main
cd ..
```

tap リポには `Formula/` ディレクトリだけ用意しておけば goreleaser が `Formula/ccw.rb` を自動 push する。

## 2. fine-grained PAT を発行

1. <https://github.com/settings/tokens?type=beta> で "Generate new token" を選ぶ
2. 名前: `ccw-cli-homebrew-tap-push`
3. Resource owner: `tqer39`
4. Repository access: Only select repositories → `tqer39/homebrew-tap`
5. Repository permissions:
   - Contents: Read and write
   - Metadata: Read-only
6. Expiration: 90 days（Renovate の更新サイクルに合わせて回す）
7. 発行されたトークン値をコピー

## 3. ccw-cli の Actions secret に登録

```bash
gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo tqer39/ccw-cli
# プロンプトに PAT 値を貼り付け
```

`gh secret list --repo tqer39/ccw-cli` に `HOMEBREW_TAP_GITHUB_TOKEN` が出れば OK。

## 4. 動作確認（Phase 5 実施前）

Phase 4 がマージされた後、以下を実行して問題がないことを確認する:

```bash
# ccw-cli リポで
git fetch origin
git checkout main
git pull
make release-check   # "config is valid"
make release-snapshot # dist/ 配下に 4 archives + ccw.rb が出る
make release-clean
```

問題なければ Phase 5 のタグ付け（`v0.1.0`）に進める。

## 5. トラブルシューティング

- `HOMEBREW_TAP_GITHUB_TOKEN` が未設定で tag push した場合:
  - goreleaser-action の `publish` で 401 が出て Release 作成も止まる
  - secret を設定後、同タグを `git tag -d` → `git push --delete` → 再打ち直し
- tap リポの `Formula/` が無い場合:
  - goreleaser がディレクトリを自動作成しようとして書き込み権限エラーになることがある
  - Step 1 の `.gitkeep` push を先に済ませること
- PAT の repo scope が不足:
  - `Resource not accessible by integration` が出る
  - Step 2 の permissions を見直す（Contents: Read and write が必要）
