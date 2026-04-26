# cspell 設定

このディレクトリは pre-commit / CI で動作する `cspell`（綴りチェッカ）の設定一式です。

## ファイル

- `cspell.json` — 本体設定。除外パス（バイナリ・superpowers 関連・ノイズ源）を `ignorePaths` で集約管理しています。
- `project-words.txt` — プロジェクト固有語の辞書。1 行 1 語。

## 偽陽性が出たとき

cspell が正当な語をエラー扱いした場合、その語を `project-words.txt` に 1 行 1 語で追記してコミットしてください。

```bash
echo "yourword" >> .cspell/project-words.txt
git add .cspell/project-words.txt
git commit -m "chore(cspell): add yourword to dictionary"
```

## ローカル全件チェック

```bash
npm exec -y --package=cspell -- cspell lint \
  --no-progress --config .cspell/cspell.json \
  "**/*"
```
