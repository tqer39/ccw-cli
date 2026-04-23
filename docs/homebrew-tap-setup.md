# Homebrew Tap セットアップ手順

Phase 4 の goreleaser 設定が機能するために、Phase 5（v0.1.0 タグ）を打つ前に
以下を済ませておくこと。すべて **ccw-cli リポ外** の作業。

GitHub Actions の `GITHUB_TOKEN` は実行中のリポジトリにしか書き込めない。
Homebrew tap は規約上別リポ (`tqer39/homebrew-tap`) なので、goreleaser が
formula を push するには「tap リポへの書き込み権限を持つ追加トークン」が必要。
本手順では **GitHub App** を使ってこのトークンを短命に発行する（PAT より安全、
期限レス、権限が細かい）。

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

## 2. GitHub App を作成

1. <https://github.com/settings/apps/new> を開く
2. GitHub App name: `tqer39-homebrew-tap-writer`（GitHub 全体で一意。衝突したら suffix を付ける）
3. Homepage URL: `https://github.com/tqer39/homebrew-tap`
4. Webhook: **Active のチェックを外す**（このユースケースでは不要）
5. Repository permissions:
   - **Contents: Read and write**（formula push のため）
   - **Metadata: Read-only**（API 前提）
   - 他はすべて `No access`
6. Where can this GitHub App be installed?: **Only on this account**
7. "Create GitHub App" を押す
8. 作成後の設定画面で:
   - **App ID** の数値をメモ（Step 4 で使用）
   - "Private keys" セクションで **"Generate a private key"** を押して `.pem` をダウンロード（Step 4 で使用）

## 3. App を tap リポにインストール

1. App 設定画面の左メニューから "Install App" を選ぶ
2. `tqer39` アカウントの "Install" を押す
3. "Only select repositories" → `tqer39/homebrew-tap` のみ選択
4. Install を押す

ccw-cli 側には **インストール不要**（App トークンは ccw-cli の workflow から
App ID と private key だけで発行できる）。

## 4. ccw-cli の Actions secret に登録

```bash
# App ID を登録（Step 2 でメモした数値）
gh secret set HOMEBREW_TAP_APP_ID --repo tqer39/ccw-cli
# プロンプトに App ID を貼り付け

# Private key を登録（Step 2 でダウンロードした .pem ファイルの中身ごと）
gh secret set HOMEBREW_TAP_APP_PRIVATE_KEY --repo tqer39/ccw-cli < /path/to/downloaded.pem
```

`gh secret list --repo tqer39/ccw-cli` に以下 2 つが出れば OK:

- `HOMEBREW_TAP_APP_ID`
- `HOMEBREW_TAP_APP_PRIVATE_KEY`

## 5. 動作確認（Phase 5 実施前）

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

## 6. トラブルシューティング

- **`Not Accessible by integration`**: App が tap リポにインストールされていない（Step 3 を再確認）
- **`Bad credentials`**: `HOMEBREW_TAP_APP_PRIVATE_KEY` の改行が壊れている。`.pem` を `gh secret set ... < file` でリダイレクト登録すれば LF が維持される
- **`App not found`**: `HOMEBREW_TAP_APP_ID` の値誤り。App 設定画面上部に表示される数値（先頭 6〜7 桁）を使う
- **tap リポの `Formula/` が無い**: Step 1 の `.gitkeep` push を先に済ませる
- **タグ再打ち直し**: secret 未設定で tag push してしまった場合、`git tag -d v0.1.0 && git push origin :refs/tags/v0.1.0` で削除して再 push
