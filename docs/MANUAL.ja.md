# wall-vault ユーザーマニュアル
*(最終更新: 2026-04-06 — v0.1.24)*

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
9. [Doctor 自己診断ツール](#doctor-自己診断ツール)
10. [RTK トークン節約](#rtk-トークン節約)
11. [環境変数リファレンス](#環境変数リファレンス)
12. [トラブルシューティング](#トラブルシューティング)

---

## wall-vaultとは？

**wall-vault = OpenClaw用のAIプロキシ + APIキー金庫**

AIサービスを利用するには**APIキー**が必要です。APIキーとは「この人はこのサービスを使う資格がある」ことを証明する**デジタル入場証**のようなものです。この入場証には一日に使える回数が決まっており、管理を誤ると漏洩のリスクもあります。

wall-vaultはこれらの入場証を安全な金庫に保管し、OpenClawとAIサービスの間で**代理人（プロキシ）**の役割を果たします。つまり、OpenClawはwall-vaultに接続するだけでよく、残りの複雑な処理はwall-vaultが自動的に行います。

wall-vaultが解決する問題：

- **APIキー自動ローテーション**：あるキーの使用量が上限に達したり、一時的にブロックされた（クールダウン）場合、静かに次のキーに切り替えます。OpenClawは中断なく動作し続けます。
- **サービス自動フォールバック**：Googleが応答しなければOpenRouterに、それもダメならOllama・LM Studio・vLLM（ローカルAI）に自動切り替えします。セッションが切れることはありません。元のサービスが復旧すれば次のリクエストから自動的に復帰します（v0.1.18+、LM Studio/vLLM: v0.1.21+）。
- **リアルタイム同期（SSE）**：金庫ダッシュボードでモデルを変更すると、1〜3秒以内にOpenClawの画面に反映されます。SSE（Server-Sent Events）とは、サーバーがリアルタイムでクライアントに変更をプッシュする技術です。
- **リアルタイム通知**：キー枯渇やサービス障害などのイベントがOpenClawのTUI（ターミナル画面）下部にすぐに表示されます。

> 💡 **Claude Code、Cursor、VS Code**も接続して使えますが、wall-vaultの本来の目的はOpenClawと一緒に使うことです。

```
OpenClaw（TUIターミナル画面）
        │
        ▼
  wall-vault プロキシ (:56244)   ← キー管理、ルーティング、フォールバック、イベント
        │
        ├─ Google Gemini API
        ├─ OpenRouter API（340以上のモデル）
        ├─ Ollama / LM Studio / vLLM（ローカルマシン、最終手段）
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

- `curl -L ...` — ファイルをインターネットからダウンロードします。
- `chmod +x` — ダウンロードしたファイルを「実行可能」にします。この手順を省略すると「権限がありません」エラーが出ます。

### Windows

PowerShell（管理者権限）を開いて以下のコマンドを実行してください。

```powershell
# ダウンロード
Invoke-WebRequest -Uri `
  "https://github.com/sookmook/wall-vault/releases/latest/download/wall-vault-windows-amd64.exe" `
  -OutFile "$env:LOCALAPPDATA\Programs\wall-vault\wall-vault.exe"

# PATHに追加（PowerShell再起動後に反映）
$env:PATH += ";$env:LOCALAPPDATA\Programs\wall-vault"
```

> 💡 **PATHとは？** コンピュータがコマンドを探すフォルダのリストです。PATHに追加すれば、どのフォルダからでも`wall-vault`と入力して実行できます。

### ソースからのビルド（開発者向け）

Go言語の開発環境がインストールされている場合のみ該当します。

```bash
git clone https://github.com/sookmook/wall-vault
cd wall-vault
make build       # bin/wall-vault（バージョン: v0.1.24.YYYYMMDD.HHmmss）
make install     # ~/.local/bin/wall-vault
```

> 💡 **ビルドタイムスタンプバージョン**：`make build`でビルドすると、バージョンが`v0.1.24.20260406.225957`のように日付・時刻を含む形式で自動生成されます。`go build ./...`で直接ビルドすると、バージョンは`"dev"`とだけ表示されます。

---

## はじめに

### setupウィザードの実行

インストール後、必ず以下のコマンドで**設定ウィザード**を実行してください。ウィザードが必要な項目を一つずつ質問しながら案内します。

```bash
wall-vault setup
```

ウィザードが進める手順は以下の通りです：

```
1. 言語選択（日本語を含む10言語）
2. テーマ選択（light / dark / gold / cherry / ocean）
3. 運用モード — 単独使用（standalone）か複数マシンで共有（distributed）かを選択
4. ボット名入力 — ダッシュボードに表示される名前
5. ポート設定 — デフォルト：プロキシ56244、金庫56243（変更不要ならそのままEnter）
6. AIサービス選択 — Google / OpenRouter / Ollama / LM Studio / vLLM
7. ツールセキュリティフィルター設定
8. 管理者トークン設定 — ダッシュボード管理機能をロックするパスワード。自動生成も可能
9. APIキー暗号化パスワード設定 — キーをより安全に保存したい場合（任意）
10. 設定ファイル保存先
```

> ⚠️ **管理者トークンは必ず覚えておいてください。** 後でダッシュボードでキーを追加したり設定を変更する際に必要です。忘れた場合は設定ファイルを直接編集する必要があります。

ウィザードが完了すると`wall-vault.yaml`設定ファイルが自動的に生成されます。

### 起動

```bash
wall-vault start
```

以下の2つのサーバーが同時に起動します：

- **プロキシ**（`http://localhost:56244`）— OpenClawとAIサービスを接続する代理人
- **キー金庫**（`http://localhost:56243`）— APIキー管理およびWebダッシュボード

ブラウザで`http://localhost:56243`を開くとダッシュボードをすぐに確認できます。

---

## APIキーの登録

APIキーを登録する方法は4つあります。**初めての方には方法1（環境変数）をお勧めします。**

### 方法1：環境変数（推奨 — 最も簡単）

環境変数とは、プログラムが起動時に読み取る**事前に設定された値**です。ターミナルで以下のように入力してください。

```bash
# Google Geminiキーを登録
export WV_KEY_GOOGLE=AIzaSy...

# OpenRouterキーを登録
export WV_KEY_OPENROUTER=sk-or-v1-...

# 登録後に起動
wall-vault start
```

キーを複数持っている場合はカンマ（,）で繋げてください。wall-vaultがキーを順番に自動で使用します（ラウンドロビン）：

```bash
export WV_KEY_GOOGLE=AIzaSy...,AIzaSy...,AIzaSy...
```

> 💡 **ヒント**：`export`コマンドは現在のターミナルセッションにのみ適用されます。コンピュータを再起動しても維持するには、`~/.bashrc`または`~/.zshrc`ファイルに上記の行を追加してください。

### 方法2：ダッシュボードUI（マウスでクリック）

1. ブラウザで`http://localhost:56243`にアクセス
2. 上部の **🔑 APIキー** カードで`[+ 追加]`ボタンをクリック
3. サービス種類、キー値、ラベル（メモ用の名前）、日次制限を入力して保存

### 方法3：REST API（自動化・スクリプト用）

REST APIとは、プログラム間でHTTPを介してデータをやり取りする方法です。スクリプトによる自動登録に便利です。

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

### 方法4：proxyフラグ（ちょっとしたテスト用）

正式登録なしで一時的にキーを入れてテストする場合に使います。プログラムを終了すると消えます。

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
        apiKey: "your-agent-token",   // 金庫エージェントトークン
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

> 💡 **より簡単な方法**：ダッシュボードのエージェントカードにある **🦞 OpenClaw設定コピー** ボタンを押すと、トークンとアドレスが既に入力されたスニペットがクリップボードにコピーされます。貼り付けるだけです。

**モデル名の先頭の`wall-vault/`はどこに接続されるのか？**

モデル名を見て、wall-vaultがどのAIサービスにリクエストを送るかを自動的に判断します：

| モデル形式 | 接続先 |
|----------|--------|
| `wall-vault/gemini-*` | Google Gemini直接接続 |
| `wall-vault/gpt-*`, `wall-vault/o3`, `wall-vault/o4*` | OpenAI直接接続 |
| `wall-vault/claude-*` | OpenRouter経由でAnthropic接続 |
| `wall-vault/hunter-alpha`, `wall-vault/healer-alpha` | OpenRouter（無料100万トークンコンテキスト） |
| `wall-vault/kimi-*`, `wall-vault/glm-*`, `wall-vault/deepseek-*` | OpenRouter接続 |
| `google/モデル名`, `openai/モデル名`, `anthropic/モデル名`等 | 該当サービスに直接接続 |
| `custom/google/モデル名`, `custom/openai/モデル名`等 | `custom/`部分を除去して再ルーティング |
| `モデル名:cloud` | `:cloud`部分を除去してOpenRouter接続 |

> 💡 **コンテキスト（context）とは？** AIが一度に記憶できる会話の量です。1M（100万トークン）なら非常に長い会話や長い文書も一度に処理できます。

### Gemini API形式での直接接続（既存ツール互換）

Google Gemini APIを直接使っていたツールがある場合、アドレスをwall-vaultに変更するだけです：

```bash
export ANTHROPIC_BASE_URL=http://localhost:56244/google
```

またはURLを直接指定するツールの場合：

```
http://localhost:56244/google/v1beta/models/gemini-2.5-flash:generateContent
```

### OpenAI SDK（Python）での使用

PythonでAIを活用するコードでもwall-vaultを接続できます。`base_url`を変更するだけです：

```python
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:56244/v1",
    api_key="not-needed"  # APIキーはwall-vaultが管理します
)

response = client.chat.completions.create(
    model="google/gemini-2.5-flash",   # provider/model形式で入力
    messages=[{"role": "user", "content": "こんにちは"}]
)
```

### 実行中にモデルを変更する

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
# 全一覧を表示
curl http://localhost:56244/api/models | python3 -m json.tool

# Googleモデルのみ表示
curl "http://localhost:56244/api/models?service=google"

# 名前で検索（例：「claude」を含むモデル）
curl "http://localhost:56244/api/models?q=claude"
```

**サービス別主要モデルの概要：**

| サービス | 主要モデル |
|---------|----------|
| Google | gemini-2.5-pro, gemini-2.5-flash, gemini-2.5-flash-8b, gemini-2.0-flash |
| OpenAI | gpt-4o, gpt-4o-mini, o3, o1, o1-mini |
| OpenRouter | 346以上（Hunter Alpha 1Mコンテキスト無料、DeepSeek R1/V3、Qwen 2.5等） |
| Ollama | ローカルにインストール済みのモデルを自動検出 |
| LM Studio | ローカルサーバー（ポート1234） |
| vLLM | ローカルサーバー（ポート8000） |

---

## キー金庫ダッシュボード

ブラウザで`http://localhost:56243`にアクセスするとダッシュボードが表示されます。

**画面構成：**
- **上部固定バー（topbar）**：ロゴ、言語・テーマセレクター、SSE接続ステータス表示
- **カードグリッド**：エージェント・サービス・APIキーカードがタイル形式で配置

### APIキーカード

登録されたAPIキーを一目で管理できるカードです。

- サービス別にキー一覧を表示します。
- `today_usage`：今日正常に処理されたトークン（AIが読み書きした文字数の単位）数
- `today_attempts`：今日の合計呼び出し回数（成功 + 失敗を含む）
- `[+ 追加]`ボタンで新しいキーを登録し、`✕`でキーを削除します。

> 💡 **トークン（token）とは？** AIがテキストを処理する際に使用する単位です。おおよそ英単語1つ、または日本語の1〜2文字に相当します。API料金は通常このトークン数で計算されます。

### エージェントカード

wall-vaultプロキシに接続されたボット（エージェント）のステータスを表示するカードです。

**接続ステータスは4段階で表示されます：**

| 表示 | ステータス | 意味 |
|------|---------|------|
| 🟢 | 実行中 | プロキシが正常に動作中 |
| 🟡 | 遅延 | 応答はあるが遅い |
| 🔴 | オフライン | プロキシが応答していない |
| ⚫ | 未接続・無効 | プロキシが金庫に接続したことがない、または無効化されている |

**エージェントカード下部のボタン案内：**

エージェントを登録する際に**エージェント種類**を指定すると、その種類に合った便利ボタンが自動的に表示されます。

---

#### 🔘 設定コピーボタン — 接続設定を自動生成します

ボタンをクリックすると、そのエージェントのトークン、プロキシアドレス、モデル情報が既に入力された設定スニペットがクリップボードにコピーされます。コピーした内容を以下の表の場所に貼り付けるだけで接続設定が完了します。

| ボタン | エージェント種類 | 貼り付け先 |
|-------|--------------|----------|
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
  "baseUrl": "http://192.168.1.20:56244/v1",
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
    apiBase: http://192.168.1.20:56244/v1
    apiKey: このエージェントのトークン
    roles:
      - chat
      - edit
      - apply
```

> ⚠️ **Continueの最新バージョンは`config.yaml`を使用します。** `config.yaml`が存在すれば`config.json`は完全に無視されます。必ず`config.yaml`に貼り付けてください。

**例 — Cursorタイプの場合：**

```
Base URL : http://192.168.1.20:56244/v1
API Key  : このエージェントのトークン

// または環境変数:
OPENAI_BASE_URL=http://192.168.1.20:56244/v1
OPENAI_API_KEY=このエージェントのトークン
```

> ⚠️ **クリップボードコピーができない場合**：ブラウザのセキュリティポリシーでコピーがブロックされることがあります。ポップアップでテキストボックスが表示されたら、Ctrl+Aで全選択後、Ctrl+Cでコピーしてください。

---

#### ⚡ 自動適用ボタン — ワンクリックで設定完了

エージェント種類が`cline`、`claude-code`、`openclaw`、`nanoclaw`の場合、エージェントカードに **⚡ 設定適用** ボタンが表示されます。このボタンを押すと、そのエージェントのローカル設定ファイルが自動的に更新されます。

| ボタン | エージェント種類 | 適用対象ファイル |
|-------|--------------|--------------|
| ⚡ Cline設定適用 | `cline` | `~/.cline/data/globalState.json` + `secrets.json` |
| ⚡ Claude Code設定適用 | `claude-code` | `~/.claude/settings.json` |
| ⚡ OpenClaw設定適用 | `openclaw` | `~/.openclaw/openclaw.json` |
| ⚡ NanoClaw設定適用 | `nanoclaw` | `~/.openclaw/openclaw.json` |

> ⚠️ このボタンは **localhost:56244**（ローカルプロキシ）にリクエストを送信します。そのマシンでプロキシが実行中でなければ動作しません。

---

#### 🔀 ドラッグ＆ドロップカード並べ替え（v0.1.17）

ダッシュボードのエージェントカードを**ドラッグ**して好きな順序に並べ替えることができます。

1. エージェントカードをマウスで掴んでドラッグします
2. 目的の位置のカードの上にドロップすると順序が入れ替わります
3. 変更された順序は**即座にサーバーに保存**され、リロードしても維持されます

> 💡 タッチデバイス（モバイル/タブレット）はまだサポートされていません。デスクトップブラウザをご利用ください。

---

#### 🔄 双方向モデル同期（v0.1.16）

金庫ダッシュボードでエージェントのモデルを変更すると、そのエージェントのローカル設定が自動的に更新されます。

**Clineの場合：**
- 金庫でモデルを変更 → SSEイベント → プロキシが`globalState.json`のモデルフィールドを更新
- 更新対象：`actModeOpenAiModelId`、`planModeOpenAiModelId`、`openAiModelId`
- `openAiBaseUrl`とAPIキーは変更しません
- **VS Codeリロード（`Ctrl+Alt+R`または`Ctrl+Shift+P` → `Developer: Reload Window`）が必要です**
  - Clineは実行中に設定ファイルを再読み込みしないため

**Claude Codeの場合：**
- 金庫でモデルを変更 → SSEイベント → プロキシが`settings.json`の`model`フィールドを更新
- WSLとWindows両方のパスを自動探索（`~/.claude/`、`/mnt/c/Users/*/.claude/`）

**逆方向（エージェント → 金庫）：**
- エージェント（Cline、Claude Code等）がプロキシにリクエストを送ると、プロキシがハートビートにそのクライアントのサービス・モデル情報を含めます
- 金庫ダッシュボードのエージェントカードに現在使用中のサービス/モデルがリアルタイムで表示されます

> 💡 **要点**：プロキシはリクエストのAuthorizationトークンでエージェントを識別し、金庫に設定されたサービス/モデルへ自動ルーティングします。ClineやClaude Codeが別のモデル名を送っても、プロキシが金庫の設定でオーバーライドします。

---

### VS CodeでClineを使う — 詳細ガイド

#### ステップ1：Clineのインストール

VS Code拡張マーケットプレイスから **Cline**（ID: `saoudrizwan.claude-dev`）をインストールします。

#### ステップ2：金庫にエージェントを登録

1. 金庫ダッシュボード（`http://金庫IP:56243`）を開きます
2. **エージェント** セクションで **+ 追加** をクリック
3. 以下のように入力します：

| フィールド | 値 | 説明 |
|---------|-----|------|
| ID | `my_cline` | 一意の識別子（英数字、スペースなし） |
| 名前 | `My Cline` | ダッシュボードに表示される名前 |
| エージェント種類 | `cline` | ← 必ず`cline`を選択 |
| サービス | 使用するサービスを選択（例：`google`） | |
| モデル | 使用するモデルを入力（例：`gemini-2.5-flash`） | |

4. **保存** をクリック — トークンが自動生成されます

#### ステップ3：Clineに接続

**方法A — 自動適用（推奨）**

1. そのマシンでwall-vault **プロキシ** が実行中であることを確認（`localhost:56244`）
2. ダッシュボードのエージェントカードで **⚡ Cline設定適用** ボタンをクリック
3. 「設定適用完了！」通知が表示されたら成功
4. VS Codeをリロード（`Ctrl+Alt+R`）

**方法B — 手動設定**

Clineサイドバーで設定（⚙️）を開き：
- **API Provider**：`OpenAI Compatible`
- **Base URL**：`http://プロキシアドレス:56244/v1`
  - 同じマシンなら`http://localhost:56244/v1`
  - ミニサーバー等、別のマシンなら`http://192.168.1.20:56244/v1`
- **API Key**：金庫で発行されたトークン（エージェントカードからコピー）
- **Model ID**：金庫で設定したモデル（例：`gemini-2.5-flash`）

#### ステップ4：確認

Clineのチャット画面で何かメッセージを送ってみます。正常であれば：
- 金庫ダッシュボードの該当エージェントカードに **緑の点（● 実行中）** が表示されます
- カードに現在のサービス/モデルが表示されます（例：`google / gemini-2.5-flash`）

#### モデルの変更

Clineのモデルを変更したい場合は **金庫ダッシュボード** から変更してください：

1. エージェントカードのサービス/モデルドロップダウンを変更
2. **適用** をクリック
3. VS Codeをリロード（`Ctrl+Alt+R`）— Clineフッターのモデル名が更新されます
4. 次のリクエストから新しいモデルが使用されます

> 💡 実際にはプロキシがClineのリクエストをトークンで識別し、金庫設定のモデルにルーティングします。VS Codeをリロードしなくても**実際に使用されるモデルは即座に変わります** — リロードはCline UIのモデル表示を更新するためのものです。

#### 接続切断の検知

VS Codeを閉じると、金庫ダッシュボードで約**90秒**後にエージェントカードが黄色（遅延）に、**3分**後に赤（オフライン）に変わります。（v0.1.18から15秒間隔のステータス確認でオフライン検知が速くなりました。）

#### トラブルシューティング

| 症状 | 原因 | 解決方法 |
|------|------|---------|
| Clineで「接続失敗」エラー | プロキシ未実行またはアドレス間違い | `curl http://localhost:56244/health`でプロキシを確認 |
| 金庫で緑の点が表示されない | APIキー（トークン）が未設定 | **⚡ Cline設定適用** ボタンを再度クリック |
| Clineフッターのモデルが変わらない | Clineが設定をキャッシュしている | VS Codeをリロード（`Ctrl+Alt+R`） |
| 間違ったモデル名が表示される | 以前のバグ（v0.1.16で修正済み） | プロキシをv0.1.16以上に更新 |

---

#### 🟣 デプロイコマンドコピーボタン — 新しいマシンにインストールする時に使います

新しいコンピュータにwall-vaultプロキシを初めてインストールして金庫に接続する際に使用します。ボタンをクリックするとインストールスクリプト全体がコピーされます。新しいコンピュータのターミナルに貼り付けて実行すると以下が一括で処理されます：

1. wall-vaultバイナリのインストール（既にインストール済みならスキップ）
2. systemdユーザーサービスの自動登録
3. サービス起動および金庫への自動接続

> 💡 スクリプト内にこのエージェントのトークンと金庫サーバーアドレスが既に入力されているため、貼り付け後に修正なしですぐに実行できます。

---

### サービスカード

使用するAIサービスのオン/オフや設定を行うカードです。

- サービスごとの有効化・無効化トグルスイッチ
- ローカルAIサーバー（自分のコンピュータで動作するOllama、LM Studio、vLLM等）のアドレスを入力すると、利用可能なモデルを自動検出します
- **ローカルサービス接続ステータス表示**：サービス名横の●ドットが**緑**なら接続済み、**灰色**なら未接続
- **ローカルサービス自動シグナル**（v0.1.23+）：ローカルサービス（Ollama、LM Studio、vLLM）は接続可否に応じて自動的に有効化/無効化されます。サービスが到達可能になると15秒以内に●が緑に変わりチェックボックスがオンになり、サービスが停止すると自動的にオフになります。クラウドサービス（Google、OpenRouter等）がAPIキーの有無に応じて自動トグルされるのと同じ方式です。

> 💡 **ローカルサービスが別のコンピュータで実行中の場合**：サービスURL入力欄にそのコンピュータのIPを入力してください。例：`http://192.168.1.20:11434`（Ollama）、`http://192.168.1.20:1234`（LM Studio）。サービスが`0.0.0.0`ではなく`127.0.0.1`にのみバインドされている場合、外部IPからのアクセスはできません。サービス設定でバインディングアドレスを確認してください。

### 管理者トークン入力

ダッシュボードでキーの追加・削除などの重要な機能を使おうとすると、管理者トークン入力ポップアップが表示されます。setupウィザードで設定したトークンを入力してください。一度入力すれば、ブラウザを閉じるまで有効です。

> ⚠️ **認証失敗が15分以内に10回を超えると、そのIPが一時的にブロックされます。** トークンを忘れた場合は、`wall-vault.yaml`ファイルの`admin_token`項目を確認してください。

---

## 分散モード（マルチボット）

複数のコンピュータでOpenClawを同時に運用する際、**一つのキー金庫を共有**する構成です。キー管理を一箇所で行えるため便利です。

### 構成例

```
[キー金庫サーバー]
  wall-vault vault    （キー金庫 :56243、ダッシュボード）

[WSL アルファ]          [ラズベリーパイ ガンマ]    [Mac Mini ローカル]
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

金庫サーバーに接続する各ボットの情報を事前に登録します：

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

> 💡 **`192.168.x.x`** の部分は金庫サーバーコンピュータの実際の内部IPアドレスに置き換えてください。ルーター設定または`ip addr`コマンドで確認できます。

---

## 自動起動設定

コンピュータを再起動するたびに手動でwall-vaultを起動するのが面倒であれば、システムサービスとして登録してください。一度登録すれば起動時に自動的に開始されます。

### Linux — systemd（ほとんどのLinux）

systemdはLinuxでプログラムを自動的に起動・管理するシステムです：

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

macOSでプログラムの自動起動を担当するシステムです：

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

## Doctor 自己診断ツール

`doctor`コマンドは、wall-vaultが正しく設定されているか**自己診断し修復するツール**です。

```bash
wall-vault doctor check   # 現在の状態を診断（読み取り専用、何も変更しない）
wall-vault doctor fix     # 問題を自動で修復
wall-vault doctor all     # 診断 + 自動修復を一度に
```

> 💡 何かおかしいと思ったら、まず`wall-vault doctor all`を実行してみてください。多くの問題を自動的に検出・修復します。

---

## RTK トークン節約

*(v0.1.24+)*

**RTK（トークン節約ツール）**は、AIコーディングエージェント（Claude Code等）が実行するシェルコマンドの出力を自動的に圧縮し、トークン使用量を削減します。例えば、`git status`の15行の出力が2行の要約に圧縮されます。

### 基本的な使い方

```bash
# コマンドをwall-vault rtkで囲むと出力が自動フィルタリングされます
wall-vault rtk git status          # 変更されたファイル一覧のみ表示
wall-vault rtk git diff HEAD~1     # 変更行 + 最小コンテキストのみ
wall-vault rtk git log -10         # ハッシュ + メッセージ1行ずつ
wall-vault rtk go test ./...       # 失敗したテストのみ表示
wall-vault rtk ls -la              # サポートされないコマンドは自動切り詰め
```

### サポートコマンドと削減効果

| コマンド | フィルター方式 | 削減率 |
|---------|-------------|--------|
| `git status` | 変更ファイル要約のみ | ~87% |
| `git diff` | 変更行 + 3行コンテキスト | ~60-94% |
| `git log` | ハッシュ + 最初の1行メッセージ | ~90% |
| `git push/pull/fetch` | 進捗表示を除去、要約のみ | ~80% |
| `go test` | 失敗のみ表示、合格はカウント | ~88-99% |
| `go build/vet` | エラーのみ表示 | ~90% |
| その他すべてのコマンド | 先頭50行 + 末尾50行、最大32KB | 可変 |

### 3段階フィルターパイプライン

1. **コマンド別構造フィルター** — git、go等の出力形式を理解し、意味のある部分のみを抽出
2. **正規表現後処理** — ANSIカラーコード除去、空行圧縮、重複行集約
3. **パススルー + 切り詰め** — 未サポートコマンドは先頭/末尾50行のみ保持

### Claude Code連携

Claude Codeの`PreToolUse`フックですべてのシェルコマンドを自動的にRTKを経由するように設定できます。

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

> 💡 **終了コード保持**：RTKは元のコマンドの終了コードをそのまま返します。コマンドが失敗した場合（終了コード ≠ 0）、AIも正確に失敗を検知します。

> 💡 **英語強制出力**：RTKは`LC_ALL=C`でコマンドを実行し、システム言語設定に関係なく常に英語出力を生成します。これによりフィルターが正確に動作します。

---

## 環境変数リファレンス

環境変数とは、プログラムに設定値を渡す方法です。ターミナルで`export 変数名=値`の形式で入力するか、自動起動サービスファイルに記載しておくと常時適用されます。

| 変数 | 説明 | 値の例 |
|------|------|--------|
| `WV_LANG` | ダッシュボード言語 | `ko`, `en`, `ja` |
| `WV_THEME` | ダッシュボードテーマ | `light`, `dark`, `gold` |
| `WV_KEY_GOOGLE` | Google APIキー（カンマ区切りで複数） | `AIza...,AIza...` |
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

ポートが既に他のプログラムに使用されていることが多いです。

```bash
ss -tlnp | grep 56244   # 56244ポートを使用しているプロセスを確認
wall-vault proxy --port 8080   # 別のポート番号で起動
```

### APIキーエラー（429, 402, 401, 403, 582）

| エラーコード | 意味 | 対処方法 |
|------------|------|---------|
| **429** | リクエストが多すぎる（使用量超過） | しばらく待つか、別のキーを追加 |
| **402** | 決済が必要またはクレジット不足 | 該当サービスでクレジットをチャージ |
| **401 / 403** | キーが間違っているか権限がない | キーの値を再確認して再登録 |
| **582** | ゲートウェイ過負荷（クールダウン5分） | 5分後に自動解除 |

```bash
# 登録済みキー一覧とステータスの確認
curl -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys

# キー使用量カウンターのリセット
curl -X POST -H "Authorization: Bearer 管理者トークン" http://localhost:56243/admin/keys/reset
```

### エージェントが「未接続」と表示される場合

「未接続」とは、プロキシプロセスが金庫にシグナル（ハートビート）を送信していない状態です。**設定が保存されていないという意味ではありません。** プロキシが金庫サーバーのアドレスとトークンを知った状態で実行されていれば、接続状態に変わります。

```bash
# 金庫サーバーアドレス、トークン、クライアントIDを指定してプロキシを起動
WV_VAULT_URL=http://金庫サーバーアドレス:56243 \
WV_VAULT_TOKEN=クライアントトークン \
WV_VAULT_CLIENT_ID=クライアントID \
wall-vault proxy
```

接続に成功すると約20秒以内にダッシュボードに🟢実行中と表示されます。

### Ollamaに接続できない場合

Ollamaは自分のコンピュータで直接AIを実行するプログラムです。まずOllamaが起動しているか確認してください。

```bash
curl http://localhost:11434/api/tags   # モデル一覧が表示されれば正常
export OLLAMA_URL=http://192.168.x.x:11434   # 別のコンピュータで実行中の場合
```

> ⚠️ Ollamaが応答しない場合は、まず`ollama serve`コマンドでOllamaを起動してください。

> ⚠️ **大型モデルは応答が遅いです**：`qwen3.5:35b`、`deepseek-r1`などの大きなモデルは応答生成に数分かかることがあります。何も起こっていないように見えても正常に処理中の場合がありますので、お待ちください。

---

## 最近の変更（v0.1.16 ~ v0.1.24）

### v0.1.24 (2026-04-06)
- **RTKトークン節約サブコマンド**：`wall-vault rtk <command>`でシェルコマンドの出力を自動フィルタリングし、AIエージェントのトークン使用量を60-90%削減します。git、go等の主要コマンド用の専用フィルターを内蔵し、未サポートコマンドも自動切り詰めします。Claude Codeの`PreToolUse`フックで透過的に連携されます。

### v0.1.23 (2026-04-06)
- **Ollamaモデル変更修正**：金庫ダッシュボードでOllamaモデルを変更しても実際のプロキシに反映されなかった問題を修正。以前は環境変数（`OLLAMA_MODEL`）のみを使用していましたが、現在は金庫設定を優先使用します。
- **ローカルサービス自動シグナル**：Ollama・LM Studio・vLLMが接続可能であれば自動的に有効化、切断されれば自動的に無効化されます。クラウドサービスのキーベース自動トグルと同じ方式です。

### v0.1.22 (2026-04-05)
- **空のcontentフィールド修正**：thinkingモデル（gemini-3.1-pro、o1、claude thinking等）がmax_tokensをreasoningに使い切り実際の応答を生成できない場合、プロキシが応答JSONの`content`/`text`フィールドを`omitempty`で省略し、OpenAI/Anthropic SDKクライアントが`Cannot read properties of undefined (reading 'trim')`エラーでクラッシュする問題を修正。公式APIスペック通りに常にフィールドを含むように変更。

### v0.1.21 (2026-04-05)
- **Gemma 4モデルサポート**：Google Gemini APIを通じて`gemma-4-31b-it`、`gemma-4-26b-a4b-it`等のGemmaファミリーモデルが使用可能になりました。
- **LM Studio / vLLMサービス正式サポート**：以前はこれらのサービスがプロキシルーティングから漏れており、常にOllamaに代替されていました。OpenAI互換APIで正常にルーティングされるようになりました。
- **ダッシュボードサービス表示修正**：フォールバックが発生しても、ダッシュボードにはユーザーが設定したサービスが常に表示されます。
- **ローカルサービスステータス表示**：ダッシュボード読み込み時にローカルサービス（Ollama、LM Studio、vLLM等）の接続状態を●ドットの色で表示します。
- **ツールフィルター環境変数**：`WV_TOOL_FILTER=passthrough`環境変数でツール（tools）パススルーモードを設定できます。

### v0.1.20 (2026-03-28)
- **包括的セキュリティ強化**：XSS防止（41箇所）、定数時間トークン比較、CORS制限、リクエストサイズ制限、パストラバーサル防止、SSE認証、レートリミッター強化など12のセキュリティ項目を改善。

### v0.1.19 (2026-03-27)
- **Claude Codeオンライン検知**：プロキシを経由しないClaude Codeもダッシュボードでオンラインとして表示されます。

### v0.1.18 (2026-03-26)
- **フォールバックサービス固着修正**：一時的エラーでOllamaにフォールバックした後、元のサービスが復旧すると自動的に復帰します。
- **オフライン検知改善**：15秒間隔のステータス確認でプロキシ停止の検知が速くなりました。

### v0.1.17 (2026-03-25)
- **ドラッグ＆ドロップカード並べ替え**：エージェントカードをドラッグして順序を変更できます。
- **インライン設定適用ボタン**：オフラインのエージェントに[⚡ 設定適用]ボタンが表示されます。
- **cokacdir エージェントタイプ追加**。

### v0.1.16 (2026-03-25)
- **双方向モデル同期**：金庫ダッシュボードでCline・Claude Codeのモデルを変更すると自動的に反映されます。

---

*より詳細なAPI情報は[API.md](API.md)を参照してください。*
