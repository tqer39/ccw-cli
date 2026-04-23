# 2026-04-24 — `-s` superpowers プラグイン自動インストール

## 背景

現状の `internal/superpowers/detect.go::EnsureInstalled`：

- **interactive + プラグイン未導入**: warning → インストールコマンド提示 → `Run now? [y/N]` プロンプト → y なら `claude plugin install claude-plugins-official/superpowers` 実行
- **non-interactive + プラグイン未導入**: warning だけ出して **error 終了**（インストールしない）

`-y` フラグ (`cli.Flags.AssumeYes`) は現状 `--clean-all` にしか波及しない。`ccw -s -y` や TTY でない環境では常にエラーで止まる。

## ゴール

`-y` 指定時 / 非対話時でも `-s` が使えるように auto-install を追加。

## 非ゴール

- 既存のインタラクティブ分岐の文言 / 挙動変更
- superpowers 以外のプラグインサポート
- `.gitignore` まわり（Part D 別扱い）

## 変更内容

### 挙動マトリクス（変更後）

| interactive | `-y` | プラグイン状態 | 挙動 |
|---|---|---|---|
| true | - | 導入済 | 何もしない（現状維持） |
| true | false | 未導入 | `Run now? [y/N]` プロンプト（現状維持） |
| true | true | 未導入 | **auto-install**（事前・結果メッセージ付き、プロンプトなし） |
| false | - | 導入済 | 何もしない（現状維持） |
| false | - | 未導入 | **auto-install**（事前・結果メッセージ付き、プロンプトなし） |

事前メッセージ:

```text
Installing superpowers plugin (claude plugin install claude-plugins-official/superpowers)…
```

結果メッセージ:

- 成功: `Installed superpowers plugin.`
- 失敗: 既存の `fmt.Errorf("plugin install failed: %w", err)` → `ui.Error` で表示して exit code 1

### コード変更箇所

- `internal/superpowers/detect.go::EnsureInstalled`
  - シグネチャに `assumeYes bool` を追加 → `EnsureInstalled(in, out, home, interactive, assumeYes)`
  - auto-install 条件を `!interactive || assumeYes` に拡張
  - auto-install 分岐は事前メッセージ → `installRunner()` → 成功時メッセージ
- `cmd/ccw/main.go::maybeSuperpowers`
  - `interactive` に加え `flags.AssumeYes` を受け取り、`EnsureInstalled` に渡す
- `cmd/ccw/main.go::run`
  - `maybeSuperpowers(flags.Superpowers, mainRepo, interactive, flags.AssumeYes)` に呼び出し変更

### CLI ヘルプ

`-y` のヘルプ文言は現状 `skip confirmation prompt`。スコープが広がるので文言微修正：

- 現状: `skip confirmation prompt`
- 変更後: `skip confirmation prompts (--clean-all, -s plugin install)`

### README の扱い

PR-A で Features 行が「未導入なら入れるか確認」と書かれるため、Usage セクションの `ccw -s` 例も現状維持で齟齬は出ない。本 PR では README に手を入れない（ヘルプテキストの更新のみで十分）。

※ PR-A と本 PR の merge 順は任意。双方とも独立した編集箇所のため衝突しない。

## テスト

- `internal/superpowers/detect_test.go` を拡張：
  - interactive=true, assumeYes=false, 未導入 → プロンプトが出ること（既存）
  - interactive=true, assumeYes=true, 未導入 → プロンプトが出ず `installRunner()` が呼ばれる
  - interactive=false, 未導入 → プロンプトが出ず `installRunner()` が呼ばれる
  - 出力に事前メッセージ・結果メッセージが含まれる
  - `installRunner()` が error を返したら `EnsureInstalled` も error を返す

## 実装手順

1. `EnsureInstalled` 拡張 + 既存テストのシグネチャ更新
2. 新 path のテスト追加（赤で先に書く）
3. 実装を通す
4. `main.go` の呼び出し箇所更新
5. `cli/help.go` の `-y` 説明文更新
6. 手動確認: `ccw -s -y`（TTY あり） / `ccw -s < /dev/null`（non-interactive） をそれぞれ走らせて挙動確認

## リスク / 影響

- 挙動変更は「エラー終了 → auto-install」の拡張なので後方互換は壊さない（エラー終了を期待していたスクリプトがあれば影響するが現実的には考えにくい）
- `claude plugin install` が失敗するのは稀。network / auth / 破損 plugin cache の 3 ケース想定。失敗時はメッセージだけ出してプロセスを殺す（リカバーしない）

## PR スコープ

この spec は **PR-B** 単独用。
