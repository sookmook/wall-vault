# wall-vault ユーザーマニュアル
*(最終更新: 2026-04-08 — v0.1.25)*

---

## 目次

1. [wall-vaultとは？](#wall-vaultとは)
2. [インストール](#インストール)
3. [はじめに（setupウィザード）](#はじめに)
4. [APIキーの登録](#apiキーの登録)
5. [プロキシの使い方](#プロキシの使い方)
6. [キー金庫ダッシュボード](#キー金庫ダッシュボード)
7. [分散モード（マルチボット）](#分散モードマルチボット)
8. [自動起動設定](#自動起動設定)
9. [Doctor（診断ツール）](#doctor診断ツール)
10. [RTKトークン節約](#rtkトークン節約)
11. [環境変数リファレンス](#環境変数リファレンス)
12. [トラブルシューティング](#トラブルシューティング)

---

## wall-vaultとは？

**wall-vault = OpenClaw用AIプロキシ + APIキー金庫**

AIサービスを利用するには**APIキー**が必要です。APIキーとは「この人はこのサービスを使う資格がある」ことを証明する**デジタル入館証**のようなものです。しかし、この入館証には1日の使用回数制限があり、管理を誤ると漏洩のリスクもあります。

wall-vaultはこれらの入館証を安全な金庫に保管し、OpenClawとAIサービスの間で**プロキシ（代理人）**の役割を果たします。簡単に言えば、OpenClawはwall-vaultに接続するだけで、残りの複雑な処理はwall-vaultが自動的に行います。

wall-vaultが解決する問題：

- **APIキーの自動ローテーション**：あるキーの使用量が上限に達したり、一時的にブロックされた場合（クールダウン）、静かに次のキーに切り替えます。OpenClawは中断なく動作し続けます。
- **サービスの自動フォールバック**：Googleが応答しない場合はOpenRouterに、それもダメならローカルにインストールされたOllama・LM Studio・vLLM（ローカルAI）に自動で切り替わります。セッションが途切れることはありません。元のサービスが回復すると、次のリクエストから自動的に復帰します（v0.1.18+、LM Studio/vLLM: v0.1.21+）。
- **リアルタイム同期（SSE）**：金庫ダッシュボードでモデルを変更すると、1〜3秒以内にOpenClaw画面に反映されます。SSE（Server-Sent Events）とは、サーバーが変更をリアルタイムでクライアントにプッシュする技術です。
- **リアルタイム通知**：キー枯渇やサービス障害などのイベントが、OpenClaw TUI（ターミナル画面）の下部にすぐ表示されます。

> 💡 **Claude Code、Cursor、VS Code**も接続できますが、wall-vaultの本来の目的はOpenClawと一緒に使うことです。

```
OpenClaw（TUIターミナル画面）
        │
        ▼
  wall-vault プロキシ (:56244)   ← キー管理、ルーティング、フォールバック、イベント
        │
        ├─ Google Gemini API
        ├─ OpenRouter API（340以上のモデル）
        ├─ Ollama / LM Studio / vLLM（ローカルPC、最後の砦）
        └─ OpenAI / Anthropic API
```

---

## インストール

### Linux / macOS

ターミナルを開いて以下のコマンドをそのまま貼り付けてください。

```bash
# Linux（一般PC、サーバー — amd64）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-linux-amd64 \
  -o ~/.local/bin/wall-vault && chmod +x ~/.local/bin/wall-vault

# macOS Apple Silicon（M1/M2/M3 Mac）
curl -L https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-darwin-arm64 \
  -o /usr/local/bin/wall-vault && chmod +x /usr/local/bin/wall-vault
```

- `curl -L ...` — インターネットからファイルをダウンロードします。
- `chmod +x` — ダウンロードしたファイルを「実行可能」にします。この手順を省略すると「権限がありません」というエラーが出ます。

### Windows

PowerShell（管理者権限）を開いて以下のコマンドを実行してください。

```powershell
# ダウンロード
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# PATHに追加（PowerShell再起動後に適用）
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATHとは？** コンピュータがコマンドを探すフォルダのリストです。PATHに追加すると、どのフォルダからでも`wall-vault`と入力して実行できるようになります。

### ソースからビルド（開発者向け）

Go言語の開発環境がインストールされている場合のみ該当します。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault（バージョン: v0.1.25.YYYYMMDD.HHmmss）
make install     # ~/.local/bin/wall-vault
```

> 💡 **ビルドタイムスタンプバージョン**：`make build`でビルドすると、バージョンは`v0.1.25.20260408.022325`のように日付・時刻を含む形式で自動生成されます。`go build ./...`で直接ビルドすると、バージョンは`"dev"`とだけ表示されます。

---

## はじめに

### setupウィザードの実行

インストール後、最初に必ず以下のコマンドで**セットアップウィザード**を実行してください。ウィザードが必要な項目を一つずつ聞きながら案内してくれます。

```bash
wall-vault setup
```

ウィザードが進行するステップは以下の通りです：

```
1. 言語選択（韓国語を含む10言語）
2. テーマ選択（light / dark / gold / cherry / ocean）
3. 運用モード — 一人で使う（standalone）か、複数台で共有する（distributed）かを選択
4. ボット名 — ダッシュボードに表示される名前
5. ポート設定 — デフォルト: プロキシ 56244、金庫 56243（変更不要ならそのままEnter）
6. AIサービス選択 — Google / OpenRouter / Ollama / LM Studio / vLLM から使うサービス
7. ツールセキュリティフィルター設定
8. 管理者トークン — ダッシュボード管理機能をロックするパスワード。自動生成も可能
9. APIキー暗号化パスワード — キーをより安全に保存したい場合（オプション）
10. 設定ファイルの保存場所
```

> ⚠️ **管理者トークンは必ず覚えておいてください。** 後でダッシュボードでキーを追加したり設定を変更する際に必要です。忘れた場合は設定ファイルを直接編集する必要があります。

ウィザードが完了すると`wall-vault.yaml`設定ファイルが自動生成されます。

### 実行

```bash
wall-vault start
```

以下の2つのサーバーが同時に起動します：

- **プロキシ**（`http://localhost:56244`）— OpenClawとAIサービスを接続する代理人
- **キー金庫**（`http://localhost:56243`）— APIキー管理とWebダッシュボード

ブラウザで`http://localhost:56243`を開くとダッシュボードをすぐに確認できます。

---

## APIキーの登録

APIキーを登録する方法は4つあります。**初めての方には方法1（環境変数）をお勧めします。**

### 方法1：環境変数（推奨 — 最も簡単）

環境変数とは、プログラムが起動時に読み込む**あらかじめ設定された値**です。ターミナルで以下のように入力します。

```bash
# Google Geminiキーの登録
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouterキーの登録
export WV_KEY_OPENROUTER=sk-or-v1-...

# 登録後に実行
wall-vault start
```

キーを複数持っている場合はカンマ（,）で連結してください。wall-vaultがキーを順番に自動で使用します（ラウンドロビン）：

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **ヒント**：`export`コマンドは現在のターミナルセッションにのみ適用されます。コンピュータを再起動しても維持されるようにするには、`~/.bashrc`または`~/.zshrc`ファイルに上記の行を追加してください。

### 方法2：ダッシュボードUI（マウスでクリック）

1. ブラウザで`http://localhost:56243`にアクセス
2. 上部の**🔑 APIキー**カードで`[+ 追加]`ボタンをクリック
3. サービス種類、キー値、ラベル（メモ用の名前）、日次上限を入力して保存

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

### 方法4：proxyフラグ（簡単なテスト用）

正式な登録なしに一時的にキーを入れてテストする場合に使います。プログラムを終了するとキーは消えます。

```bash
wall-vault proxy --key-google=AIzaSy... --key-openrouter=sk-or-...
```

---

## プロキシの使い方

### OpenClawでの使用（主目的）

OpenClawがwall-vaultを通じてAIサービスに接続するための設定方法です。

`~/.openclaw/openclaw.json`ファイルを開いて以下の内容を追加してください：

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
          { id: "wall-vault/hunter-alpha" },    // 無料 1Mコンテキスト
          { id: "wall-vault/claude-opus-4-6" }
        ]
      }
    }
  }
}
```

> 💡 **もっと簡単な方法**：ダッシュボードのエージェントカードの**🦞 OpenClaw設定コピー**ボタンを押すと、トークンとアドレスが入力済みのスニペットがクリップボードにコピーされます。貼り付けるだけです。

**モデル名の先頭の`wall-vault/`はどこに接続されますか？**

モデル名を見て、wall-vaultがどのAIサービスにリクエストを送るかを自動判断します：

| モデル形式 | 接続先サービス |
|----------|-------------|
| `wall-vault/gemini-*` | Google Geminiに直接接続 |
| `wall-vault/gpt-*`、`wall-vault/o3`、`wall-vault/o4*` | OpenAIに直接接続 |
| `wall-vault/claude-*` | OpenRouter経由でAnthropicに接続 |
| `wall-vault/hunter-alpha`、`wall-vault/healer-alpha` | OpenRouter（無料100万トークンコンテキスト） |
| `wall-vault/kimi-*`、`wall-vault/glm-*`、`wall-vault/deepseek-*` | OpenRouterに接続 |
| `google/モデル名`、`openai/モデル名`、`anthropic/モデル名`など | 該当サービスに直接接続 |
| `custom/google/モデル名`、`custom/openai/モデル名`など | `custom/`部分を除去して再ルーティング |
| `モデル名:cloud` | `:cloud`部分を除去してOpenRouterに接続 |

> 💡 **コンテキスト（context）とは？** AIが一度に記憶できる会話の量です。1M（100万トークン）あれば、非常に長い会話や長い文書も一度に処理できます。

### Gemini APIフォーマットでの直接接続（既存ツール互換）

Google Gemini APIを直接使用していたツールがある場合、アドレスをwall-vaultに変更するだけです：

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

またはURLを直接指定するツールの場合：

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### OpenAI SDK（Python）での使用

PythonでAIを活用するコードからもwall-vaultに接続できます。`base_url`を変更するだけです：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # APIキーはwall-vaultが管理します
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # provider/model 形式で入力
    messages=[{"role": "user", "content": "こんにちは"}]
)
```

### 実行中のモデル変更

wall-vaultが既に実行中の状態で使用するAIモデルを変更するには：

```bash
# プロキシに直接リクエストしてモデル変更
curl -X PUT http://localhost:56244/api/config/model \
  -H "Content-Type: application/json" \
  -d '{"service": "openrouter", "model": "anthropic/claude-3.5-sonnet"}'

# 分散モード（マルチボット）では金庫サーバーで変更 → SSEで即時反映
curl -X PUT http://localhost:56243/admin/clients/ボットID \
  -H "Authorization: Bearer 管理者トークン" \
  -H "Content-Type: application/json" \
  -d '{"default_service": "google", "default_model": "gemini-2.5-pro"}'
```

### 利用可能なモデル一覧の確認

```bash
# 全リスト表示
curl http://localhost:56244/api/models | python3 -m json.tool

# Googleモデルのみ表示
curl "http://localhost:56244/api/models?service=google"

# 名前で検索（例: "claude"を含むモデル）
curl "http://localhost:56244/api/models?q=claude"
```

**サービス別主要モデルの概要：**

| サービス | 主要モデル |
|---------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346以上（Hunter Alpha 1Mコンテキスト無料、DeepSeek R1/V3、Qwen 2.5など） |
| Ollama | ローカルサーバーにインストールされたモデルを自動検出 |
| LM Studio | ローカルサーバー（ポート1234） |
| vLLM | ローカルサーバー（ポート8000） |

---

## キー金庫ダッシュボード

ブラウザで`http://localhost:56243`にアクセスするとダッシュボードを表示できます。

**画面構成：**
- **上部固定バー（topbar）**：ロゴ、言語・テーマ選択、SSE接続状態表示
- **カードグリッド**：エージェント・サービス・APIキーカードがタイル形式で配置

### APIキーカード

登録されたAPIキーを一目で管理できるカードです。

- サービス別にキーリストを表示します。
- `today_usage`：今日正常に処理されたトークン（AIが読み書きした文字数）数
- `today_attempts`：今日の総呼び出し回数（成功＋失敗を含む）
- `[+ 追加]`ボタンで新しいキーを登録し、`✕`でキーを削除します。

> 💡 **トークン（token）とは？** AIがテキストを処理する際に使用する単位です。おおよそ英語の単語1つ、または日本語1〜2文字に相当します。API料金は通常このトークン数に基づいて計算されます。

### エージェントカード

wall-vaultプロキシに接続されたボット（エージェント）の状態を表示するカードです。

**接続状態は4段階で表示されます：**

| 表示 | 状態 | 意味 |
|------|------|------|
| 🟢 | 実行中 | プロキシが正常に動作中 |
| 🟡 | 遅延 | 応答はあるが遅い |
| 🔴 | オフライン | プロキシが応答しない |
| ⚫ | 未接続・無効 | プロキシが金庫に接続したことがないか、無効化されている |

**エージェントカード下部のボタン案内：**

エージェントを登録する際に**エージェント種類**を指定すると、その種類に合った便利ボタンが自動的に表示されます。

---

#### 🔘 設定コピーボタン — 接続設定を自動生成します

ボタンをクリックすると、そのエージェントのトークン、プロキシアドレス、モデル情報が入力済みの設定スニペットがクリップボードにコピーされます。コピーした内容を以下の表の場所に貼り付けるだけで接続設定が完了します。

| ボタン | エージェント種類 | 貼り付け先 |
|-------|-------------|----------|
| 🦞 OpenClaw設定コピー | `openclaw` | `~/.openclaw/openclaw.json` |
| 🦀 NanoClaw設定コピー | `nanoclaw` | `~/.openclaw/openclaw.json` |
| 🟠 Claude Code設定コピー | `claude-code` | `~/.claude/settings.json` |
| ⌨ Cursor設定コピー | `cursor` | Cursor → Settings → AI |
| 💻 VSCode設定コピー | `vscode` | `~/.continue/config.json` |

**例 — Claude Codeタイプの場合、以下の内容がコピーされます：**

```json
// ~/.claude/settings.json
{
  "apiProvider": "openai",
  "baseUrl": "http://192.168.0.6:56244/v1",
  "apiKey": "このエージェントのトークン"
}
```

**例 — VSCode（Continue）タイプの場合：**

```yaml
# ~/.continue/config.yaml  ← config.jsonではなくconfig.yamlに貼り付け
name: My Config
version: 0.0.1
schema: v1

models:
  - name: wall-vault proxy
    provider: openai
    model: gemini-2.5-flash
    apiBase: http://192.168.0.6:56244/v1
    apiKey: このエージェントのトークン
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Continueの最新バージョンは`config.yaml`を使用します。** `config.yaml`が存在する場合、`config.json`は完全に無視されます。必ず`config.yaml`に貼り付けてください。

**例 — Cursorタイプの場合：**

```
Base URL : http://192.168.0.6:56244/v1
API Key  : このエージェントのトークン

// または環境変数：
OPENAI_BASE_URL=http://192.168.0.6:56244/v1
OPENAI_API_KEY=このエージェントのトークン
```

> ⚠️ **クリップボードコピーができない場合**：ブラウザのセキュリティポリシーでコピーがブロックされる場合があります。ポップアップでテキストボックスが表示されたら、Ctrl+Aで全選択してCtrl+Cでコピーしてください。

---

#### ⚡ 自動適用ボタン — ワンクリックで設定完了

エージェント種類が`cline`、`claude-code`、`openclaw`、`nanoclaw`の場合、エージェントカードに**⚡ 設定適用**ボタンが表示されます。このボタンを押すと、そのエージェントのローカル設定ファイルが自動更新されます。

| ボタン | エージェント種類 | 対象ファイル |
|-------|-------------|-----------|
| ⚡ Cline設定適用 | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Claude Code設定適用 | `claude-code` | `~/.claude/settings.json` |
| ⚡ OpenClaw設定適用 | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ NanoClaw設定適用 | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ このボタンは**localhost:56244**（ローカルプロキシ）にリクエストを送信します。そのマシンでプロキシが実行中である必要があります。

---

#### 🔀 ドラッグ＆ドロップでカード並べ替え（v0.1.17、改善 v0.1.25）

ダッシュボードのエージェントカードを**ドラッグ**して好きな順序に並べ替えることができます。

1. カード左上の**信号機（●）**エリアをマウスで掴んでドラッグします
2. 目的の位置のカードの上に置くと順序が変わります

> 💡 カード本体（入力フィールド、ボタンなど）はドラッグできません。信号機エリアからのみ掴めます。

#### 🟠 エージェントプロセス検出（v0.1.25）

プロキシは正常に動作しているがローカルエージェントプロセス（NanoClaw、OpenClaw）が停止している場合、カードの信号機が**オレンジ色（点滅）**に変わり、「エージェントプロセス停止」メッセージが表示されます。

- 🟢 緑：プロキシ＋エージェント正常
- 🟠 オレンジ（点滅）：プロキシ正常、エージェント停止
- 🔴 赤：プロキシオフライン
3. 変更された順序は**即座にサーバーに保存**され、リフレッシュしても維持されます

> 💡 タッチデバイス（モバイル/タブレット）ではまだサポートされていません。デスクトップブラウザをご使用ください。

---

#### 🔄 双方向モデル同期（v0.1.16）

金庫ダッシュボードでエージェントのモデルを変更すると、そのエージェントのローカル設定が自動更新されます。

**Clineの場合：**
- 金庫でモデルを変更 → SSEイベント → プロキシが`globalState.json`のモデルフィールドを更新
- 更新対象：`actModeOpenAiModelId`、`planModeOpenAiModelId`、`openAiModelId`
- `openAiBaseUrl`とAPIキーは変更しない
- **VS Codeリロード（`Ctrl+Alt+R`または`Ctrl+Shift+P` → `Developer: Reload Window`）が必要です**
  - Clineは実行中に設定ファイルを再読み込みしないため

**Claude Codeの場合：**
- 金庫でモデルを変更 → SSEイベント → プロキシが`settings.json`の`model`フィールドを更新
- WSLとWindows両方のパスを自動探索（`~/.claude/`、`/mnt/c/Users/*/.claude/`）

**逆方向（エージェント → 金庫）：**
- エージェント（Cline、Claude Codeなど）がプロキシにリクエストを送ると、プロキシがハートビートにそのクライアントのサービス・モデル情報を含めます
- 金庫ダッシュボードのエージェントカードに、現在使用中のサービス/モデルがリアルタイムで表示されます

> 💡 **ポイント**：プロキシはリクエストのAuthorizationトークンでエージェントを識別し、金庫に設定されたサービス/モデルに自動ルーティングします。ClineやClaude Codeが別のモデル名を送っても、プロキシが金庫の設定でオーバーライドします。

---

### VS CodeでClineを使用する — 詳細ガイド

#### ステップ1：Clineのインストール

VS Code拡張機能マーケットプレイスから**Cline**（ID: `saoudrizwan.claude-dev`）をインストールします。

#### ステップ2：金庫にエージェントを登録

1. 金庫ダッシュボード（`http://金庫IP:56243`）を開きます
2. **エージェント**セクションで**+ 追加**をクリック
3. 以下のように入力します：

| フィールド | 値 | 説明 |
|----------|---|------|
| ID | `my_cline` | 一意の識別子（英数字、スペースなし） |
| 名前 | `My Cline` | ダッシュボードに表示される名前 |
| エージェント種類 | `cline` | ← 必ず`cline`を選択 |
| サービス | 使用するサービスを選択（例：`google`） | |
| モデル | 使用するモデルを入力（例：`gemini-2.5-flash`） | |

4. **保存**を押すとトークンが自動生成されます

#### ステップ3：Clineに接続

**方法A — 自動適用（推奨）**

1. そのマシンでwall-vault**プロキシ**が実行中か確認（`localhost:56244`）
2. ダッシュボードのエージェントカードで**⚡ Cline設定適用**ボタンをクリック
3. 「設定適用完了！」通知が出れば成功
4. VS Codeをリロード（`Ctrl+Alt+R`）

**方法B — 手動設定**

Clineサイドバーの設定（⚙️）を開き：
- **API Provider**：`OpenAI Compatible`
- **Base URL**：`http://プロキシアドレス:56244/v1`
  - 同じマシン：`http://localhost:56244/v1`
  - 別のマシン（例：Mac Mini）：`http://192.168.0.6:56244/v1`
- **API Key**：金庫で発行されたトークン（エージェントカードからコピー）
- **Model ID**：金庫で設定したモデル（例：`gemini-2.5-flash`）

#### ステップ4：確認

Clineのチャットに何かメッセージを送信してみます。正常なら：
- 金庫ダッシュボードのエージェントカードに**緑の点（● 実行中）**が表示されます
- カードに現在のサービス/モデルが表示されます（例：`google / gemini-2.5-flash`）

#### モデル変更

Clineのモデルを変更したい場合は**金庫ダッシュボード**から変更してください：

1. エージェントカードのサービス/モデルドロップダウンを変更
2. **適用**をクリック
3. VS Codeをリロード（`Ctrl+Alt+R`）— Clineフッターのモデル名が更新されます
4. 次のリクエストから新しいモデルが使用されます

> 💡 実際にはプロキシがClineのリクエストをトークンで識別し、金庫設定のモデルにルーティングします。VS Codeをリロードしなくても**実際に使用されるモデルは即座に変わります** — リロードはCline UIのモデル表示を更新するためのものです。

#### 切断検出

VS Codeを閉じると、金庫ダッシュボードのエージェントカードは約**90秒**後に黄色（遅延）に、**3分**後に赤色（オフライン）に変わります。（v0.1.18から15秒間隔の状態確認でオフライン検出が速くなりました。）

#### トラブルシューティング

| 症状 | 原因 | 解決方法 |
|------|------|---------|
| Clineで「接続失敗」エラー | プロキシ未実行またはアドレスが間違い | `curl http://localhost:56244/health`でプロキシを確認 |
| 金庫で緑の点が出ない | APIキー（トークン）が設定されていない | **⚡ Cline設定適用**ボタンを再度クリック |
| Clineフッターのモデルが変わらない | Clineが設定をキャッシュ中 | VS Codeリロード（`Ctrl+Alt+R`） |
| 間違ったモデル名が表示される | 旧バグ（v0.1.16で修正済み） | プロキシをv0.1.16以上にアップデート |

---

#### 🟣 デプロイコマンドコピーボタン — 新しいマシンへのインストール用

新しいコンピュータにwall-vaultプロキシを初めてインストールし、金庫に接続する際に使用します。ボタンをクリックするとインストールスクリプト全体がコピーされます。新しいコンピュータのターミナルに貼り付けて実行すると、以下が一度に処理されます：

1. wall-vaultバイナリのインストール（既にインストール済みの場合はスキップ）
2. systemdユーザーサービスの自動登録
3. サービス起動と金庫への自動接続

> 💡 スクリプトにはこのエージェントのトークンと金庫サーバーアドレスが既に入力されているため、貼り付け後に修正なしですぐに実行できます。

---

### サービスカード

使用するAIサービスのオン/オフや設定を行うカードです。

- サービスごとの有効化・無効化トグルスイッチ
- ローカルAIサーバー（自分のコンピュータで実行するOllama、LM Studio、vLLMなど）のアドレスを入力すると、利用可能なモデルを自動検出します。
- **ローカルサービス接続状態表示**：サービス名の横の●点が**緑色**なら接続済み、**灰色**なら未接続
- **ローカルサービス自動信号機**（v0.1.23+）：ローカルサービス（Ollama、LM Studio、vLLM）は接続可否に応じて自動的に有効化/無効化されます。サービスが接続されると15秒以内に●が緑色に変わりチェックボックスがオンになり、切断されると自動的にオフになります。クラウドサービス（Google、OpenRouterなど）がAPIキーの有無に応じて自動トグルされるのと同じ方式です。

> 💡 **ローカルサービスが別のコンピュータで実行中の場合**：サービスURL入力欄にそのコンピュータのIPを入力してください。例：`http://192.168.0.6:11434`（Ollama）、`http://192.168.0.6:1234`（LM Studio）。サービスが`0.0.0.0`ではなく`127.0.0.1`にのみバインドされている場合、外部IPからのアクセスができません。サービス設定でバインドアドレスを確認してください。

### 管理者トークン入力

ダッシュボードでキーの追加・削除など重要な機能を使おうとすると、管理者トークン入力ポップアップが表示されます。setupウィザードで設定したトークンを入力してください。一度入力すると、ブラウザを閉じるまで維持されます。

> ⚠️ **認証失敗が15分以内に10回を超えると、そのIPが一時的にブロックされます。** トークンを忘れた場合は、`wall-vault.yaml`ファイルの`admin_token`項目を確認してください。

---

## 分散モード（マルチボット）

複数のコンピュータでOpenClawを同時に運用する際に、**一つのキー金庫を共有**する構成です。キー管理を一箇所でまとめて行えるので便利です。

### 構成例

```
[キー金庫サーバー]
  wall-vault vault    (キー金庫 :56243、ダッシュボード)

[WSLアルファ]         [ラズベリーパイ ガンマ]    [Mac Miniローカル]
  wall-vault proxy      wall-vault proxy        wall-vault proxy
  openclaw TUI          openclaw TUI            openclaw TUI
  ↕ SSE同期             ↕ SSE同期               ↕ SSE同期
```

すべてのボットが中央の金庫サーバーを参照しているため、金庫でモデルを変更したりキーを追加すると、すべてのボットに即座に反映されます。

### ステップ1：キー金庫サーバーの起動

金庫サーバーとして使うコンピュータで実行します：

```bash
wall-vault vault
```

### ステップ2：各ボット（クライアント）の登録

金庫サーバーに接続する各ボットの情報を事前に登録しておきます：

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

### ステップ3：各ボットのコンピュータでプロキシを起動

ボットがインストールされた各コンピュータで、金庫サーバーのアドレスとトークンを指定してプロキシを実行します：

```bash
WV_VAULT_URL=http://192.168.x.x:56243 \
WV_VAULT_TOKEN=bota-secret \
WV_VAULT_CLIENT_ID=botA \
wall-vault proxy
```

> 💡 **`192.168.x.x`**の部分は金庫サーバーのコンピュータの実際の内部IPアドレスに置き換えてください。ルーター設定または`ip addr`コマンドで確認できます。

---

## 自動起動設定

コンピュータを再起動するたびに手動でwall-vaultを起動するのが面倒な場合は、システムサービスとして登録しておきましょう。一度登録すれば起動時に自動的に開始されます。

### Linux — systemd（ほとんどのLinux）

systemdはLinuxでプログラムを自動起動・管理するシステムです：

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

1. [nssm.cc](https://nssm.cc/download)からNSSMをダウンロードしてPATHに追加します。
2. 管理者権限のPowerShellで：

```powershell
wall-vault doctor deploy windows
```

---

## Doctor（診断ツール）

`doctor`コマンドは、wall-vaultが正しく設定されているかを**自己診断し修復するツール**です。

```bash
wall-vault doctor check   # 現在の状態を診断（読み取りのみ、何も変更しない）
wall-vault doctor fix     # 問題を自動修復
wall-vault doctor all     # 診断＋自動修復を一度に
```

> 💡 何かおかしいと感じたら、まず`wall-vault doctor all`を実行してみてください。多くの問題を自動的に解決してくれます。

---

## RTKトークン節約

*(v0.1.24+)*

**RTK（トークン節約ツール）**は、AIコーディングエージェント（Claude Codeなど）が実行するシェルコマンドの出力を自動圧縮して、トークン使用量を削減します。例えば、`git status`の15行の出力が2行の要約に圧縮されます。

### 基本的な使い方

```bash
# コマンドをwall-vault rtkで包むと出力が自動フィルタリングされます
wall-vault rtk git status          # 変更されたファイルリストのみ表示
wall-vault rtk git diff HEAD~1     # 変更行＋最小コンテキストのみ
wall-vault rtk git log -10         # ハッシュ＋1行メッセージ
wall-vault rtk go test ./...       # 失敗したテストのみ表示
wall-vault rtk ls -la              # 未対応コマンドは自動切り詰め
```

### 対応コマンドと削減効果

| コマンド | フィルター方式 | 削減率 |
|---------|-------------|--------|
| `git status` | 変更ファイル要約のみ | ~87% |
| `git diff` | 変更行＋3行コンテキスト | ~60-94% |
| `git log` | ハッシュ＋最初の1行メッセージ | ~90% |
| `git push/pull/fetch` | 進捗除去、要約のみ | ~80% |
| `go test` | 失敗のみ表示、合格はカウント | ~88-99% |
| `go build/vet` | エラーのみ表示 | ~90% |
| その他すべてのコマンド | 先頭50行＋末尾50行、最大32KB | 可変 |

### 3段階フィルターパイプライン

1. **コマンド別構造フィルター** — git、goなどの出力形式を理解し、意味のある部分のみ抽出
2. **正規表現後処理** — ANSIカラーコード除去、空白行の圧縮、重複行の集約
3. **パススルー＋切り詰め** — 未対応コマンドは先頭/末尾50行のみ保持

### Claude Code連携

Claude Codeの`PreToolUse`フックで、すべてのシェルコマンドを自動的にRTKを通すよう設定できます。

```bash
# フックインストール（Claude Code settings.jsonに自動追加）
wall-vault rtk hook install
```

または手動で`~/.claude/settings.json`に追加：

```json
{
  "hooks": {
    "PreToolUse": [{
      "matcher": "Bash",
      "command": "wall-vault rtk rewrite"
    }]
  }
}
```

> 💡 **終了コードの保持**：RTKは元のコマンドの終了コードをそのまま返します。コマンドが失敗した場合（終了コード ≠ 0）、AIも正確に失敗を検出します。

> 💡 **英語出力の強制**：RTKは`LC_ALL=C`でコマンドを実行し、システムの言語設定に関係なく常に英語の出力を生成します。これによりフィルターが正確に動作します。

---

## 環境変数リファレンス

環境変数はプログラムに設定値を渡す方法です。ターミナルで`export 変数名=値`の形式で入力するか、自動起動サービスファイルに記述しておけば常に適用されます。

| 変数 | 説明 | 値の例 |
|------|------|--------|
| `WV_LANG` | ダッシュボード言語 | `ko`、`en`、`ja` |
| `WV_THEME` | ダッシュボードテーマ | `light`、`dark`、`gold` |
| `WV_KEY_GOOGLE` | Google APIキー（カンマで複数） | `AIza...,AIza...` |
| `WV_KEY_OPENROUTER` | OpenRouter APIキー | `sk-or-v1-...` |
| `WV_VAULT_URL` | 分散モードでの金庫サーバーアドレス | `http://192.168.x.x:56243` |
| `WV_VAULT_TOKEN` | クライアント（ボット）認証トークン | `my-secret-token` |
| `WV_ADMIN_TOKEN` | 管理者トークン | `admin-token-here` |
| `WV_MASTER_PASS` | APIキー暗号化パスワード | `my-password` |
| `WV_AVATAR` | アバター画像ファイルパス（`~/.openclaw/`基準の相対パス） | `workspace/avatars/avatar.png` |
| `OLLAMA_URL` | Ollamaローカルサーバーアドレス | `http://192.168.x.x:11434` |

---

## トラブルシューティング

### プロキシが起動しない場合

ポートが既に他のプログラムに使用されている場合が多いです。

```bash
ss -tlnp | grep 56244   # ポート56244を使っているプロセスを確認
wall-vault proxy --port 8080   # 別のポート番号で起動
```

### APIキーエラーが出る場合（429, 402, 401, 403, 582）

| エラーコード | 意味 | 対処方法 |
|------------|------|---------|
| **429** | リクエストが多すぎる（使用量超過） | しばらく待つか、別のキーを追加 |
| **402** | 支払いが必要またはクレジット不足 | 該当サービスでクレジットをチャージ |
| **401 / 403** | キーが無効または権限なし | キー値を再確認して再登録 |
| **582** | ゲートウェイ過負荷（5分間クールダウン） | 5分後に自動解除 |

```bash
# 登録済みキーのリストと状態を確認
curl -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys

# キー使用量カウンターのリセット
curl -X POST -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys/reset
```

### エージェントが「未接続」と表示される場合

「未接続」とは、プロキシプロセスが金庫にハートビート信号を送っていない状態です。**設定が保存されていないという意味ではありません。** プロキシが金庫サーバーのアドレスとトークンを持って実行されている必要があります。

```bash
# 金庫サーバーアドレス、トークン、クライアントIDを指定してプロキシを起動
WV_VAULT_URL=http://金庫サーバーアドレス:56243 \
WV_VAULT_TOKEN=クライアントトークン \
WV_VAULT_CLIENT_ID=クライアントID \
wall-vault proxy
```

接続に成功すると、約20秒以内にダッシュボードで🟢 実行中に変わります。

### Ollamaに接続できない場合

Ollamaは自分のコンピュータで直接AIを実行するプログラムです。まずOllamaが起動しているか確認してください。

```bash
curl http://localhost:11434/api/tags   # モデルリストが表示されれば正常
export OLLAMA_URL=http://192.168.x.x:11434   # 別のコンピュータで実行中の場合
```

> ⚠️ Ollamaが応答しない場合は、`ollama serve`コマンドでまずOllamaを起動してください。

> ⚠️ **大型モデルは応答が遅いです**：`qwen3.5:35b`や`deepseek-r1`のような大きなモデルは応答生成まで数分かかることがあります。応答がないように見えても正常に処理中の場合がありますので、お待ちください。

---

## 最近の変更点（v0.1.16 ~ v0.1.25）

### v0.1.25（2026-04-08）
- **エージェントプロセス検出**：プロキシがローカルエージェント（NanoClaw/OpenClaw）の生存状態を検出し、ダッシュボードにオレンジ色の信号機で表示します。
- **ドラッグハンドルの改善**：カード並べ替え時、信号機（●）エリアからのみ掴めるように変更。入力フィールドやボタンでの誤ドラッグを防止します。

### v0.1.24（2026-04-06）
- **RTKトークン節約サブコマンド**：`wall-vault rtk <command>`でシェルコマンドの出力を自動フィルタリングし、AIエージェントのトークン使用量を60〜90%削減します。git、goなどの主要コマンド用の専用フィルターを内蔵し、未対応コマンドも自動切り詰めします。Claude Code `PreToolUse`フックで透過的に連携します。

### v0.1.23（2026-04-06）
- **Ollamaモデル変更の修正**：金庫ダッシュボードでOllamaモデルを変更しても実際にプロキシに反映されない問題を修正。以前は環境変数（`OLLAMA_MODEL`）のみ使用していましたが、今後は金庫設定が優先されます。
- **ローカルサービス自動信号機**：Ollama、LM Studio、vLLMが接続可能な場合は自動有効化、切断時は自動無効化されます。クラウドサービスのキーベース自動トグルと同じ仕組みです。

### v0.1.22（2026-04-05）
- **空のcontentフィールド修正**：thinkingモデル（gemini-3.1-pro、o1、claude thinkingなど）がmax_tokensをreasoningに使い切って実際の応答を生成できない場合、プロキシが`content`/`text`フィールドを`omitempty`で省略し、OpenAI/Anthropic SDKクライアントが`Cannot read properties of undefined (reading 'trim')`エラーでクラッシュする問題を修正。公式APIスペック通り常にフィールドを含むように変更。

### v0.1.21（2026-04-05）
- **Gemma 4モデルサポート**：Google Gemini APIを通じて`gemma-4-31b-it`や`gemma-4-26b-a4b-it`などのGemmaモデルが使用できるようになりました。
- **LM Studio / vLLM正式サポート**：以前はこれらのサービスがプロキシルーティングから漏れており、常にOllamaにフォールバックしていました。現在はOpenAI互換APIで正常にルーティングされます。
- **ダッシュボードサービス表示の修正**：フォールバックが発生しても、ダッシュボードには常にユーザーが設定したサービスが表示されます。
- **ローカルサービス状態表示**：ダッシュボード読み込み時にローカルサービス（Ollama、LM Studio、vLLMなど）の接続状態を●点の色で表示します。
- **ツールフィルター環境変数**：`WV_TOOL_FILTER=passthrough`環境変数でツール転送モードを設定できます。

### v0.1.20（2026-03-28）
- **包括的セキュリティ強化**：XSS防止（41箇所）、定数時間トークン比較、CORS制限、リクエストサイズ制限、パストラバーサル防止、SSE認証、レートリミッター強化など12のセキュリティ項目を改善。

### v0.1.19（2026-03-27）
- **Claude Codeオンライン検出**：プロキシを経由しないClaude Codeもダッシュボードでオンラインとして表示されます。

### v0.1.18（2026-03-26）
- **フォールバックサービスの固着修正**：一時的なエラーでOllamaにフォールバック後、元のサービスが回復すると自動復帰します。
- **オフライン検出の改善**：15秒間隔の状態確認でプロキシ停止の検出が速くなりました。

### v0.1.17（2026-03-25）
- **ドラッグ＆ドロップカード並べ替え**：エージェントカードをドラッグして順序を変更できます。
- **インライン設定適用ボタン**：オフラインエージェントに[⚡ 設定適用]ボタンが表示されます。
- **cokacdir エージェントタイプ追加**。

### v0.1.16（2026-03-25）
- **双方向モデル同期**：金庫ダッシュボードからCline・Claude Codeのモデルを変更すると自動反映されます。

---

*詳しいAPI情報は[API.md](API.md)を参照してください。*
