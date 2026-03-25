# wall-vault ユーザーマニュアル
*(最終更新: 2026-03-20 — v0.1.15)*

---

## 目次

1. [wall-vaultとは？](#wall-vaultとは)
2. [インストール](#インストール)
3. [はじめての起動（setupウィザード）](#はじめての起動)
4. [APIキーの登録](#apiキーの登録)
5. [プロキシの使い方](#プロキシの使い方)
6. [キー金庫ダッシュボード](#キー金庫ダッシュボード)
7. [分散モード（マルチボット）](#分散モードマルチボット)
8. [自動起動の設定](#自動起動の設定)
9. [Doctorコマンド](#doctorコマンド)
10. [環境変数リファレンス](#環境変数リファレンス)
11. [トラブルシューティング](#トラブルシューティング)

---

## wall-vaultとは？

**wall-vault = OpenClaw向けのAI代理人（プロキシ） + APIキー金庫**

AIサービスを利用するには **APIキー**（＝「デジタル入場証」）が必要です。このAPIキーは1日に使える回数が決まっており、管理を誤ると漏洩するリスクもあります。

wall-vaultはこれらの入場証を安全な金庫に保管し、OpenClawとAIサービスの間で**代理人（プロキシ）**として動作します。つまり、OpenClawはwall-vaultにだけ接続すれば良く、残りの複雑な処理はwall-vaultがすべて引き受けてくれます。

wall-vaultが解決してくれる問題：

- **APIキーの自動ローテーション**: あるキーの使用量が上限に達したり一時的にブロックされた（クールダウン）場合、静かに次のキーへ切り替えます。OpenClawは中断なく動き続けます。
- **サービスの自動フォールバック**: Googleが応答しなければOpenRouterへ、それも駄目なら自分のPCにインストールしたOllama（ローカルAI）へ自動的に切り替わります。セッションが途切れることはありません。
- **リアルタイム同期（SSE）**: 金庫ダッシュボードでモデルを変更すると、1〜3秒以内にOpenClawの画面へ反映されます。SSE（Server-Sent Events）とは、サーバーが変化をリアルタイムでクライアントに送り届ける技術です。
- **リアルタイム通知**: キーの枯渇やサービス障害といったイベントが、OpenClawのTUI（ターミナル画面）下部にすぐ表示されます。

> 💡 **Claude Code、Cursor、VS Code**からも接続して使えますが、wall-vaultの本来の目的はOpenClawと一緒に使うことです。

```
OpenClaw（TUIターミナル画面）
        │
        ▼
  wall-vault プロキシ (:56244)   ← キー管理、ルーティング、フォールバック、イベント
        │
        ├─ Google Gemini API
        ├─ OpenRouter API（340以上のモデル）
        └─ Ollama（自分のPC、最後の砦）
```

---

## インストール

### Linux / macOS

ターミナルを開いて、以下のコマンドをそのまま貼り付けてください。

```bash
# Linux（一般PC、サーバー — amd64）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon（M1/M2/M3 Mac）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — ファイルをインターネットからダウンロードします。
- `chmod +x` — ダウンロードしたファイルを「実行可能」な状態にします。この手順を省略すると「権限がありません」というエラーが発生します。

### Windows

PowerShell（管理者権限）を開き、以下のコマンドを実行してください。

```powershell
# ダウンロード
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# PATHに追加（PowerShell再起動後に有効になります）
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATHとは？** コンピューターがコマンドを探しに行くフォルダのリストです。PATHに追加することで、どのフォルダからでも `wall-vault` と入力して実行できるようになります。

### ソースからビルドする（開発者向け）

Go言語の開発環境がインストールされている場合のみ対象です。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault（バージョン: v0.1.6.YYYYMMDD.HHmmss）
make install     # ~/.local/bin/wall-vault
```

> 💡 **ビルドタイムスタンプ版**: `make build` でビルドすると、バージョンが `v0.1.6.20260314.231308` のように日付と時刻を含む形式で自動生成されます。`go build ./...` で直接ビルドするとバージョンは `"dev"` とだけ表示されます。

---

## はじめての起動

### setupウィザードを実行する

インストール後、最初は必ず以下のコマンドで**設定ウィザード**を実行してください。ウィザードが必要な項目を一つずつ質問しながら案内してくれます。

```bash
wall-vault setup
```

ウィザードの進行ステップは以下の通りです：

```
1. 言語選択（日本語を含む10言語）
2. テーマ選択（light / dark / gold / cherry / ocean）
3. 運用モード — 一人で使う（standalone）か、複数台で共有する（distributed）かを選択
4. ボット名の入力 — ダッシュボードに表示される名前
5. ポート設定 — デフォルト：プロキシ 56244、金庫 56243（変更不要ならそのままEnter）
6. AIサービス選択 — Google / OpenRouter / Ollama から使うサービスを選択
7. ツールセキュリティフィルターの設定
8. 管理者トークンの設定 — ダッシュボードの管理機能をロックするパスワード。自動生成も可能
9. APIキー暗号化パスワードの設定 — キーをより安全に保存したい場合（任意）
10. 設定ファイルの保存先
```

> ⚠️ **管理者トークンは必ず覚えておいてください。** あとでダッシュボードからキーを追加したり設定を変更したりする際に必要です。忘れてしまった場合は設定ファイルを直接編集する必要があります。

ウィザードが完了すると、`wall-vault.yaml` 設定ファイルが自動的に生成されます。

### 起動する

```bash
wall-vault start
```

以下の2つのサーバーが同時に起動します：

- **プロキシ** (`http://localhost:56244`) — OpenClawとAIサービスをつなぐ代理人
- **キー金庫** (`http://localhost:56243`) — APIキー管理とWebダッシュボード

ブラウザで `http://localhost:56243` を開くと、ダッシュボードをすぐに確認できます。

---

## APIキーの登録

APIキーを登録する方法は4通りあります。**はじめての方には方法1（環境変数）をおすすめします。**

### 方法1：環境変数（推奨 — 最も簡単）

環境変数とは、プログラムが起動時に読み込む**あらかじめ設定しておいた値**のことです。ターミナルで以下のように入力します。

```bash
# Google Gemini キーを登録
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouter キーを登録
export WV_KEY_OPENROUTER=sk-or-v1-...

# 登録後に起動
wall-vault start
```

複数のキーを持っている場合はカンマ（,）でつなげてください。wall-vaultがキーを順番に自動で使い回します（ラウンドロビン）：

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **ヒント**: `export` コマンドは現在のターミナルセッションにのみ適用されます。PCを再起動しても有効にしたい場合は、`~/.bashrc` または `~/.zshrc` ファイルに上記の行を追記してください。

### 方法2：ダッシュボードUI（マウスで操作）

1. ブラウザで `http://localhost:56243` にアクセス
2. 上部の **🔑 APIキー** カードで `[+ 追加]` ボタンをクリック
3. サービス種別、キーの値、ラベル（メモ用の名前）、1日の上限を入力して保存

### 方法3：REST API（自動化・スクリプト用）

REST APIとは、プログラム同士がHTTPでデータをやり取りする方式です。スクリプトで自動登録する際に便利です。

```bash
curl -X POST http://localhost:56243/admin/keys \
  -H "Authorization: Bearer 管理者トークン" \
  -H "Content-Type: application/json" \
  -d '{
    "service": "google",
    "key": "AIzaSy...",
    "label": "メインキー",
    "daily_limit": 1000
  }'
```

### 方法4：proxyフラグ（ちょっと試したいとき）

正式に登録せず、一時的にキーを設定してテストしたいときに使います。プログラムを終了すると消えます。

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## プロキシの使い方

### OpenClawから使う（主な用途）

OpenClawがwall-vaultを経由してAIサービスに接続するように設定する方法です。

`~/.openclaw/openclaw.json` ファイルを開いて、以下の内容を追加してください：

```json5
// ~/.openclaw/openclaw.json
{
  models: {
    providers: {
      "wall-vault": {
        baseUrl: "http://localhost:56244/v1",
        apiKey: "your-agent-token",   // vault エージェントトークン
        api: "openai-completions",
        models: [
          { id: "wall-vault/gemini-2.5-flash" },
          { id: "wall-vault/gemini-2.5-pro" },
          { id: "wall-vault/hunter-alpha" },    // 無料1Mコンテキスト
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **もっと簡単な方法**: ダッシュボードのエージェントカードにある **🦞 OpenClaw設定をコピー** ボタンを押すと、トークンとアドレスがすでに入力済みのスニペットがクリップボードにコピーされます。あとは貼り付けるだけです。

**モデル名の先頭についている `wall-vault/` はどこへつながるの？**

モデル名を見て、wall-vaultがどのAIサービスへリクエストを送るかを自動的に判断します：

| モデルの形式 | つながるサービス |
|------------|--------------|
| `wall-vault/gemini-*` | Google Gemini へ直接接続 |
| `wall-vault/gpt-*`、`wall-vault/o3`、`wall-vault/o4*` | OpenAI へ直接接続 |
| `wall-vault/claude-*` | OpenRouter経由でAnthropicへ接続 |
| `wall-vault/hunter-alpha`、`wall-vault/healer-alpha` | OpenRouter（無料100万トークンコンテキスト） |
| `wall-vault/kimi-*`、`wall-vault/glm-*`、`wall-vault/deepseek-*` | OpenRouter へ接続 |
| `google/モデル名`、`openai/モデル名`、`anthropic/モデル名` など | 該当サービスへ直接接続 |
| `custom/google/モデル名`、`custom/openai/モデル名` など | `custom/` 部分を除いて再ルーティング |
| `モデル名:cloud` | `:cloud` 部分を除いてOpenRouterへ接続 |

> 💡 **コンテキスト（context）とは？** AIが一度に記憶できる会話の量のことです。1M（100万トークン）あれば、非常に長い会話や長い文書も一度に処理できます。

### Gemini API形式で直接接続する（既存ツールとの互換性）

Google Gemini APIを直接使っていたツールがある場合、アドレスをwall-vaultに変えるだけです：

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

またはURLを直接指定するツールの場合：

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### OpenAI SDK（Python）から使う

PythonでAIを活用するコードからもwall-vaultに接続できます。`base_url` を変えるだけです：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # APIキーはwall-vaultが管理してくれます
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # provider/model 形式で入力
    messages=[{"role": "user", "content": "こんにちは"}]
)
```

### 実行中にモデルを変更する

wall-vaultが稼働中の状態で使用するAIモデルを変更するには：

```bash
# プロキシに直接リクエストしてモデルを変更
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 分散モード（マルチボット）では金庫サーバーから変更 → SSEで即時反映
curl -X PUT http://localhost:56243/admin/clients/自分のボットID \
  -H "Authorization: Bearer 管理者トークン" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### 利用可能なモデル一覧を確認する

```bash
# 全リストを表示
curl http://localhost:56244/api/models | python3 -m json.tool

# Googleモデルだけ表示
curl "http://localhost:56244/api/models?service=google"

# 名前で検索（例：「claude」を含むモデル）
curl "http://localhost:56244/api/models?q=claude"
```

**サービス別の主なモデル一覧：**

| サービス | 主なモデル |
|---------|----------|
| Google | gemini-2.5-pro、gemini-2.5-flash、gemini-2.5-flash-8b、gemini-2.0-flash |
| OpenAI | gpt-4o、gpt-4o-mini、o3、o1、o1-mini |
| OpenRouter | 346以上（Hunter Alpha 1Mコンテキスト無料、DeepSeek R1/V3、Qwen 2.5 など） |
| Ollama | 自分のPCにインストールされたローカルサーバーを自動検出 |

---

## キー金庫ダッシュボード

ブラウザで `http://localhost:56243` にアクセスするとダッシュボードが開きます。

**画面の構成：**
- **上部固定バー（topbar）**: ロゴ、言語・テーマ選択、SSE接続状態の表示
- **カードグリッド**: エージェント・サービス・APIキーのカードがタイル形式で並んでいます

### APIキーカード

登録済みのAPIキーを一覧で管理できるカードです。

- サービスごとに区分してキーの一覧を表示します。
- `today_usage`: 本日正常に処理されたトークン（AIが読み書きした文字数）の数
- `today_attempts`: 本日の合計呼び出し回数（成功＋失敗を含む）
- `[+ 追加]` ボタンで新しいキーを登録し、`✕` でキーを削除します。

> 💡 **トークン（token）とは？** AIがテキストを処理する際に使う単位です。おおよそ英単語1つ、または日本語の1〜2文字程度に相当します。API料金は通常このトークン数に基づいて計算されます。

### エージェントカード

wall-vaultプロキシに接続しているボット（エージェント）の状態を表示するカードです。

**接続状態は4段階で表示されます：**

| 表示 | 状態 | 意味 |
|------|------|------|
| 🟢 | 実行中 | プロキシが正常に動作している |
| 🟡 | 遅延 | 応答は来ているが遅い |
| 🔴 | オフライン | プロキシが応答しない |
| ⚫ | 未接続・非アクティブ | プロキシが金庫に接続したことがない、または無効化されている |

**エージェントカード下部のボタンについて：**

エージェントを登録するとき**エージェントの種別**を指定すると、その種別に応じた便利ボタンが自動的に表示されます。

---

#### 🔘 設定コピーボタン — 接続設定を自動で作成してくれます

ボタンをクリックすると、そのエージェントのトークン、プロキシアドレス、モデル情報があらかじめ入力された設定スニペットがクリップボードにコピーされます。コピーした内容を下の表の場所に貼り付けるだけで接続設定が完了します。

| ボタン | エージェント種別 | 貼り付ける場所 |
|--------|---------------|-------------|
| 🦞 OpenClaw設定をコピー | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw設定をコピー | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code設定をコピー | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor設定をコピー | `cursor` | Cursor → Settings → AI |
| 💻 VSCode設定をコピー | `vscode` | `~/.continue/config.json` |

**例 — Claude Codeタイプの場合、こんな内容がコピーされます：**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.1.20:56244/v1",
  "apiKey": "このエージェントのトークン"
}
```

**例 — VSCode（Continue）タイプの場合：**

```json
// ~/.continue/config.json
{
  "models": [{
    "title": "wall-vault proxy",
    "provider": "openai",
    "model": "gemini-2.5-flash",
    "apiBase": "http://192.168.1.20:56244/v1",
    "apiKey": "このエージェントのトークン"
  }]
}
```

**例 — Cursorタイプの場合：**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : このエージェントのトークン

// または環境変数：
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=このエージェントのトークン
```

> ⚠️ **クリップボードへのコピーができない場合**: ブラウザのセキュリティポリシーによってコピーがブロックされることがあります。ポップアップでテキストボックスが開いたら、Ctrl+Aで全選択してからCtrl+Cでコピーしてください。

---

#### ⚡ 自動適用ボタン — ワンクリックで設定完了

エージェント種別が `cline`、`claude-code`、`openclaw`、`nanoclaw` の場合、エージェントカードに **⚡ 設定を適用** ボタンが表示されます。このボタンを押すと、該当エージェントのローカル設定ファイルが自動的に更新されます。

| ボタン | エージェント種別 | 適用対象ファイル |
|--------|---------------|----------------|
| ⚡ Cline設定を適用 | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Claude Code設定を適用 | `claude-code` | `~/.claude/settings.json` |
| ⚡ OpenClaw設定を適用 | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ NanoClaw設定を適用 | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ このボタンは **localhost:56244**（ローカルプロキシ）にリクエストを送信します。そのマシンでプロキシが実行中でなければ動作しません。

---

#### 🔀 ドラッグ＆ドロップによるカード並べ替え (v0.1.17)

ダッシュボードのエージェントカードを**ドラッグ**して、好きな順番に並べ替えることができます。

1. エージェントカードをマウスでつかんでドラッグします
2. 目的の位置のカードの上にドロップすると、順番が入れ替わります
3. 変更された順番は**即座にサーバーに保存**され、ページを更新しても維持されます

> 💡 タッチデバイス（モバイル／タブレット）にはまだ対応していません。デスクトップブラウザでご利用ください。

---

#### 🔄 双方向モデル同期 (v0.1.16)

ボールトダッシュボードでエージェントのモデルを変更すると、該当エージェントのローカル設定が自動的に更新されます。

**Clineの場合：**
- ボールトでモデルを変更 → SSEイベント → プロキシが `globalState.json` のモデルフィールドを更新
- 更新対象：`actModeOpenAiModelId`、`planModeOpenAiModelId`、`openAiModelId`
- `openAiBaseUrl` とAPIキーは変更されません
- **VS Codeのリロード（`Ctrl+Alt+R` または `Ctrl+Shift+P` → `Developer: Reload Window`）が必要です**
  - Clineは実行中に設定ファイルを再読み込みしないため

**Claude Codeの場合：**
- ボールトでモデルを変更 → SSEイベント → プロキシが `settings.json` の `model` フィールドを更新
- WSLとWindows両方のパスを自動探索（`~/.claude/`、`/mnt/c/Users/*/.claude/`）

**逆方向（エージェント → ボールト）：**
- エージェント（Cline、Claude Codeなど）がプロキシにリクエストを送ると、プロキシがハートビートに該当クライアントのサービス・モデル情報を含めます
- ボールトダッシュボードのエージェントカードに、現在使用中のサービス/モデルがリアルタイムで表示されます

> 💡 **ポイント**: プロキシはリクエストのAuthorizationトークンでエージェントを識別し、ボールトで設定されたサービス/モデルへ自動ルーティングします。ClineやClaude Codeが異なるモデル名を送信しても、プロキシがボールトの設定でオーバーライドします。

---

### VS CodeでClineを使う — 詳細ガイド

#### ステップ1：Clineのインストール

VS Code拡張機能マーケットプレイスで **Cline**（ID: `saoudrizwan.claude-dev`）をインストールします。

#### ステップ2：ボールトにエージェントを登録

1. ボールトダッシュボード（`http://ボールトIP:56243`）を開きます
2. **エージェント** セクションで **+ 追加** をクリック
3. 以下のように入力します：

| フィールド | 値 | 説明 |
|-----------|-----|------|
| ID | `my_cline` | 一意の識別子（英数字、スペースなし） |
| 名前 | `My Cline` | ダッシュボードに表示される名前 |
| エージェント種別 | `cline` | ← 必ず `cline` を選択 |
| サービス | 使用するサービスを選択（例：`google`） | |
| モデル | 使用するモデルを入力（例：`gemini-2.5-flash`） | |

4. **保存** をクリックするとトークンが自動生成されます

#### ステップ3：Clineに接続する

**方法A — 自動適用（推奨）**

1. そのマシンでwall-vault **プロキシ** が実行中であることを確認（`localhost:56244`）
2. ダッシュボードのエージェントカードで **⚡ Cline設定を適用** ボタンをクリック
3. 「設定の適用が完了しました！」という通知が表示されれば成功
4. VS Codeをリロードします（`Ctrl+Alt+R`）

**方法B — 手動設定**

Clineサイドバーの設定（⚙️）を開き：
- **API Provider**: `OpenAI Compatible`
- **Base URL**: `http://プロキシアドレス:56244/v1`
  - 同じマシンなら `http://localhost:56244/v1`
  - Miniサーバーなど別のマシンなら `http://192.168.1.20:56244/v1`
- **API Key**: ボールトで発行されたトークン（エージェントカードからコピー）
- **Model ID**: ボールトで設定したモデル（例：`gemini-2.5-flash`）

#### ステップ4：確認

Clineのチャットウィンドウで任意のメッセージを送信してみます。正常であれば：
- ボールトダッシュボードの該当エージェントカードに **緑色の点（● 実行中）** が表示されます
- カードに現在のサービス/モデルが表示されます（例：`google / gemini-2.5-flash`）

#### モデルの変更

Clineのモデルを変更したい場合は **ボールトダッシュボード** で変更してください：

1. エージェントカードのサービス/モデルドロップダウンを変更
2. **適用** をクリック
3. VS Codeをリロード（`Ctrl+Alt+R`）— Clineフッターのモデル名が更新されます
4. 次のリクエストから新しいモデルが使用されます

> 💡 実際にはプロキシがClineのリクエストをトークンで識別し、ボールトの設定に基づくモデルへルーティングします。VS Codeをリロードしなくても **実際に使用されるモデルは即座に切り替わります** — リロードはCline UIのモデル表示を更新するためのものです。

#### 接続切断の検知

VS Codeを閉じると、ボールトダッシュボードでは約 **2～3分** 後にエージェントカードが黄色（遅延）に、**5分** 後に赤色（オフライン）に変わります。

#### トラブルシューティング

| 症状 | 原因 | 解決方法 |
|------|------|---------|
| Clineで「接続失敗」エラー | プロキシ未起動またはアドレスの誤り | `curl http://localhost:56244/health` でプロキシを確認 |
| ボールトで緑色の点が表示されない | APIキー（トークン）が設定されていない | **⚡ Cline設定を適用** ボタンを再度クリック |
| Clineフッターのモデルが変わらない | Clineが設定をキャッシュしている | VS Codeリロード（`Ctrl+Alt+R`） |
| 間違ったモデル名が表示される | 以前のバグ（v0.1.16で修正済み） | プロキシをv0.1.16以上に更新 |

---

#### 🟣 デプロイコマンドコピーボタン — 新しいマシンにインストールするときに使います

新しいコンピューターにwall-vaultプロキシを初めてインストールして金庫に接続する際に使います。ボタンをクリックするとインストールスクリプト全体がコピーされます。新しいコンピューターのターミナルに貼り付けて実行すると、以下が一括で処理されます：

1. wall-vaultバイナリのインストール（すでにインストール済みの場合はスキップ）
2. systemdユーザーサービスの自動登録
3. サービスの起動と金庫への自動接続

> 💡 スクリプトの中にこのエージェントのトークンと金庫サーバーのアドレスがあらかじめ入力されているので、貼り付け後は追加の修正なしにそのまま実行できます。

---

### サービスカード

使用するAIサービスをオン・オフしたり設定したりするカードです。

- サービスごとの有効・無効トグルスイッチ
- ローカルAIサーバー（自分のPCで動かしているOllama、LM Studio、vLLMなど）のアドレスを入力すると、利用可能なモデルを自動で見つけてくれます。
- **ローカルサービス接続状態の表示**: サービス名の横の ● 点が**緑色**なら接続済み、**灰色**なら未接続
- **チェックボックスの自動同期**: ページを開いたとき、ローカルサービス（Ollamaなど）が起動中であれば自動的にチェック状態になります。

> 💡 **ローカルサービスが別のコンピューターで動いている場合**: サービスURLの入力欄にそのコンピューターのIPを入力してください。例：`http://192.168.1.20:11434`（Ollama）、`http://192.168.1.20:1234`（LM Studio）

### 管理者トークンの入力

ダッシュボードでキーの追加・削除といった重要な操作をしようとすると、管理者トークンの入力ポップアップが表示されます。setupウィザードで設定したトークンを入力してください。一度入力するとブラウザを閉じるまで保持されます。

> ⚠️ **15分以内に認証が10回失敗すると、そのIPが一時的にブロックされます。** トークンを忘れた場合は `wall-vault.yaml` ファイルの `admin_token` 項目を確認してください。

---

## 分散モード（マルチボット）

複数のコンピューターでOpenClawを同時に運用する際に、**一つのキー金庫を共有**する構成です。キー管理を一か所で行えるので便利です。

### 構成例

```
[キー金庫サーバー]
  wall-vault vault    （キー金庫 :56243、ダッシュボード）

[WSL アルファ]            [Raspberry Pi ガンマ]    [Mac Mini ローカル]
  wall-vault proxy      wall-vault proxy          wall-vault proxy
  openclaw TUI          openclaw TUI              openclaw TUI
  ↕ SSE同期             ↕ SSE同期                 ↕ SSE同期
```

すべてのボットが中央の金庫サーバーを参照しているため、金庫でモデルを変更したりキーを追加したりすると、すべてのボットに即時反映されます。

### ステップ1：キー金庫サーバーを起動する

金庫サーバーとして使うコンピューターで実行します：

```bash
wall-vault vault
```

### ステップ2：各ボット（クライアント）を登録する

金庫サーバーに接続する各ボットの情報をあらかじめ登録しておきます：

```bash
curl -X POST http://localhost:56243/admin/clients \
  -H "Authorization: Bearer 管理者トークン" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "botA",
    "name": "ボットA",
    "token": "bota-secret",
    "default_service": "google",
    "default_model": "gemini-2.5-flash"
  }'
```

### ステップ3：各ボットのコンピューターでプロキシを起動する

ボットがインストールされた各コンピューターで、金庫サーバーのアドレスとトークンを指定してプロキシを実行します：

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **`192.168.x.x`** の部分は、金庫サーバーコンピューターの実際のローカルIPアドレスに置き換えてください。ルーターの設定画面または `ip addr` コマンドで確認できます。

---

## 自動起動の設定

コンピューターを再起動するたびに手動でwall-vaultを起動するのが手間なら、システムサービスとして登録しておきましょう。一度登録すれば、起動時に自動的に開始されます。

### Linux — systemd（ほとんどのLinux）

systemdは、Linuxでプログラムを自動的に起動・管理するシステムです：

```bash
wall-vault doctor deploy
systemctl --user daemon-reload
systemctl --user enable --now wall-vault
```

ログの確認：

```bash
journalctl --user -u wall-vault -f
```

### macOS — launchd

macOSでプログラムの自動実行を担当するシステムです：

```bash
wall-vault doctor deploy launchd
launchctl load ~/Library/LaunchAgents/com.wall-vault.plist
```

### Windows — NSSM

1. [nssm.cc](https://nssm.cc/download) からNSSMをダウンロードしてPATHに追加します。
2. 管理者権限のPowerShellで：

```powershell
wall-vault doctor deploy windows
```

---

## Doctorコマンド

`doctor` コマンドは、wall-vaultが正しく設定されているかどうかを**自己診断して修復してくれるツール**です。

```bash
wall-vault doctor check   # 現在の状態を診断（読み取りのみ、何も変更しない）
wall-vault doctor fix     # 問題を自動的に修復
wall-vault doctor all     # 診断＋自動修復を一括実行
```

> 💡 何かおかしいと感じたら、まず `wall-vault doctor all` を実行してみてください。多くの問題を自動的に解決してくれます。

---

## 環境変数リファレンス

環境変数とは、プログラムに設定値を渡す方法です。`export 変数名=値` の形式でターミナルに入力するか、自動起動サービスファイルに記載しておくと常に適用されます。

| 変数 | 説明 | 例 |
|------|------|----|
| `WV_LANG` | ダッシュボードの言語 | `ko`、`en`、`ja` |
| `WV_THEME` | ダッシュボードのテーマ | `light`、`dark`、`gold` |
| `WV_KEY_GOOGLE` | Google APIキー（カンマ区切りで複数指定可） | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter APIキー | `sk-or-v1-...` |
| `WV_VAULT_URL` | 分散モードでの金庫サーバーアドレス | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | クライアント（ボット）認証トークン | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 管理者トークン | `admin-token-here` |
| `WV_MASTER_PASS` | APIキー暗号化パスワード | `my-password` |
| `WV_AVATAR` | アバター画像ファイルのパス（`~/.openclaw/` からの相対パス） | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollamaローカルサーバーのアドレス | `http://192.168.x.x:11434` |

---

## トラブルシューティング

### プロキシが起動しない場合

ポートがすでに別のプログラムに使われていることが多いです。

```bash
ss -tlnp | grep 56244   # 56244ポートを誰が使っているか確認
wall-vault proxy --port 8080   # 別のポート番号で起動
```

### APIキーエラーが発生する場合（429、402、401、403、582）

| エラーコード | 意味 | 対処方法 |
|------------|------|---------|
| **429** | リクエストが多すぎる（使用量超過） | しばらく待つか、別のキーを追加する |
| **402** | 支払いが必要またはクレジット不足 | 該当サービスでクレジットをチャージする |
| **401 / 403** | キーが間違っているか権限がない | キーの値を再確認して再登録する |
| **582** | ゲートウェイ過負荷（クールダウン5分） | 5分後に自動解除される |

```bash
# 登録済みキーの一覧と状態を確認
curl -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys

# キーの使用量カウンターをリセット
curl -X POST -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys/reset
```

### エージェントが「未接続」と表示される場合

「未接続」とは、プロキシプロセスが金庫にシグナル（heartbeat）を送っていない状態です。**設定が保存されていないという意味ではありません。** プロキシが金庫サーバーのアドレスとトークンを知った状態で実行されて初めて、接続済みの状態に変わります。

```bash
# 金庫サーバーアドレス、トークン、クライアントIDを指定してプロキシを起動
WV_VAULT_URL=http://金庫サーバーアドレス:56243 \
WV_VAULT_TOKEN=クライアントトークン \
WV_VAULT_CLIENT_ID=クライアントID \
wall-vault proxy
```

接続に成功すると、約20秒以内にダッシュボードで 🟢 実行中 に変わります。

### Ollamaに接続できない場合

Ollamaは自分のPC上でAIを直接実行するプログラムです。まずOllamaが起動しているか確認してください。

```bash
curl http://localhost:11434/api/tags   # モデル一覧が表示されれば正常
export OLLAMA_URL=http://192.168.x.x:11434   # 別のコンピューターで動いている場合
```

> ⚠️ Ollamaが応答しない場合は、`ollama serve` コマンドでまずOllamaを起動してください。

> ⚠️ **大型モデルは応答が遅くなります**: `qwen3.5:35b`、`deepseek-r1` のような大きなモデルは、応答の生成まで数分かかることがあります。応答がないように見えても正常に処理中の場合がありますので、しばらくお待ちください。

---

*より詳しいAPI情報は [API.md](API.md) をご参照ください。*
