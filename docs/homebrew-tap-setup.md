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
2. GitHub App name: `tqer39-gha-runner`（GitHub 全体で一意。衝突したら suffix を付ける）
3. Homepage URL: `https://github.com/tqer39`
4. Webhook: **Active のチェックを外す**（このユースケースでは不要）
5. Repository permissions:
   - **Contents: Read and write**（formula push のため）
   - **Metadata: Read-only**（API 前提）
   - 他はすべて `No access`
6. Where can this GitHub App be installed?: **Only on this account**
7. "Create GitHub App" を押す
8. 作成後の設定画面で:
   - **Client ID**（`Iv23li...` 形式の文字列）をメモ（Step 4 で使用）
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
# Client ID を登録（Step 2 でメモした Iv23li... 文字列）
gh secret set GHA_APP_CLIENT_ID --repo tqer39/ccw-cli
# プロンプトに Client ID を貼り付け

# Private key を登録（Step 2 でダウンロードした .pem ファイルの中身ごと）
gh secret set GHA_APP_PRIVATE_KEY --repo tqer39/ccw-cli < /path/to/downloaded.pem
```

`gh secret list --repo tqer39/ccw-cli` に以下 2 つが出れば OK:

- `GHA_APP_CLIENT_ID`
- `GHA_APP_PRIVATE_KEY`

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

問題なければリリース実行に進める（次節）。

## 6. リリース実行（workflow_dispatch）

リリースは **GitHub Actions の workflow_dispatch から** 実行する。ローカルから tag push はしない:

1. <https://github.com/tqer39/ccw-cli/actions/workflows/release.yml> を開く
2. "Run workflow" を押す
3. `version` 入力に `vX.Y.Z` (例: `v0.1.0`) を入れて実行
4. workflow 内で以下が自動実行される:
   - `version` 形式チェック（`^v\d+\.\d+\.\d+(-.+)?$`）
   - `gh release create --draft --generate-notes` で **GitHub Release と tag を同時に作成**（REST API 経由、tag 作成は GitHub 側）
   - goreleaser が既存 draft release に 4 archive + checksums を append
   - homebrew-tap に `Formula/ccw.rb` を push
   - `gh release edit --draft=false` で release を publish

`gh workflow run release.yml -f version=v0.1.0 --repo tqer39/ccw-cli` でも同等。

リリース後の確認:

```bash
brew tap tqer39/tap
brew install tqer39/tap/ccw
ccw -v  # vX.Y.Z
```

## 7. トラブルシューティング

- **`Not Accessible by integration`**: App が tap リポにインストールされていない（Step 3 を再確認）
- **`Bad credentials`**: `GHA_APP_PRIVATE_KEY` の改行が壊れている。`.pem` を `gh secret set ... < file` でリダイレクト登録すれば LF が維持される
- **`App not found`**: `GHA_APP_CLIENT_ID` の値誤り。App 設定画面上部に表示される Client ID (`Iv23li...`) を使う
- **tap リポの `Formula/` が無い**: Step 1 の `.gitkeep` push を先に済ませる
- **422 `already_exists` on asset upload**: 同一 tag の Release が既にある。`gh release delete vX.Y.Z --cleanup-tag` で削除してから再 dispatch
- **workflow が「tag already exists」で失敗**: 手動で先に tag を push していないか確認。dispatch は新規 tag を作るので事前 push 不要
