# actions/create-github-app-token: app-id → client-id 移行

## 背景

`release` ワークフローの run（[run 24938206359](https://github.com/tqer39/ccw-cli/actions/runs/24938206359)）で以下の deprecation 警告が annotations に表示されている。

```text
##[warning]Input 'app-id' has been deprecated with message: Use 'client-id' instead.
```

`actions/create-github-app-token@v3` 系では入力フィールド `app-id` が廃止予定となり、`client-id` の使用が推奨されている。Secret `GHA_APP_CLIENT_ID` には既に GitHub App の Client ID 値（`Iv23li...` 形式）が格納済みのため、ワークフロー側のフィールド名のみを仕様準拠に揃える。

## 目的

- `release` ワークフローの deprecation 警告を解消する
- 将来 `actions/create-github-app-token` の major バージョンで `app-id` が完全削除された際にリリースが壊れないようにする

## 変更対象

`.github/workflows/release.yml`（該当ステップ: `Mint Homebrew tap token`）

## 変更内容

```diff
       - name: Mint Homebrew tap token
         id: tap-token
         uses: actions/create-github-app-token@1b10c78c7865c340bc4f6099eb2f838309f1e8c3 # v3.1.1
         with:
-          app-id: ${{ secrets.GHA_APP_CLIENT_ID }}
+          client-id: ${{ secrets.GHA_APP_CLIENT_ID }}
           private-key: ${{ secrets.GHA_APP_PRIVATE_KEY }}
           owner: tqer39
           repositories: homebrew-tap
```

Secret 名（`GHA_APP_CLIENT_ID`）と Secret 値（Client ID 文字列）は変更しない。アクションの pin（`@1b10c78...` / v3.1.1）も変更しない。

## 前提条件

- Secret `GHA_APP_CLIENT_ID` の値は既に GitHub App の Client ID（`Iv23li...` 形式）であること（ユーザー確認済み）

## 検証方法

### 静的検証

- 既存の lefthook / pre-commit に組み込まれている YAML 系チェック（`yamllint`、`actionlint` 相当）を通す
- `pinact` で SHA pin が崩れていないことを確認する

### 動的検証

`release` ワークフローは `workflow_dispatch` でのみ起動するため、PR 単独ではトリガーされない。次のいずれかで確認する。

1. マージ後に最初のリリース機会（`gh workflow run release.yml -f version=vX.Y.Z`）で実行し、Mint Homebrew tap token ステップの annotations から該当 warning が消えていること、ステップが success すること、Homebrew tap への push が成功することを確認する
2. もしくはマージ前に同等の workflow を fork した検証用リポジトリで dry run する（任意）

## スコープ外

同 run で観測された以下は本変更の対象外として、別 issue / PR で扱う。

- goreleaser の `brews → homebrew_casks` deprecation（`brews is being phased out in favor of homebrew_casks`）

## リスク / ロールバック

- 影響は `Mint Homebrew tap token` ステップのみ。Secret 値が事前確認のとおり Client ID であれば、フィールド名差し替えだけで挙動は等価
- 万一トークン発行に失敗した場合は当該行を `app-id:` に戻す revert PR で即座に復旧可能（warning は再発するが機能は復旧）
